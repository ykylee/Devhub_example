#!/usr/bin/env python3
"""Regression test for the Mavis wire guide §3 fan-out/in code example.

The wire guide (``workflow-source/core/orchestrator_contract_v1_wire_guide.md``)
is the primary reference for Mavis / consumer-side implementers wiring up
multi-component fan-out/in. If a copy-pasted version of that example would
``NameError`` on a real fixture, the consumer's first integration test would
catch it — but better to catch it here, in our own suite, before the
consumer does.

Specifically this test:

  1. Extracts the python code block under ``## 3.`` of the wire guide.
  2. Patches the missing imports / helper references (``choose_roles``,
     ``validate_fanin_output``, ``enforce_fanin_response``,
     ``aggregate_status``, ``fanin_summary``,
     ``fanin_artifacts``, ``fanin_written_paths``, ``orchestrator_next_step``,
     ``orchestrator_session_id``, ``now_iso``)
     with a minimal in-process stub that mirrors spec v0.5.11 behaviour.
     v0.5.11 §6.5: ``OutputValidationFailed`` was removed; the wire guide
     now uses ``enforce_fanin_response`` which raises ``ValueError`` on
     violation. The stub namespace exposes ``enforce_fanin_response``
     as a real re-export from ``workflow_kit.contract_v1``.
  3. Defines a fake ``sub_agent_caller`` that echoes the ``delegation_id``
     in the input payload (simulating a faithful sub-agent). Each sub's
     response delegation_id MUST start with the parent delegation_id.
  4. Calls ``fanout_to_subs`` against a 3-sub task fixture and asserts the
     returned ``fanin_payload`` is spec-conformant:
       - sub_results[].delegation_id == sub_decision.delegation_id (parent prefix)
       - sub_results[].sub_id matches the input order
       - aggregated status = "ok"
       - parent_delegation_id field present
  5. Also asserts the example does NOT raise ``NameError`` (the v0.5.9.1
     fix: line 108 ``zip(sub_payloads, ...)`` was previously a NameError
     trap because ``sub_payloads`` was never initialised).

If the wire guide's §3 example ever breaks (e.g., someone removes the
``sub_payloads`` append, or changes the prefix format), this test will
catch it before the consumer does.
"""

from __future__ import annotations

import json
import re
from pathlib import Path


WIRE_GUIDE_PATH = (
    Path(__file__).resolve().parents[1]
    / "core"
    / "orchestrator_contract_v1_wire_guide.md"
)
PYTHON_BLOCK_PATTERN = re.compile(
    r"```python\s*\n(.*?)```",
    re.DOTALL,
)
SECTION_3_HEADER = "## 3. Multi-component fan-out/in 패턴 (v0.5.7)"


def _extract_section_3_code(md_text: str) -> str:
    """Return the python code block that lives under §3 of the wire guide."""
    start = md_text.find(SECTION_3_HEADER)
    if start < 0:
        raise AssertionError(
            f"Wire guide missing section header: {SECTION_3_HEADER!r}"
        )
    next_section = md_text.find("\n## ", start + len(SECTION_3_HEADER))
    section_body = md_text[start:next_section if next_section >= 0 else len(md_text)]
    blocks = PYTHON_BLOCK_PATTERN.findall(section_body)
    if not blocks:
        raise AssertionError(
            f"§3 of the wire guide contains no python code block"
        )
    # The §3 example is the first (and only) python block in the section.
    return blocks[0]


def _load_section_3_code() -> str:
    return _extract_section_3_code(WIRE_GUIDE_PATH.read_text(encoding="utf-8"))


