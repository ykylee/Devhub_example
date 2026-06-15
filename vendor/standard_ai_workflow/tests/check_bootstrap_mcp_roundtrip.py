#!/usr/bin/env python3
"""Smoke test: end-to-end MCP round-trip via bootstrap-emitted config.

This complements the existing read-only JSON-RPC bridge smoke tests by
verifying that the **mcp.json file emitted by ``bootstrap --enable-mcp``**
can actually be spawned by a harness, run a JSON-RPC session, and serve
``tools/list`` + ``tools/call`` without errors. It catches the case where
the bootstrap generates a syntactically valid config but the underlying
bridge entry point is broken (wrong PYTHONPATH, missing module, etc.).

By default this test runs the smoke for **every supported harness** so a
single run covers Codex / OpenCode / Gemini CLI / Antigravity / MiniMax Code.
A specific subset can be requested via the ``--harness`` flag (repeatable).

Usage::

    PYTHONPATH=workflow-source python3 workflow-source/tests/check_bootstrap_mcp_roundtrip.py
    PYTHONPATH=workflow-source python3 workflow-source/tests/check_bootstrap_mcp_roundtrip.py --harness minimax-code

The test is CWD-independent: it anchors PYTHONPATH at this repository's
``workflow-source`` directory and the harness cwd at the bootstrap output
directory (a temp dir per harness run).
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import tempfile
import time
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE_ROOT = REPO_ROOT / "workflow-source"

#: Harness name → key in the bootstrap manifest's ``generated_harness_files``
#: dict that points at the emitted MCP config file.
HARNESS_CONFIG_KEY = {
    "codex": "codex_mcp_config",
    "opencode": "opencode_mcp_config",
    "gemini-cli": "gemini_cli_mcp_config",
    "antigravity": "antigravity_mcp_config",
    "minimax-code": "minimax_code_mcp_config",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--harness",
        action="append",
        choices=list(HARNESS_CONFIG_KEY),
        help="Limit the smoke to a single harness. Repeatable; defaults to all.",
    )
    return parser.parse_args()


def run_bootstrap(target_root: Path, harness: str) -> dict[str, object]:
    """Run ``bootstrap --enable-mcp`` for a single harness and return the manifest payload.

    The bootstrap script writes a progress preamble to stdout before
    printing the final manifest. We therefore locate the manifest by
    parsing successive JSON candidates from the end of stdout rather
    than relying on a fixed delimiter.
    """
    args = [
        sys.executable,
        str(SOURCE_ROOT / "scripts" / "bootstrap_workflow_kit.py"),
        "--target-root",
        str(target_root),
        "--project-slug",
        f"mcp_smoke_{harness.replace('-', '_')}",
        "--project-name",
        f"MCP Smoke {harness}",
        "--harness",
        harness,
        "--adoption-mode",
        "new",
        "--copy-core-docs",
        "--enable-mcp",
    ]
    env = os.environ.copy()
    env["PYTHONPATH"] = f"{SOURCE_ROOT}{os.pathsep}{SOURCE_ROOT.parent / 'scripts'}"
    completed = subprocess.run(
        args,
        cwd=str(REPO_ROOT),
        capture_output=True,
        text=True,
        check=True,
        env=env,
    )
    # Walk backwards over the lines, trying to parse a JSON object out of
    # the tail. The manifest is the last valid JSON object on stdout.
    lines = completed.stdout.splitlines()
    for end in range(len(lines), 0, -1):
        found_start = None
        for start in range(end - 1, -1, -1):
            if lines[start].lstrip().startswith("{"):
                candidate = "\n".join(lines[start:end])
                try:
                    payload = json.loads(candidate)
                except json.JSONDecodeError:
                    continue
                if isinstance(payload, dict) and "generated_harness_files" in payload:
                    return payload
                found_start = start
                break
        if found_start is not None:
            end = found_start
    raise AssertionError(
        f"bootstrap did not produce a recognisable manifest on stdout. "
        f"Last 200 chars:\n{completed.stdout[-200:]}"
    )


def spawn_bridge(
    manifest: dict[str, object],
    target_root: Path,
    harness: str,
) -> subprocess.Popen[str]:
    """Spawn the MCP bridge described by the bootstrap manifest for ``harness``.

    Anchors ``PYTHONPATH`` at the kit's ``workflow-source`` so the entry
    point can resolve ``workflow_kit`` regardless of the caller's cwd,
    and rewrites the emitted ``python3`` command to the test runner's
    interpreter (``sys.executable``). The bootstrap's relative ``python3``
    resolves against the harness's ``$PATH``; in the smoke test we
    cannot assume the harness PATH contains a Python that has the kit's
    dependencies, so we pin to the same interpreter we used to run the
    test (which definitely has them).
    """
    config_key = HARNESS_CONFIG_KEY[harness]
    if config_key not in manifest["generated_harness_files"]:
        raise AssertionError(
            f"bootstrap did not emit {config_key!r} for harness {harness!r}. "
            f"Keys present: {sorted(manifest['generated_harness_files'])}"
        )
    config_path = Path(str(manifest["generated_harness_files"][config_key]))
    config_text = config_path.read_text(encoding="utf-8")
    if config_path.suffix == ".toml":
        # Codex uses TOML; we only care about the [mcp_servers.<alias>] table.
        server_block = _parse_codex_toml_server_block(config_text)
    else:
        config = json.loads(config_text)
        # Resolve to the actual server block under our canonical alias.
        # Each harness uses a slightly different top-level key, and the
        # ``mcp`` / ``mcpServers`` / ``mcp_servers`` keys wrap a dict
        # keyed by the server alias (``standardAiWorkflowReadOnly``).
        if "mcp" in config and isinstance(config["mcp"], dict):
            server_block = (
                config["mcp"].get("standardAiWorkflowReadOnly") or config["mcp"]
            )
        elif "mcpServers" in config and isinstance(config["mcpServers"], dict):
            server_block = (
                config["mcpServers"].get("standardAiWorkflowReadOnly") or config["mcpServers"]
            )
        elif "mcp_servers" in config and isinstance(config["mcp_servers"], dict):
            server_block = (
                config["mcp_servers"].get("standardAiWorkflowReadOnly") or config["mcp_servers"]
            )
        else:
            server_block = {}
    if not server_block or "command" not in server_block:
        raise AssertionError(
            f"Emitted MCP config for {harness!r} has no recognisable server block. "
            f"Top-level keys: {sorted(config)}"
        )
    cmd = [sys.executable, *server_block["args"]]
    env = os.environ.copy()
    kit_env = {**server_block.get("env", {}), "PYTHONPATH": str(SOURCE_ROOT)}
    env.update({k: str(v) for k, v in kit_env.items()})
    env["STANDARD_AI_WORKFLOW_ROOT"] = str(target_root.resolve())
    return subprocess.Popen(
        cmd,
        cwd=str(target_root),
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=0,
        env=env,
    )


def round_trip(
    proc: subprocess.Popen[str], request: dict[str, object], *, timeout: float = 5.0
) -> dict[str, object]:
    """Send a single JSON-RPC request over stdio and return the response."""
    assert proc.stdin is not None and proc.stdout is not None
    proc.stdin.write(json.dumps(request, ensure_ascii=False) + "\n")
    proc.stdin.flush()
    deadline = time.time() + timeout
    line = ""
    while time.time() < deadline:
        line = proc.stdout.readline()
        if line:
            return json.loads(line)
        time.sleep(0.05)
    raise AssertionError(f"No response within {timeout}s for request {request['method']}")


def _parse_codex_toml_server_block(toml_text: str) -> dict[str, object]:
    """Best-effort parser for the Codex ``[mcp_servers.<alias>]`` TOML block.

    Codex's snippet uses a flat layout with simple key = value pairs plus a
    single string array for ``args``. We don't need full TOML semantics —
    just enough to extract ``command``, ``args`` (parsed from a TOML list
    literal), and the env vars.
    """
    import re

    # Locate the [mcp_servers.standardAiWorkflowReadOnly] section.
    section_match = re.search(
        r"^\[mcp_servers\.standardAiWorkflowReadOnly\]\s*$(.*?)(?=^\[|\Z)",
        toml_text,
        re.MULTILINE | re.DOTALL,
    )
    if not section_match:
        raise AssertionError("Codex TOML is missing the [mcp_servers.standardAiWorkflowReadOnly] section.")
    body = section_match.group(1)
    server: dict[str, object] = {}

    # command = "python3"
    command_match = re.search(r'^\s*command\s*=\s*"([^"]+)"', body, re.MULTILINE)
    if not command_match:
        raise AssertionError("Codex TOML is missing `command = \"...\"`.")
    server["command"] = command_match.group(1)

    # args = ["-m", "workflow_kit.server.read_only_jsonrpc", "--stdio-lines"]
    args_match = re.search(r"^\s*args\s*=\s*\[(.*?)\]\s*$", body, re.MULTILINE | re.DOTALL)
    if args_match:
        items = re.findall(r'"([^"]+)"', args_match.group(1))
        server["args"] = items
    else:
        server["args"] = []

    # env entries like: PYTHONPATH = "workflow-source"
    env: dict[str, str] = {}
    for env_match in re.finditer(r'^\s*([A-Z_][A-Z0-9_]*)\s*=\s*"([^"]*)"', body, re.MULTILINE):
        key = env_match.group(1)
        if key in {"command"} or key in {"startup_timeout_sec", "tool_timeout_sec"}:
            continue
        env[key] = env_match.group(2)
    if env:
        server["env"] = env
    return server


def smoke_one_harness(harness: str) -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        target_root = Path(tmpdir) / f"mcp-smoke-{harness}"
        target_root.mkdir(parents=True, exist_ok=True)
        manifest = run_bootstrap(target_root, harness)
        proc = spawn_bridge(manifest, target_root, harness)
        try:
            init = round_trip(
                proc,
                {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            )
            if init["result"]["serverInfo"]["name"] != "workflow_read_only_bundle":
                raise AssertionError("initialize did not return the expected server name.")

            tools = round_trip(
                proc,
                {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}},
            )["result"]["tools"]
            if not tools:
                raise AssertionError("tools/list returned no tools.")
            names = {tool["name"] for tool in tools}
            for required in ("latest_backlog", "check_doc_metadata", "check_doc_links"):
                if required not in names:
                    raise AssertionError(f"required tool missing: {required}")

            call_result = round_trip(
                proc,
                {
                    "jsonrpc": "2.0",
                    "id": 3,
                    "method": "tools/call",
                    "params": {
                        "name": "latest_backlog",
                        "arguments": {
                            "work_backlog_index_path": str(
                                REPO_ROOT
                                / "workflow-source"
                                / "examples"
                                / "acme_delivery_platform"
                                / "work_backlog.md"
                            )
                        },
                    },
                },
            )["result"]
            if call_result.get("_meta", {}).get("bridge_phase") != "jsonrpc_draft_fixture":
                raise AssertionError("latest_backlog call did not return the expected bridge phase.")
            structured = call_result.get("structuredContent", {})
            if structured.get("status") != "ok":
                raise AssertionError(
                    f"latest_backlog structuredContent status is {structured.get('status')!r}, expected 'ok'."
                )
        finally:
            try:
                if proc.stdin is not None:
                    proc.stdin.close()
            except Exception:
                pass
            try:
                proc.wait(timeout=2)
            except Exception:
                proc.kill()
            stderr = proc.stderr.read() if proc.stderr is not None else ""
            if stderr.strip():
                snippet = stderr.strip().splitlines()[:5]
                print(f"  [{harness}] bridge stderr (first 5 lines):\n    " + "\n    ".join(snippet))


def main() -> int:
    args = parse_args()
    harnesses = args.harness or list(HARNESS_CONFIG_KEY)
    print(f"Running MCP round-trip smoke for: {harnesses}")
    for harness in harnesses:
        print(f"  - {harness} ...", end=" ", flush=True)
        smoke_one_harness(harness)
        print("ok")
    print("Bootstrap-emitted MCP config round-trip smoke check passed for all selected harnesses.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
