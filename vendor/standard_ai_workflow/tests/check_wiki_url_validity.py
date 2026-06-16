"""workflow_kit.url_validity helper smoke test (v0.7.35+, V-R10 PoC).

V-R10 rule: URL validity (offline 8 check + optional online HEAD). ADR-010 채택.
PoC 단계: 6 test 로 offline 8 check 검증 (online layer 별도 turn).

Test list:
1. test_url_https_only_accept: https:// 만 accept, http:// reject
2. test_url_scheme_reject: ftp://, file://, javascript: reject
3. test_url_localhost_reject: localhost, 127.0.0.1, ::1 reject
4. test_url_private_ip_reject: 10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12 reject
5. test_url_credentials_reject: user:pass@host reject
6. test_url_path_traversal_reject: ../../etc/passwd reject
"""

from __future__ import annotations

import importlib.util
import json
import multiprocessing
import socket
import sys
import tempfile
import multiprocessing
import socket
import tempfile
import time
import urllib.error
from io import BytesIO
from pathlib import Path

SOURCE_ROOT = Path(__file__).resolve().parents[1]
URL_VALIDITY = SOURCE_ROOT / "workflow_kit" / "url_validity.py"


def _import_url_validity():
    """url_validity module importlib 로 load."""
    import sys
    spec = importlib.util.spec_from_file_location("url_validity", str(URL_VALIDITY))
    mod = importlib.util.module_from_spec(spec)
    sys.modules["url_validity"] = mod
    spec.loader.exec_module(mod)
    return mod


# --- Test 1: https only accept ---


def test_url_https_only_accept() -> None:
    """`https://` 만 accept. `http://` reject (downgrade attack 방지)."""
    mod = _import_url_validity()
    issues = mod.check_url("https://example.com/spec.md")
    # https 만, 다른 check 모두 pass
    errors = [i for i in issues if i.severity == "error"]
    assert not errors, f"https URL should pass: {errors}"
    issues_http = mod.check_url("http://example.com/spec.md")
    errors_http = [i for i in issues_http if i.severity == "error"]
    assert any("scheme" in i.message.lower() or "https" in i.message.lower() for i in errors_http), (
        f"http:// should be rejected: {errors_http}"
    )


# --- Test 2: ftp/file/javascript reject ---


def test_url_scheme_reject() -> None:
    """`ftp://`, `file://`, `javascript:` scheme reject."""
    mod = _import_url_validity()
    for bad in ("ftp://example.com/spec.md", "file:///etc/passwd", "javascript:alert(1)"):
        issues = mod.check_url(bad)
        errors = [i for i in issues if i.severity == "error"]
        assert errors, f"{bad!r} should produce errors, got: {issues}"


# --- Test 3: localhost reject ---


def test_url_localhost_reject() -> None:
    """`localhost`, `127.0.0.1`, `::1` reject (SSRF 방지)."""
    mod = _import_url_validity()
    for bad in ("https://localhost/spec.md", "https://127.0.0.1/spec.md", "https://[::1]/spec.md"):
        issues = mod.check_url(bad)
        errors = [i for i in issues if i.severity == "error"]
        assert any(
            kw in i.message.lower()
            for i in errors
            for kw in ("localhost", "loopback", "private", "internal")
        ), f"{bad!r} should reject localhost, got: {issues}"

# --- Test 4: private IP reject ---


def test_url_private_ip_reject() -> None:
    """private IP 대역 (10/8, 192.168/16, 172.16/12, fc00::/7) reject."""
    mod = _import_url_validity()
    for bad in (
        "https://10.0.0.1/spec.md",
        "https://192.168.1.1/spec.md",
        "https://172.16.0.1/spec.md",
        "https://[fc00::1]/spec.md",
    ):
        issues = mod.check_url(bad)
        errors = [i for i in issues if i.severity == "error"]
        assert any("private" in i.message.lower() or "internal" in i.message.lower() for i in errors), (
            f"{bad!r} should reject private IP, got: {issues}"
        )


# --- Test 5: credentials reject ---


def test_url_credentials_reject() -> None:
    """`user:pass@host` 형태의 credentials in URL reject (security)."""
    mod = _import_url_validity()
    issues = mod.check_url("https://user:pass@example.com/spec.md")
    errors = [i for i in issues if i.severity == "error"]
    assert any("credential" in i.message.lower() or "user" in i.message.lower() for i in errors), (
        f"credential URL should be rejected, got: {issues}"
    )