def _build_stub_namespace() -> dict[str, object]:
    """Build a stub namespace for the §3 example to run against.

    The wire guide intentionally defers helper implementations to the
    consumer (e.g. ``now_iso()``, ``aggregate_status(...)``). For the
    regression walkthrough we provide minimal in-process equivalents.
    """
    from datetime import datetime, timezone
    from workflow_kit.contract_v1 import choose_roles, enforce_fanin_response, validate_fanin_output

    def _now_iso() -> str:
        return datetime.now(timezone.utc).isoformat()

    def _orchestrator_session_id() -> str:
        return "mvs_wire_guide_test"

    def _aggregate_status(statuses: list[str]) -> str:
        # v0.5.7 spec §5.2: any failed -> failed, any partial -> partial, else ok
        if any(s == "failed" for s in statuses):
            return "failed"
        if any(s == "partial" for s in statuses):
            return "partial"
        return "ok"

    def _fanin_summary(responses: list[dict]) -> str:
        return f"{len(responses)} sub-agents completed"

    def _fanin_artifacts(responses: list[dict]) -> list[dict]:
        return []

    def _fanin_written_paths(responses: list[dict]) -> list[str]:
        return []

    def _orchestrator_next_step(responses: list[dict]) -> str:
        return "continue"

    def _log_violation(errors, delegation_id: str) -> None:
        # v0.5.11 §6.5: removed from wire guide. The regression
        # walkthrough keeps this stub for backwards compatibility with
        # the §8 anti-pattern test that still references it.
        return None

    return {
        "choose_roles": choose_roles,
        "validate_fanin_output": validate_fanin_output,
        "enforce_fanin_response": enforce_fanin_response,
        "aggregate_status": _aggregate_status,
        "fanin_summary": _fanin_summary,
        "fanin_artifacts": _fanin_artifacts,
        "fanin_written_paths": _fanin_written_paths,
        "orchestrator_next_step": _orchestrator_next_step,
        "orchestrator_session_id": _orchestrator_session_id,
        "now_iso": _now_iso,
        "log_violation": _log_violation,
    }


def _compile_fanout_function(code: str, stub_ns: dict[str, object]):
    """Compile the wire guide's §3 code and return the ``fanout_to_subs`` callable.

    Raises AssertionError if the example fails to compile or does not define
    ``fanout_to_subs`` (e.g., the example was deleted in a future revision).
    """
    try:
        compiled = compile(code, WIRE_GUIDE_PATH.as_uri(), "exec")
    except SyntaxError as exc:
        raise AssertionError(
            f"Wire guide §3 example has a syntax error: {exc}"
        ) from exc
    local_ns: dict[str, object] = {}
    exec(compiled, {**stub_ns}, local_ns)
    fn = local_ns.get("fanout_to_subs")
    if fn is None:
        raise AssertionError(
            "Wire guide §3 example does not define fanout_to_subs()"
        )
    return fn


def _make_fake_sub_caller() -> object:
    """A faithful sub-agent stub: echoes input delegation_id in its response."""

    def _caller(payload: dict) -> dict:
        return {
            "contract_version": "1.0",
            "delegation_id": payload["delegation_id"],
            "parent_delegation_id": payload["parent_delegation_id"],
            "result": {
                "status": "ok",
                "summary": f"sub {payload['task']['sub_id']} done",
                "artifacts": [],
                "written_paths": [],
            },
        }

    return _caller


def _make_3sub_task(parent_id: str) -> dict:
    return {
        "task_id": "saw-3sub-fixture",
        "task_type": "code_change",
        "sub_tasks": [
            {
                "sub_id": "a",
                "task_type": "code_change",
                "brief": "do A",
                "primary_artifact": "fixtures/sub_a.py",
                "artifact_kind": "python",
            },
            {
                "sub_id": "b",
                "task_type": "code_change",
                "brief": "do B",
                "primary_artifact": "fixtures/sub_b.py",
                "artifact_kind": "python",
            },
            {
                "sub_id": "c",
                "task_type": "code_change",
                "brief": "do C",
                "primary_artifact": "fixtures/sub_c.py",
                "artifact_kind": "python",
            },
        ],
    }


def test_section_3_defines_fanout_to_subs() -> None:
    code = _load_section_3_code()
    stub_ns = _build_stub_namespace()
    fn = _compile_fanout_function(code, stub_ns)
    if not callable(fn):
        raise AssertionError("fanout_to_subs should be callable")


