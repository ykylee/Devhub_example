"""workflow_kit.path_resolver helper smoke test (v0.7.34+, ADR-008).

5+ test: HTTPS origin, SSH→HTTPS normalize, GitHub Actions env, path traversal reject,
absolute path reject, URL pass-through.

Test list:
1. test_resolve_https_origin: HTTPS origin → blob URL
2. test_resolve_ssh_origin_normalize: `git@github.com:foo/bar.git` → HTTPS form
3. test_resolve_github_actions_env: `$GITHUB_SERVER_URL` + `$GITHUB_REPOSITORY` 우선
4. test_resolve_path_traversal_reject: `../escape.md` → None
5. test_resolve_absolute_path_reject: `/etc/passwd` → None
6. test_url_passthrough: 이미 URL form 인 path 는 그대로 반환
"""

from __future__ import annotations

import importlib.util
import os
import sys
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
PATH_RESOLVER = SOURCE_ROOT / "workflow_kit" / "path_resolver.py"


def _import_path_resolver():
    """path_resolver module importlib 로 load."""
    import sys
    spec = importlib.util.spec_from_file_location("path_resolver", str(PATH_RESOLVER))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["path_resolver"] = mod
    spec.loader.exec_module(mod)
    return mod


class _EnvPatch:
    """Minimal env var + module attribute patcher. Restores on __exit__."""

    def __init__(self) -> None:
        self._orig_env = dict(os.environ)
        self._patches: list[tuple[object, str, object]] = []

    def __enter__(self) -> "_EnvPatch":
        return self

    def __exit__(self, *exc: object) -> None:
        for k in list(os.environ.keys()):
            if k not in self._orig_env:
                os.environ.pop(k, None)
        for k, v in self._orig_env.items():
            os.environ[k] = v
        for mod, attr, original in self._patches:
            setattr(mod, attr, original)

    def clear_env(self, *keys: str) -> None:
        for k in keys:
            os.environ.pop(k, None)

    def set_env(self, key: str, value: str) -> None:
        os.environ[key] = value

    def patch_attr(self, mod: object, attr: str, value: object) -> None:
        original = getattr(mod, attr)
        setattr(mod, attr, value)
        self._patches.append((mod, attr, original))


# --- Test 1: HTTPS origin ---


def test_resolve_https_origin() -> None:
    """`https://github.com/foo/bar.git` → `https://github.com/foo/bar/blob/main/<path>`."""
    with _EnvPatch() as env:
        env.clear_env("GITHUB_SERVER_URL", "GITHUB_REPOSITORY")
        mod = _import_path_resolver()
        env.patch_attr(mod, "_detect_origin_url", lambda repo_root: "https://github.com/foo/bar")
        env.patch_attr(mod, "_detect_default_branch", lambda repo_root: "main")
        url = mod.resolve_in_repo_path_to_url("workflow/x.md", Path("/fake/repo"))
        assert url == "https://github.com/foo/bar/blob/main/workflow/x.md", f"got {url!r}"


# --- Test 2: SSH origin normalize ---


def test_resolve_ssh_origin_normalize() -> None:
    """`git@github.com:foo/bar.git` → HTTPS form."""
    mod = _import_path_resolver()
    url = mod._normalize_origin_url("git@github.com:foo/bar.git")
    assert url == "https://github.com/foo/bar", f"got {url!r}"
    # HTTPS .git suffix
    url = mod._normalize_origin_url("https://github.com/foo/bar.git")
    assert url == "https://github.com/foo/bar", f"got {url!r}"
    # without .git
    url = mod._normalize_origin_url("https://github.com/foo/bar")
    assert url == "https://github.com/foo/bar", f"got {url!r}"


# --- Test 3: GitHub Actions env ---


def test_resolve_github_actions_env() -> None:
    """`$GITHUB_SERVER_URL` + `$GITHUB_REPOSITORY` env var 가 git config 보다 우선."""
    with _EnvPatch() as env:
        env.set_env("GITHUB_SERVER_URL", "https://github.com")
        env.set_env("GITHUB_REPOSITORY", "acme/cool-proj")
        mod = _import_path_resolver()
        origin = mod._detect_origin_url(Path("/fake/repo"))
        assert origin == "https://github.com/acme/cool-proj", f"got {origin!r}"


# --- Test 4: path traversal reject ---


