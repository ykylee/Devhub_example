"""atomic_write helper — DevHub 자체 구현 (v0.7.15 follow-up, D).

v0.7.15 atomic_write helper = POSIX os.replace 기반 atomic write.
본 follow-up 의 핵심 = partial write 시 *원본 보존* + power-loss safety.

설계 정공법 (2026-06-15):
1. **temp file + os.replace**: 원본 path 와 같은 directory 에 temp file 작성 →
   os.replace 로 atomic rename. 실패 시 temp file cleanup (best-effort).
2. **fsync (best-effort)**: write 후 os.fsync(fd) 호출 → OS-level durability.
   parent dir 도 fsync (POSIX 한정, 구현 단순화 — best-effort).
3. **POSIX 전용**: os.replace 가 atomic 보장 (POSIX rename(2) 기반). Windows 는
   follow-up (v0.7.18+).
4. **no vendor 의존**: workflow_kit.common.atomic_write (v0.7.15+) 와 *pattern* 은
   동일하지만 본 follow-up 은 자체 구현. vendor 와 import 경로 ❌.

주 사용처:
- ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md, log.md}
  write path 의 partial-write 안전성 보장
- wiki frontmatter / index / page 의 in-place update 시 원본 보존

Reference:
- vendor/standard_ai_workflow/v0.7.15+ workflow_kit.common.atomic_write
  (pattern 동일, 본 구현은 독립)
- POSIX rename(2) man page — atomic 보장 spec
- POSIX fsync(2) man page — power-loss safety
"""

from __future__ import annotations

import errno
import json
import os
import sys
import tempfile
import uuid
from pathlib import Path
from typing import Any

# POSIX-only guard. Windows 는 follow-up.
if sys.platform == "win32":
    raise OSError(
        "atomic_write 는 POSIX 전용 (macOS/Linux). "
        "Windows 는 v0.7.18+ follow-up. 현재 platform: win32"
    )


def _fsync_directory(dir_path: Path) -> None:
    """Parent directory fsync (best-effort, POSIX only).

    이유: rename(2) 만 fsync 하면 file content 는 durable 하지만
    *directory entry* (file name → inode mapping) 는 power-loss 시
    사라질 수 있음. directory fsync 가 이 mapping 을 durable 화.
    """
    try:
        fd = os.open(str(dir_path), os.O_RDONLY)
        try:
            os.fsync(fd)
        finally:
            os.close(fd)
    except OSError:
        # best-effort. 일부 FS (e.g. tmpfs) 또는 권한 부족 시 silent.
        pass


def atomic_write_text(path: Path, text: str, encoding: str = "utf-8") -> None:
    """Atomic text write (POSIX only).

    Sequence:
    1. parent dir mkdir (parents=True, exist_ok=True)
    2. temp file 생성 (path + ".tmp.<uuid>.atomic")
    3. temp file 에 text write (UTF-8 default)
    4. temp file fsync
    5. os.replace (atomic rename, POSIX guarantee)
    6. parent dir fsync (best-effort)

    Args:
        path: 최종 file path. parent dir 자동 생성.
        text: file content. binary 아닌 text.
        encoding: file encoding. default utf-8.

    Raises:
        OSError: write 실패 (temp file cleanup 시도 후).
    """
    path = Path(path)
    parent = path.parent
    parent.mkdir(parents=True, exist_ok=True)

    # unique temp file name (같은 dir 내, 같은 FS — atomic rename 전제)
    tmp_name = f"{path.name}.tmp.{uuid.uuid4().hex}.atomic"
    tmp_path = parent / tmp_name

    try:
        # write to temp (write + flush)
        with open(tmp_path, "w", encoding=encoding) as f:
            f.write(text)
            f.flush()
            os.fsync(f.fileno())

        # atomic rename (POSIX)
        os.replace(tmp_path, path)

        # parent dir fsync (best-effort)
        _fsync_directory(parent)
    except BaseException:
        # cleanup temp on failure (best-effort)
        try:
            if tmp_path.exists():
                tmp_path.unlink()
        except OSError:
            pass
        raise