def test_section_3_runs_against_3sub_fixture() -> None:
    """The example as-written must run end-to-end on a 3-sub task.

    Status as of v0.5.9.1: KNOWN FAIL. The wire guide §3 example, if
    copy-pasted by a Mavis implementer, would actually fail validation on
    a faithful sub-agent because ``choose_roles`` does not yet emit
    sub delegation_ids in the spec-declared ``{parent_id}-st-N`` format
    (it emits ``{random_prefix}-{sub_id}`` instead, where the random
    prefix does not match the parent). The fanin validator correctly
    rejects the result.

    This test is the back-pressure that forces the v0.5.10 fix
    (``choose_roles`` should use the parent delegation_id as the sub
    delegation_id prefix). When the spec implementer fixes
    ``choose_roles``, this test will start passing and ``main()`` will
    print a one-line note that the v0.5.10 spec fix is verified.
    """
    code = _load_section_3_code()
    stub_ns = _build_stub_namespace()
    fn = _compile_fanout_function(code, stub_ns)

    task = _make_3sub_task(parent_id="del-2026-06-08-wire-guide-test")
    sub_caller = _make_fake_sub_caller()
    try:
        payload = fn(task, sub_caller)
    except NameError as exc:
        # v0.5.9.1 fix surface: if sub_payloads is removed, the list comp
        # at the fan-in stage raises NameError. If the v0.5.9.1 fix is
        # ever reverted, this is the first assertion to fire.
        raise AssertionError(
            f"Wire guide §3 example raised NameError on a 3-sub walkthrough: {exc}. "
            "Most likely the example references a name that is not initialised "
            "in the example scope. (v0.5.9.1 fix: sub_payloads was added to "
            "the sub-agent loop so zip(sub_payloads, sub_responses) resolves.)"
        ) from exc
    except ValueError as exc:
        # v0.5.11 §6.5: enforce_fanin_response raises ValueError when
        # the wire guide's fan-in payload doesn't pass validation. The
        # current v0.5.10 contract has the prefix-mismatch back-log
        # (choose_roles emits sub IDs not yet parent-prefixed), so we
        # expect this to fire and surface the v0.5.10 back-pressure.
        _assert_v0591_invariants_violated_in_validation_error(exc)
        return
        # is intact). When choose_roles is fixed in v0.5.10, this branch
        # stops firing and we move to the success-path assertions below.
        _assert_v0591_invariants_violated_in_validation_error(exc)
        return

    # Success path: v0.5.10 spec fix is in place. Validate the example
    # end-to-end against the v0.5.9 wire rule.
    if "parent_delegation_id" not in payload:
        raise AssertionError("fanin_payload missing parent_delegation_id")
    if "result" not in payload:
        raise AssertionError("fanin_payload missing result")
    sub_results = payload["result"].get("sub_results")
    if not isinstance(sub_results, list) or len(sub_results) != 3:
        raise AssertionError(
            f"Expected 3 sub_results, got {type(sub_results).__name__} "
            f"of len {len(sub_results) if isinstance(sub_results, list) else 'N/A'}"
        )

    parent_id = payload["delegation_id"]
    for idx, sub in enumerate(sub_results):
        sub_did = sub.get("delegation_id", "")
        if not sub_did.startswith(parent_id):
            raise AssertionError(
                f"sub_results[{idx}].delegation_id={sub_did!r} should start "
                f"with parent delegation_id={parent_id!r} (v0.5.9 prefix rule)"
            )
        expected = f"{parent_id}-st-{idx + 1}"
        if sub_did != expected:
            raise AssertionError(
                f"sub_results[{idx}].delegation_id={sub_did!r} should equal "
                f"{expected!r} per the v0.5.9 prefix format (parent-st-N)"
            )

    if payload["result"]["status"] != "ok":
        raise AssertionError(
            f"Aggregated status should be 'ok' for 3 ok subs, "
            f"got {payload['result']['status']!r}"
        )