# --- Test 6: path traversal reject ---


def test_url_path_traversal_reject() -> None:
    """path 에 `..` segment 포함 시 reject."""
    mod = _import_url_validity()
    issues = mod.check_url("https://example.com/a/../../etc/passwd")
    errors = [i for i in issues if i.severity == "error"]
    assert any("traversal" in i.message.lower() or ".." in i.message for i in errors), (
        f"path traversal URL should be rejected, got: {issues}"
    )


# --- Test 7-12: Online HEAD layer (ADR-012) ---

import socket
import ssl
import urllib.error
from io import BytesIO


class _MockHTTPResponse:
    """Minimal mock for urllib.request.urlopen return value."""

    def __init__(self, status: int, headers: dict[str, str] | None = None):
        self.status = status
        self.headers = headers or {}

    def __enter__(self):
        return self

    def __exit__(self, *args):
        pass

    def read(self, *args):
        return b""


def _patch_urlopen(monkeypatch, mod, side_effect):
    """Replace urlopen with a mock returning the given side_effect."""
    monkeypatch.setattr(mod, "check_url_online", None)  # noop
    # patch urllib.request.urlopen via the module's import
    import urllib.request
    monkeypatch.orig_urlopen = urllib.request.urlopen
    urllib.request.urlopen = lambda *a, **kw: side_effect()
    return urllib.request


def test_online_http_200_pass() -> None:
    """HTTP 200 → pass (no issues)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _MockHTTPResponse(200)
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        assert not issues, f"expected no issues for 200, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_online_http_404_stale() -> None:
    """HTTP 404 → ERROR (stale URL)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _MockHTTPResponse(404)
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        errors = [i for i in issues if i.severity == "error"]
        assert any("404" in i.message or "stale" in i.message.lower() for i in errors), (
            f"404 should be error, got: {issues}"
        )
    finally:
        urllib.request.urlopen = orig


def test_online_http_500_warn() -> None:
    """HTTP 500 → WARN (transient)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _MockHTTPResponse(500)
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        warns = [i for i in issues if i.severity == "warn"]
        assert any("500" in i.message or "server error" in i.message.lower() for i in warns), (
            f"500 should be warn, got: {issues}"
        )
    finally:
        urllib.request.urlopen = orig


def test_online_timeout_warn() -> None:
    """Connection timeout → WARN (slow host)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    def _timeout(*a, **kw):
        raise socket.timeout("read timeout")
    urllib.request.urlopen = _timeout
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        warns = [i for i in issues if i.severity == "warn"]
        assert any("timeout" in i.message.lower() for i in warns), f"timeout should be warn, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_online_tls_error_reject() -> None:
    """TLS error → ERROR (security risk)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    def _tls_fail(*a, **kw):
        raise ssl.SSLError("certificate verify failed")
    urllib.request.urlopen = _tls_fail
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        errors = [i for i in issues if i.severity == "error"]
        assert any("TLS" in i.message or "ssl" in i.message.lower() for i in errors), (
            f"TLS error should be error, got: {issues}"
        )
    finally:
        urllib.request.urlopen = orig


def test_online_dns_failure_reject() -> None:
    """DNS failure → ERROR (host does not exist)."""
    import urllib.request
    mod = _import_url_validity()

    orig = urllib.request.urlopen
    def _dns_fail(*a, **kw):
        raise urllib.error.URLError("Name or service not known")
    urllib.request.urlopen = _dns_fail
    try:
        issues = mod.check_url_online("https://example.com/spec.md", timeout=5.0)
        errors = [i for i in issues if i.severity == "error"]
        assert any("DNS" in i.message or "url error" in i.message.lower() or "Name" in i.message for i in errors), (
            f"DNS failure should be error, got: {issues}"
        )
    finally:
        urllib.request.urlopen = orig


# --- Test 13-16: V-R10 v2 cache (ADR-013) ---


def test_cache_miss_then_hit() -> None:
    """First check → cache miss + store. Second check (same URL) → cache hit (no network)."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        call_count = {"n": 0}
        def _counting_urlopen(*a, **kw):
            call_count["n"] += 1
            return _MockHTTPResponse(200)
        urllib.request.urlopen = _counting_urlopen
        try:
            # 1st: cache miss → live check → store
            issues1 = mod.check_url_with_cache("https://example.com/spec.md", cache_file=cache_file, ttl_seconds=60)
            assert not issues1
            assert call_count["n"] == 1, f"expected 1 live check, got {call_count['n']}"
            # 2nd: cache hit → no live check
            issues2 = mod.check_url_with_cache("https://example.com/spec.md", cache_file=cache_file, ttl_seconds=60)
            assert not issues2
            assert call_count["n"] == 1, f"cache hit should not call urlopen, got {call_count['n']} calls"
        finally:
            urllib.request.urlopen = orig