def test_resolve_path_traversal_reject() -> None:
    """`../escape.md` 같은 path traversal → None."""
    mod = _import_path_resolver()
    assert mod._is_path_safe("../escape.md") is False
    assert mod._is_path_safe("a/../b.md") is False
    assert mod._is_path_safe("..") is False
    url = mod.resolve_in_repo_path_to_url("../escape.md", Path("/fake"))
    assert url is None, f"expected None, got {url!r}"


# --- Test 5: absolute path reject ---


def test_resolve_absolute_path_reject() -> None:
    """`/etc/passwd` 같은 absolute path → None."""
    mod = _import_path_resolver()
    assert mod._is_path_safe("/etc/passwd") is False
    url = mod.resolve_in_repo_path_to_url("/etc/passwd", Path("/fake"))
    assert url is None, f"expected None, got {url!r}"


# --- Test 6: URL pass-through ---


def test_url_passthrough() -> None:
    """이미 `https://...` form 인 path 는 그대로 반환."""
    mod = _import_path_resolver()
    url = mod.resolve_in_repo_path_to_url("https://example.com/spec.md", Path("/fake"))
    assert url == "https://example.com/spec.md", f"got {url!r}"


# --- Test 7-9: V-R12 commit-pinned URL (ADR-018) ---


def test_resolve_commit_pinned() -> None:
    """commit_sha → /blob/<sha>/<path> (immutable)."""
    with _EnvPatch() as env:
        env.clear_env("GITHUB_SERVER_URL", "GITHUB_REPOSITORY")
        mod = _import_path_resolver()
        env.patch_attr(mod, "_detect_origin_url", lambda repo_root: "https://github.com/foo/bar")
        url = mod.resolve_in_repo_path_to_url_pinned(
            "docs/spec.md", Path("/fake"), commit_sha="abc1234"
        )
        assert url == "https://github.com/foo/bar/blob/abc1234/docs/spec.md", f"got {url!r}"


def test_resolve_ref_pinned() -> None:
    """ref (branch/tag) → /blob/<ref>/<path>."""
    with _EnvPatch() as env:
        env.clear_env("GITHUB_SERVER_URL", "GITHUB_REPOSITORY")
        mod = _import_path_resolver()
        env.patch_attr(mod, "_detect_origin_url", lambda repo_root: "https://github.com/foo/bar")
        url = mod.resolve_in_repo_path_to_url_pinned(
            "docs/spec.md", Path("/fake"), ref="v0.7.37"
        )
        assert url == "https://github.com/foo/bar/blob/v0.7.37/docs/spec.md", f"got {url!r}"


def test_resolve_pinned_invalid_sha() -> None:
    """invalid SHA format → None (validate hex + length 7-40)."""
    with _EnvPatch() as env:
        env.clear_env("GITHUB_SERVER_URL", "GITHUB_REPOSITORY")
        mod = _import_path_resolver()
        env.patch_attr(mod, "_detect_origin_url", lambda repo_root: "https://github.com/foo/bar")
        # too short
        assert mod.resolve_in_repo_path_to_url_pinned("docs/spec.md", Path("/fake"), commit_sha="abc") is None
        # non-hex
        assert mod.resolve_in_repo_path_to_url_pinned("docs/spec.md", Path("/fake"), commit_sha="xyz1234") is None
        # no commit_sha, no ref
        assert mod.resolve_in_repo_path_to_url_pinned("docs/spec.md", Path("/fake")) is None


# --- 메인 실행 ---


def main() -> int:
    test_funcs = [
        test_resolve_https_origin,
        test_resolve_ssh_origin_normalize,
        test_resolve_github_actions_env,
        test_resolve_path_traversal_reject,
        test_resolve_absolute_path_reject,
        test_url_passthrough,
        test_resolve_commit_pinned,
        test_resolve_ref_pinned,
        test_resolve_pinned_invalid_sha,
    ]
    failed: list[str] = []
    for fn in test_funcs:
        name = fn.__name__
        try:
            fn()
            print(f"  PASS  {name}")
        except Exception as e:
            print(f"  FAIL  {name}: {type(e).__name__}: {e}")
            failed.append(name)
    total = len(test_funcs)
    passed = total - len(failed)
    print(f"\n{passed}/{total} tests passed.")
    if failed:
        print(f"\n{len(failed)} tests failed:")
        for name in failed:
            print(f"  - {name}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