def atomic_append_text(path: Path, text: str, encoding: str = "utf-8") -> None:
    """Atomic text append (POSIX only).

    O_APPEND 로 file 을 열어 단일 write(2) call 로 끝까지 append.
    read-modify-write race 가 본질적으로 없음 (kernel 이 write 위치 결정).
    fsync 후 닫아 power-loss safety.

    기존 `atomic_write_text(existing + line)` 패턴은 race 가 있어 (두 process 가
    같은 existing 을 read → 한 쪽 lost update). log.md 같은 append-only 파일은
    본 helper 로 교체 필수.

    Sequence:
    1. parent dir mkdir (parents=True, exist_ok=True)
    2. O_WRONLY | O_CREAT | O_APPEND 로 open
    3. text write (UTF-8 default)
    4. fsync (durable)
    5. close (fd 가 high-level wrapper 에서 관리될 경우 명시)

    Args:
        path: 대상 file path. 없으면 생성.
        text: append 할 content.
        encoding: file encoding. default utf-8.

    Raises:
        OSError: write 실패 시. partial write 가능 (write 가 atomic 이지만
        fsync 실패 시 durability 미보장).
    """
    path = Path(path)
    parent = path.parent
    parent.mkdir(parents=True, exist_ok=True)

    fd = os.open(str(path), os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0o644)
    try:
        with os.fdopen(fd, "a", encoding=encoding, closefd=True) as f:
            f.write(text)
            f.flush()
            os.fsync(f.fileno())
    except BaseException:
        # append 실패 시: 이미 write 된 byte 는 남고 fsync 만 실패할 수 있음.
        # file 자체를 지우지 않음 (이미 다른 line 이 들어있을 수 있음).
        raise


def atomic_write_json(
    path: Path,
    data: Any,
    indent: int = 2,
    ensure_ascii: bool = False,
) -> None:
    """Atomic JSON write (POSIX only).

    json.dump + atomic_write_text wrapper. ensure_ascii=False 가 default
    (한글 등 non-ASCII 보존). indent=2 가 default (human-readable).

    Args:
        path: 최종 file path.
        data: JSON-serializable object.
        indent: JSON indent. default 2.
        ensure_ascii: ASCII escape 여부. default False (Unicode 보존).
    """
    path = Path(path)
    text = json.dumps(data, indent=indent, ensure_ascii=ensure_ascii, sort_keys=False)
    # trailing newline (POSIX text file convention)
    if not text.endswith("\n"):
        text += "\n"
    atomic_write_text(path, text)


# ----- CLI entry (선택) -----

def _main() -> int:
    """CLI: atomic-write <path> [--json] [--text] [-- encoding utf-8].

    stdin 에서 content read → file 에 atomic write. pipe-friendly.
    """
    import argparse

    p = argparse.ArgumentParser(
        prog="atomic-write",
        description="Atomic file write (POSIX only, v0.7.15 DevHub follow-up)",
    )
    p.add_argument("path", help="target file path")
    p.add_argument("--text", action="store_true", help="treat stdin as text (default)")
    p.add_argument("--json", action="store_true", help="validate stdin as JSON before write")
    p.add_argument("--encoding", default="utf-8", help="text encoding (default utf-8)")
    args = p.parse_args()

    content = sys.stdin.read()

    if args.json:
        try:
            data = json.loads(content)
        except json.JSONDecodeError as e:
            print(f"[atomic-write] error: invalid JSON: {e}", file=sys.stderr)
            return 1
        atomic_write_json(Path(args.path), data)
    else:
        atomic_write_text(Path(args.path), content, encoding=args.encoding)

    print(f"[atomic-write] OK: {args.path} ({len(content)} bytes)", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(_main())