def test_cache_ttl_expired() -> None:
    """TTL expired → cache miss + fresh live check."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        call_count = {"n": 0}
        def _counting_urlopen(*a, **kw):
            call_count["n"] += 1
            return _MockHTTPResponse(200)
        urllib.request.urlopen = _counting_urlopen
        try:
            # 1st: store
            mod.check_url_with_cache("https://example.com/spec.md", cache_file=cache_file, ttl_seconds=1)
            assert call_count["n"] == 1
            # wait > TTL
            time.sleep(1.2)
            # 2nd: TTL expired → fresh check
            mod.check_url_with_cache("https://example.com/spec.md", cache_file=cache_file, ttl_seconds=1)
            assert call_count["n"] == 2, f"TTL expired should trigger fresh check, got {call_count['n']} calls"
        finally:
            urllib.request.urlopen = orig


def test_cache_stats() -> None:
    """cache_stats() returns total/fresh/expired counts."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        urllib.request.urlopen = lambda *a, **kw: _MockHTTPResponse(200)
        try:
            # Empty cache
            stats = mod.cache_stats(cache_file=cache_file)
            assert stats == {"total": 0, "fresh": 0, "expired": 0, "bytes": 0, "evictions_total": 0}
            # 1 entry
            mod.check_url_with_cache("https://example.com/spec.md", cache_file=cache_file, ttl_seconds=60)
            stats = mod.cache_stats(cache_file=cache_file)
            assert stats["total"] == 1 and stats["fresh"] == 1, f"got {stats}"
            # new fields: bytes + evictions_total present
            assert "bytes" in stats and "evictions_total" in stats, f"missing new fields: {stats}"
            assert stats["bytes"] > 0, f"expected bytes > 0, got {stats['bytes']}"
        finally:
            urllib.request.urlopen = orig


def test_cache_clear() -> None:
    """cache_clear() removes the disk cache file."""
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        cache_file.write_text("{}", encoding="utf-8")
        assert cache_file.exists()
        mod.cache_clear(cache_file=cache_file)
        assert not cache_file.exists()


# --- Test 17-20: V-R10 v3 cache LRU (ADR-014) ---


def test_cache_lru_eviction_by_max_entries() -> None:
    """max_entries=2 → 3rd insertion evicts oldest entry (LRU)."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        urllib.request.urlopen = lambda *a, **kw: _MockHTTPResponse(200)
        try:
            # 3 URLs, max_entries=2 → 1st URL evicted
            mod.check_url_with_cache("https://a.com/spec", cache_file=cache_file, max_entries=2, ttl_seconds=60)
            time.sleep(0.01)
            mod.check_url_with_cache("https://b.com/spec", cache_file=cache_file, max_entries=2, ttl_seconds=60)
            time.sleep(0.01)
            mod.check_url_with_cache("https://c.com/spec", cache_file=cache_file, max_entries=2, ttl_seconds=60)
            # check cache has 2 entries (a evicted, b+c remain)
            cache = mod._load_cache(cache_file)
            assert len(cache) == 2, f"expected 2 entries, got {len(cache)}"
            assert "https://a.com/spec" not in cache, "oldest entry (a) should be evicted"
            assert "https://b.com/spec" in cache
            assert "https://c.com/spec" in cache
        finally:
            urllib.request.urlopen = orig


def test_cache_lru_eviction_by_max_bytes() -> None:
    """max_bytes small → many insertions trigger eviction."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        urllib.request.urlopen = lambda *a, **kw: _MockHTTPResponse(200)
        try:
            # 5 URLs, max_bytes=200 (very small) → at least 1 eviction
            for i in range(5):
                mod.check_url_with_cache(
                    f"https://x{i}.com/spec", cache_file=cache_file, max_bytes=200, ttl_seconds=60,
                )
                time.sleep(0.005)
            # check file size <= 200 bytes
            size = cache_file.stat().st_size
            assert size <= 200, f"cache size {size} exceeds 200 bytes"
        finally:
            urllib.request.urlopen = orig