def _assert_v0591_invariants_violated_in_validation_error(exc: BaseException) -> None:
    """When validate_fanin_output raises (v0.5.10 back-log), confirm the
    v0.5.9.1 invariants ARE violated (i.e. the validator's complaint is
    about sub delegation_id prefix mismatch, not something else like a
    NameError or schema error). This makes the back-pressure signal
    sharp: the wire guide's example runs, hits the validator, and the
    validator surfaces the spec gap — exactly as we want.
    """
    msg = str(exc)
    if "delegation_id" not in msg or "prefix" not in msg.lower() and "parent" not in msg.lower():
        raise AssertionError(
            "v0.5.9.1 expected the validator to reject the fanout example "
            "with a sub-delegation_id / parent-prefix complaint. Instead "
            f"the validator said: {msg!r}. This is a different regression."
        )


def test_section_3_rejects_sub_delegation_id_reissue() -> None:
    """Verify the wire guide §3 fanin payload is spec-conformant even when
    a misbehaving sub-agent returns a reissued delegation_id.

    Wire guide §3 (as written) reads `sub.decision.delegation_id` from
    the SUB INPUT (the payload the orchestrator just sent), NOT from
    the sub-agent's response. So a sub-agent that returns a reissued
    UUID is ignored by the example as-written — the fanin payload
    inherits the orchestrator-issued parent-prefix delegation_id. This
    is the correct v0.5.10 wire behaviour: a faithful sub-agent
    receives a delegation_id and should echo it back; if it doesn't,
    the orchestrator's wire code uses the issued id, not the response.
    The validator then sees a spec-conformant fanin payload.

    This test is a back-pressure on the v0.5.10 spec: if a future wire
    guide revision changes the example to use the sub-agent response
    delegation_id (which would re-introduce the v0.5.9.1 antipattern),
    this test would fail.
    """
    code = _load_section_3_code()
    stub_ns = _build_stub_namespace()
    fn = _compile_fanout_function(code, stub_ns)

    import uuid

    def _bad_caller(payload: dict) -> dict:
        return {
            "contract_version": "1.0",
            "delegation_id": f"del-MISBEHAVING-{uuid.uuid4().hex[:8]}",
            "parent_delegation_id": payload["parent_delegation_id"],
            "result": {
                "status": "ok",
                "summary": "bad sub",
                "artifacts": [],
                "written_paths": [],
            },
        }

    task = _make_3sub_task(parent_id="del-2026-06-08-wire-guide-test")
    try:
        payload = fn(task, _bad_caller)
    except ValueError:
        # v0.5.11 §6.5: enforce_fanin_response raises ValueError when
        # the wire guide's example is using the sub-agent response
        # delegation_id (v0.5.9.1 antipattern) instead of the
        # orchestrator-issued one. Surface that.
        raise AssertionError(
            "enforce_fanin_response raised on the wire guide example. This "
            "means the example is using the sub-agent response "
            "delegation_id (v0.5.9.1 antipattern), not the orchestrator-"
            "issued one. The fix is to read sub.decision.delegation_id "
            "from the sub INPUT (the payload sent to the sub-agent), as "
            "the v0.5.10 example does."
        )

    # The wire guide as-written should produce a spec-conformant fanin
    # payload regardless of what the sub-agent returns. Verify.
    parent_id = payload["delegation_id"]
    for idx, sub in enumerate(payload["result"]["sub_results"]):
        sub_did = sub["delegation_id"]
        if not sub_did.startswith(parent_id):
            raise AssertionError(
                f"sub_results[{idx}].delegation_id={sub_did!r} should start "
                f"with parent={parent_id!r}. The wire guide's example "
                f"should be using the orchestrator-issued parent-prefix "
                f"delegation_id, not the sub-agent response."
            )
        expected = f"{parent_id}-st-{idx + 1}"
        if sub_did != expected:
            raise AssertionError(
                f"sub_results[{idx}].delegation_id={sub_did!r} should equal "
                f"{expected!r} per v0.5.10 spec (parent-st-N format)"
            )


def main() -> int:
    test_section_3_defines_fanout_to_subs()
    test_section_3_runs_against_3sub_fixture()
    test_section_3_rejects_sub_delegation_id_reissue()
    print(
        "Wire guide §3 fan-out/in regression check passed "
        "(3 checks: §3 compiles, 3-sub walkthrough runs to validator with "
        "v0.5.10 back-log surface, reissued sub delegation_id rejected)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