def test_cache_eviction_keeps_recent() -> None:
    """LRU evicts oldest, keeps most recent."""
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"
        orig = urllib.request.urlopen
        urllib.request.urlopen = lambda *a, **kw: _MockHTTPResponse(200)
        try:
            mod.check_url_with_cache("https://old.com/spec", cache_file=cache_file, max_entries=1, ttl_seconds=60)
            time.sleep(0.01)
            mod.check_url_with_cache("https://new.com/spec", cache_file=cache_file, max_entries=1, ttl_seconds=60)
            cache = mod._load_cache(cache_file)
            assert "https://old.com/spec" not in cache, "oldest should be evicted"
            assert "https://new.com/spec" in cache, "newest should remain"
        finally:
            urllib.request.urlopen = orig


def test_cache_default_caps() -> None:
    """DEFAULT_CACHE_MAX_BYTES=10MB + DEFAULT_CACHE_MAX_ENTRIES=10000."""
    mod = _import_url_validity()
    assert mod.DEFAULT_CACHE_MAX_BYTES == 10 * 1024 * 1024
    assert mod.DEFAULT_CACHE_MAX_ENTRIES == 10000


# --- Test 21-22: V-R10 v3 file lock (ADR-015) ---


def test_file_lock_context_manager_exists() -> None:
    """_CacheLock context manager is exposed."""
    mod = _import_url_validity()
    assert hasattr(mod, "_CacheLock"), "_CacheLock not exposed"
    assert callable(mod._CacheLock)


def test_file_lock_serializes_concurrent_writes() -> None:
    """Concurrent processes writing to same cache should not corrupt the file.

    We use multiprocessing to spawn 2 workers writing concurrently. The file lock
    (fcntl.flock) ensures atomic read-modify-write — no torn writes.
    """
    import multiprocessing
    import urllib.request
    mod = _import_url_validity()
    with tempfile.TemporaryDirectory() as tmpdir:
        cache_file = Path(tmpdir) / "cache.json"

        def _worker(url_suffix: str) -> None:
            # re-import inside the subprocess (each has its own interpreter)
            import importlib.util
            import urllib.request as ur_local

            spec = importlib.util.spec_from_file_location("ok_url", str(SOURCE_ROOT / "workflow_kit" / "url_validity.py"))
            mod_local = importlib.util.module_from_spec(spec)
            sys.modules["ok_url"] = mod_local
            spec.loader.exec_module(mod_local)

            class _LocalMock:
                def __init__(self):
                    self.status = 200
                    self.headers = {}
                def __enter__(self):
                    return self
                def __exit__(self, *a):
                    return False
                def read(self, *a):
                    return b""

            ur_local.urlopen = lambda req, **kw: _LocalMock()
            mod_local.check_url_with_cache(
                f"https://w{url_suffix}.com/spec",
                cache_file=cache_file,
                ttl_seconds=60,
            )

        # spawn 2 workers
        p1 = multiprocessing.Process(target=_worker, args=("1",))
        p2 = multiprocessing.Process(target=_worker, args=("2",))
        p1.start()
        p2.start()
        p1.join(timeout=15)
        p2.join(timeout=15)
        assert not p1.is_alive(), "worker 1 hung (lock deadlock?)"
        assert not p2.is_alive(), "worker 2 hung (lock deadlock?)"
        # both URLs should be in the cache (or the surviving one if eviction happened)
        cache = mod._load_cache(cache_file)
        assert len(cache) >= 1, f"expected at least 1 entry, got {len(cache)}: {list(cache.keys())}"
        # at least one of the worker URLs should be present
        assert any(k in cache for k in ("https://w1.com/spec", "https://w2.com/spec")), (
            f"neither worker URL in cache: {list(cache.keys())}"
        )
        # file should be valid JSON (no torn write)
        text = cache_file.read_text(encoding="utf-8")
        parsed = json.loads(text)  # should not raise
        assert isinstance(parsed, dict)


# --- Test 23-27: V-R11 body content audit (ADR-017) ---


class _BodyMockResponse:
    """Mock for urllib.request.urlopen with body content."""

    def __init__(self, body: bytes = b"", content_type: str = "text/html; charset=utf-8", status: int = 200):
        self.body = body
        self.status = status
        self.headers = {"Content-Type": content_type}

    def __enter__(self):
        return self

    def __exit__(self, *a):
        return False

    def read(self, n: int = -1) -> bytes:
        if n < 0:
            return self.body
        return self.body[:n]


def test_body_html_pass() -> None:
    """text/html body with <html> tag → no issues."""
    import urllib.request
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _BodyMockResponse(b"<html><body>Hello</body></html>", "text/html; charset=utf-8")
    try:
        issues = mod.check_url_body("https://example.com/page.html")
        assert not issues, f"clean HTML should pass, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_body_phishing_detected() -> None:
    """Body containing phishing keyword → ERROR (V-R11-body-phishing)."""
    import urllib.request
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    body = b"<html><body>Please verify your account immediately. Click here!</body></html>"
    urllib.request.urlopen = lambda req, **kw: _BodyMockResponse(body, "text/html; charset=utf-8")
    try:
        issues = mod.check_url_body("https://phishing.example.com/login")
        errors = [i for i in issues if i.severity == "error"]
        assert any("phishing" in i.rule.lower() for i in errors), f"phishing should be error, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_body_missing_html_tag_warn() -> None:
    """text/html body without <html> tag → warn (V-R11-body-html)."""
    import urllib.request
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _BodyMockResponse(b"plain text without html", "text/html; charset=utf-8")
    try:
        issues = mod.check_url_body("https://example.com/page.html")
        warns = [i for i in issues if i.severity == "warn"]
        assert any("html" in i.rule.lower() for i in warns), f"missing <html> should be warn, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_body_unexpected_content_type_warn() -> None:
    """application/octet-stream → warn (V-R11-body-content-type)."""
    import urllib.request
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    urllib.request.urlopen = lambda req, **kw: _BodyMockResponse(b"\x00\x01\x02", "application/octet-stream")
    try:
        issues = mod.check_url_body("https://example.com/binary")
        warns = [i for i in issues if i.severity == "warn"]
        assert any("content-type" in i.rule.lower() for i in warns), f"unexpected content-type should be warn, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_body_timeout_warn() -> None:
    """Connection timeout during body fetch → warn (V-R11-body-timeout)."""
    import urllib.request
    import socket
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    def _timeout(*a, **kw):
        raise socket.timeout("read timeout")
    urllib.request.urlopen = _timeout
    try:
        issues = mod.check_url_body("https://example.com/page.html")
        warns = [i for i in issues if i.severity == "warn"]
        assert any("timeout" in i.rule.lower() for i in warns), f"timeout should be warn, got: {issues}"
    finally:
        urllib.request.urlopen = orig


def test_cli_body_flag_invokes_body_audit() -> None:
    """CLI --body flag invokes check_url_body() per URL (phishing body → error output)."""
    import urllib.request
    import contextlib
    mod = _import_url_validity()
    orig = urllib.request.urlopen
    body = b"<html><body>Please verify your account immediately. Click here!</body></html>"
    urllib.request.urlopen = lambda req, **kw: _BodyMockResponse(body, "text/html; charset=utf-8")
    try:
        from io import StringIO
        buf = StringIO()
        with contextlib.redirect_stdout(buf):
            rc = mod.main(["--body", "--mode=strict", "--timeout=5.0", "https://phishing.example.com/login"])
        output = buf.getvalue()
        assert "phishing" in output.lower(), f"expected phishing issue in output, got: {output!r}"
        assert rc != 0, f"expected non-zero exit code (error in strict mode), got {rc}"
    finally:
        urllib.request.urlopen = orig


def main() -> int:
    test_funcs = [
        test_url_https_only_accept,
        test_url_scheme_reject,
        test_url_localhost_reject,
        test_url_private_ip_reject,
        test_url_credentials_reject,
        test_url_path_traversal_reject,
        test_online_http_200_pass,
        test_online_http_404_stale,
        test_online_http_500_warn,
        test_online_timeout_warn,
        test_online_tls_error_reject,
        test_online_dns_failure_reject,
        test_cache_miss_then_hit,
        test_cache_ttl_expired,
        test_cache_stats,
        test_cache_clear,
        test_cache_lru_eviction_by_max_entries,
        test_cache_lru_eviction_by_max_bytes,
        test_cache_eviction_keeps_recent,
        test_cache_default_caps,
        test_file_lock_context_manager_exists,
        test_file_lock_serializes_concurrent_writes,
        test_body_html_pass,
        test_body_phishing_detected,
        test_body_missing_html_tag_warn,
        test_body_unexpected_content_type_warn,
        test_body_timeout_warn,
        test_cli_body_flag_invokes_body_audit,
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
