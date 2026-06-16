"""workflow_kit.url_validity — URL validity check (V-R10, v0.7.35 PoC + v0.7.36 v2 cache).

ADR-010 (offline 8 check) + ADR-012 (online HEAD) + ADR-013 (24h disk cache + smart retry).
wiki frontmatter `resource` (ADR-006) + `last_ingested_from` URL 의 validity 검증.

Usage:
    from workflow_kit.url_validity import check_url, check_url_online, check_url_with_cache

    issues = check_url("https://example.com/spec.md")              # offline only
    issues = check_url_online("https://example.com/spec.md")       # online HEAD (opt-in)
    issues = check_url_with_cache("https://example.com/spec.md")   # with 24h disk cache

CLI:
    python -m workflow_kit.url_validity <url>...                    # offline
    python -m workflow_kit.url_validity <url>... --online            # online HEAD
    python -m workflow_kit.url_validity <url>... --online --cache    # with cache
"""

from __future__ import annotations

import argparse
import ipaddress
import json
import re
import socket
import ssl
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Literal
from urllib.parse import urlparse


# ---------------------------------------------------------------------------
# Severity
# ---------------------------------------------------------------------------
Severity = Literal["error", "warn"]


# ---------------------------------------------------------------------------
# Issue dataclass
# ---------------------------------------------------------------------------
@dataclass(frozen=True)
class UrlIssue:
    """Single URL validity issue."""

    rule: str
    severity: Severity
    message: str


# ---------------------------------------------------------------------------
# Offline check (ADR-010)
# ---------------------------------------------------------------------------
_ALLOWED_SCHEMES: frozenset[str] = frozenset({"https"})
_LOCALHOST_NAMES: frozenset[str] = frozenset({"localhost"})


def _is_private_ip(host: str) -> bool:
    raw = host.strip("[]")
    try:
        ip = ipaddress.ip_address(raw)
    except ValueError:
        return False
    return ip.is_private or ip.is_loopback or ip.is_link_local or ip.is_reserved or ip.is_multicast


def _is_github_url(url: str) -> bool:
    try:
        parsed = urlparse(url)
    except ValueError:
        return False
    return parsed.hostname in ("github.com", "www.github.com")


def _has_credentials(url: str) -> bool:
    try:
        parsed = urlparse(url)
    except ValueError:
        return False
    return bool(parsed.username) or bool(parsed.password)


def _has_path_traversal(url: str) -> bool:
    try:
        parsed = urlparse(url)
    except ValueError:
        return False
    if not parsed.path:
        return False
    return any(seg == ".." for seg in parsed.path.split("/"))


def check_url(url: str) -> list[UrlIssue]:
    """Run all offline V-R10 checks on a URL. ADR-010."""
    issues: list[UrlIssue] = []
    try:
        parsed = urlparse(url)
    except ValueError as e:
        issues.append(UrlIssue(rule="V-R10-parse", severity="error", message=f"URL parse failed: {e}"))
        return issues

    scheme = parsed.scheme.lower()
    if not scheme:
        issues.append(UrlIssue(rule="V-R10-scheme", severity="error", message="URL missing scheme"))
    elif scheme not in _ALLOWED_SCHEMES:
        issues.append(UrlIssue(rule="V-R10-scheme", severity="error", message=f"scheme {scheme!r} not allowed (only https://)"))

    host = parsed.hostname or ""
    if not host:
        if scheme:
            issues.append(UrlIssue(rule="V-R10-host", severity="error", message="URL missing host"))
    else:
        if host.lower() in _LOCALHOST_NAMES or host.lower().endswith(".local"):
            issues.append(UrlIssue(rule="V-R10-localhost", severity="error", message=f"localhost host {host!r} not allowed"))
        elif _is_private_ip(host):
            issues.append(UrlIssue(rule="V-R10-private-ip", severity="error", message=f"private/internal IP {host!r} not allowed"))

    if _has_path_traversal(url):
        issues.append(UrlIssue(rule="V-R10-traversal", severity="error", message="URL path contains `..` segment (traversal)"))

    if scheme == "file":
        issues.append(UrlIssue(rule="V-R10-file-scheme", severity="error", message="file:// scheme not allowed"))

    if _has_credentials(url):
        issues.append(UrlIssue(rule="V-R10-credentials", severity="error", message="URL contains user:pass@ credentials (security risk)"))

    if _is_github_url(url):
        path = parsed.path
        parts = [p for p in path.split("/") if p]
        if len(parts) >= 3 and parts[2] not in ("blob", "tree", "raw", "commits", "issues", "pull", "wiki", "actions"):
            issues.append(UrlIssue(rule="V-R10-github-form", severity="warn", message=f"github.com URL form unusual: {url!r}"))

    return issues


# ---------------------------------------------------------------------------
# Online HEAD layer (ADR-012)
# ---------------------------------------------------------------------------
def check_url_online(
    url: str,
    *,
    timeout: float = 10.0,
    user_agent: str = "workflow-kit-url-validity/0.7.35",
    max_retries: int = 3,
) -> list[UrlIssue]:
    """Online V-R10 check: HTTP HEAD request to detect stale URL. ADR-012.

    Returns UrlIssue list:
    - HTTP 200 OK: pass (no issues)
    - HTTP 3xx: follow 1 hop, recheck
    - HTTP 404/410: ERROR (stale URL)
    - HTTP 5xx: WARN (transient, smart retry with exponential backoff)
    - 429: WARN (back off)
    - Connection timeout: WARN (slow host)
    - TLS error: ERROR (security risk)
    - DNS failure: ERROR (host does not exist)
    """
    issues: list[UrlIssue] = []
    last_status: int | None = None
    last_error: str | None = None
    headers: dict[str, str] | None = None

    for attempt in range(max_retries + 1):
        try:
            req = urllib.request.Request(url, method="HEAD")
            req.add_header("User-Agent", user_agent)
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                last_status = resp.status
                last_error = None
                headers = dict(resp.headers) if resp.headers else None
                break
        except urllib.error.HTTPError as e:
            last_status = e.code
            last_error = e.reason
            if 400 <= last_status < 500 and last_status != 429:
                break
            if attempt < max_retries:
                time.sleep(2 ** attempt)
                continue
            break
        except socket.timeout:
            last_error = "timeout"
            if attempt < max_retries:
                time.sleep(2 ** attempt)
                continue
            break
        except (ssl.SSLError, urllib.error.URLError, OSError) as e:
            # TLS / DNS / connection — no retry, surface immediately
            if isinstance(e, ssl.SSLError):
                return [UrlIssue(rule="V-R10-online-tls", severity="error", message=f"TLS error: {e}")]
            # URLError / OSError → DNS or connection
            reason = getattr(e, "reason", str(e))
            return [UrlIssue(rule="V-R10-online-url-error", severity="error", message=f"URL error (DNS / connection): {reason}")]

    if last_error == "timeout":
        issues.append(UrlIssue(rule="V-R10-online-timeout", severity="warn", message=f"connection timeout after {timeout}s (×{max_retries + 1} retries)"))
        return issues
    if last_status is None:
        return issues

    status = last_status
    if 200 <= status < 300:
        return issues
    if 300 <= status < 400 and status != 304:
        if headers and headers.get("Location"):
            try:
                return check_url_online(headers["Location"], timeout=timeout, user_agent=user_agent, max_retries=max_retries)
            except Exception as e:  # noqa: BLE001
                issues.append(UrlIssue(rule="V-R10-online-redirect", severity="warn", message=f"redirect target failed: {e}"))
                return issues
        return issues
    if status in (404, 410):
        issues.append(UrlIssue(rule="V-R10-online-stale", severity="error", message=f"HTTP {status}: URL appears stale (not found / gone)"))
    elif 500 <= status < 600:
        issues.append(UrlIssue(rule="V-R10-online-server-error", severity="warn", message=f"HTTP {status}: server error (transient after {max_retries} retries)"))
    elif status == 429:
        issues.append(UrlIssue(rule="V-R10-online-rate-limit", severity="warn", message="HTTP 429: rate limited (back off)"))
    else:
        issues.append(UrlIssue(rule="V-R10-online-unexpected", severity="warn", message=f"HTTP {status}: unexpected status"))

    return issues


# ---------------------------------------------------------------------------
# Body content audit (ADR-017, v0.7.37+)
# ---------------------------------------------------------------------------
DEFAULT_BODY_MAX_BYTES: int = 1 * 1024 * 1024  # 1 MB body cap (ADR-017)
PHISHING_KEYWORDS: tuple[str, ...] = (
    "verify your account",
    "click here immediately",
    "your account will be suspended",
    "urgent action required",
    "confirm your password",
    "wire transfer",
    "lottery winner",
    "nigerian prince",
)


def check_url_body(
    url: str,
    *,
    timeout: float = 10.0,
    user_agent: str = "workflow-kit-url-validity/0.7.37",
    max_body_bytes: int = DEFAULT_BODY_MAX_BYTES,
    follow_redirect: bool = True,
    max_redirects: int = 3,
) -> list[UrlIssue]:
    """V-R11 body content audit: GET request + content checks. ADR-017.

    Body checks (4 case):
    1. Content-Type: text/html expected, application/json OK, others warn
    2. Body size: 0 bytes = warn, > max_body_bytes = warn (truncated)
    3. Phishing keywords: detected in body = ERROR
    4. HTML renderable: <html> tag detection for text/html

    Returns:
        list[UrlIssue]:
        - HTTP 200: continue body checks
        - HTTP 4xx/5xx: surface as error/warn (similar to online)
        - timeout/TLS/DNS: surface as error/warn
    """
    import socket
    import ssl
    issues: list[UrlIssue] = []
    try:
        req = urllib.request.Request(url, method="GET")
        req.add_header("User-Agent", user_agent)
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            content_type = resp.headers.get("Content-Type", "")
            body_bytes = resp.read(max_body_bytes + 1)  # read 1 extra to detect truncation
    except urllib.error.HTTPError as e:
        issues.append(UrlIssue(rule="V-R11-body-http-error", severity="warn", message=f"HTTP {e.code}: {e.reason}"))
        return issues
    except socket.timeout:
        issues.append(UrlIssue(rule="V-R11-body-timeout", severity="warn", message=f"connection timeout after {timeout}s"))
        return issues
    except (ssl.SSLError, urllib.error.URLError, OSError) as e:
        if isinstance(e, ssl.SSLError):
            issues.append(UrlIssue(rule="V-R11-body-tls", severity="error", message=f"TLS error: {e}"))
        else:
            reason = getattr(e, "reason", str(e))
            issues.append(UrlIssue(rule="V-R11-body-url-error", severity="error", message=f"URL error: {reason}"))
        return issues

    # 1. Content-Type check
    if not content_type:
        issues.append(UrlIssue(rule="V-R11-body-content-type", severity="warn", message="missing Content-Type header"))
    elif "text/html" not in content_type.lower() and "application/json" not in content_type.lower() and "text/" not in content_type.lower():
        issues.append(UrlIssue(rule="V-R11-body-content-type", severity="warn", message=f"unexpected Content-Type: {content_type!r}"))

    # 2. Body size check
    if len(body_bytes) == 0:
        issues.append(UrlIssue(rule="V-R11-body-empty", severity="warn", message="empty body"))
    elif len(body_bytes) > max_body_bytes:
        issues.append(UrlIssue(rule="V-R11-body-truncated", severity="warn", message=f"body exceeds {max_body_bytes} bytes (truncated)"))
        body_bytes = body_bytes[:max_body_bytes]

    # 3. Phishing keyword check
    try:
        body_text = body_bytes.decode("utf-8", errors="ignore").lower()
    except UnicodeDecodeError:
        body_text = ""
    for kw in PHISHING_KEYWORDS:
        if kw in body_text:
            issues.append(UrlIssue(rule="V-R11-body-phishing", severity="error", message=f"phishing keyword detected: {kw!r}"))
            break  # one phishing keyword is enough

    # 4. HTML renderable check (only for text/html)
    if "text/html" in content_type.lower():
        if b"<html" not in body_bytes.lower():
            issues.append(UrlIssue(rule="V-R11-body-html", severity="warn", message="text/html body missing <html> tag"))

    return issues


# Disk cache layer (ADR-013, v0.7.36+) + size cap + LRU (ADR-014, v0.7.37+)
# ---------------------------------------------------------------------------
DEFAULT_CACHE_TTL_SECONDS: int = 86400
DEFAULT_CACHE_FILE: Path = Path.home() / ".workflow_kit" / "url_validity_cache.json"
DEFAULT_CACHE_MAX_BYTES: int = 10 * 1024 * 1024  # 10 MB
DEFAULT_CACHE_MAX_ENTRIES: int = 10000  # LRU cap


@dataclass(frozen=True)
class CacheEntry:
    url: str
    timestamp: float
    issues: tuple[dict, ...]


def _load_cache(cache_file: Path) -> dict[str, CacheEntry]:
    if not cache_file.exists():
        return {}
    try:
        raw = json.loads(cache_file.read_text(encoding="utf-8"))
        return {
            url: CacheEntry(url=url, timestamp=float(data["timestamp"]), issues=tuple(data["issues"]))
            for url, data in raw.items()
        }
    except (json.JSONDecodeError, KeyError, ValueError, OSError):
        return {}


# Module-level counter for total LRU evictions (ADR-014 follow-up)
_evictions_total: int = 0


class _CacheLock:
    """Context manager for cross-process file lock on cache file. ADR-015.

    Uses `fcntl.flock` for advisory locking (POSIX only).
    Falls back to no-op on Windows (non-blocking). Emits WARN on failure.
    """

    def __init__(self, cache_file: Path):
        self.cache_file = cache_file
        self._lock_path: Path | None = None
        self._fd = None

    def __enter__(self):
        try:
            import fcntl  # POSIX only
        except ImportError:
            # Windows: no-op
            return self
        # Use a sidecar lock file (cache_file.lock) to avoid interfering with cache reads
        self.cache_file.parent.mkdir(parents=True, exist_ok=True)
        self._lock_path = self.cache_file.with_suffix(self.cache_file.suffix + ".lock")
        try:
            self._fd = open(self._lock_path, "w")
            fcntl.flock(self._fd.fileno(), fcntl.LOCK_EX)
        except OSError as e:
            import sys
            print(f"WARN: failed to acquire cache file lock: {e}", file=sys.stderr)
            if self._fd:
                self._fd.close()
                self._fd = None
        return self

    def __exit__(self, *args):
        if self._fd is not None:
            try:
                import fcntl
                fcntl.flock(self._fd.fileno(), fcntl.LOCK_UN)
            except (OSError, ImportError):
                pass
            self._fd.close()
            self._fd = None
        return False




def _save_cache(
    cache_file: Path,
    entries: dict[str, CacheEntry],
    max_bytes: int = DEFAULT_CACHE_MAX_BYTES,
    max_entries: int = DEFAULT_CACHE_MAX_ENTRIES,
) -> None:
    """Save cache to disk JSON. Enforce size cap + LRU eviction (ADR-014).

    Strategy:
    1. Serialize → check size + entry count
    2. If over either cap: evict oldest (LRU by timestamp) until under both caps
    3. Write. Warn if single entry exceeds cap (best-effort).
    """
    global _evictions_total
    cache_file.parent.mkdir(parents=True, exist_ok=True)
    raw = {url: {"timestamp": entry.timestamp, "issues": list(entry.issues)} for url, entry in entries.items()}
    serialized = json.dumps(raw, indent=2, sort_keys=True)
    size = len(serialized.encode("utf-8"))

    # LRU eviction: sort by timestamp ascending (oldest first)
    while (size > max_bytes or len(entries) > max_entries) and entries:
        oldest_url = min(entries.keys(), key=lambda u: entries[u].timestamp)
        del entries[oldest_url]
        _evictions_total += 1
        raw = {url: {"timestamp": entry.timestamp, "issues": list(entry.issues)} for url, entry in entries.items()}
        serialized = json.dumps(raw, indent=2, sort_keys=True)
        size = len(serialized.encode("utf-8"))

    cache_file.write_text(serialized, encoding="utf-8")

    if size > max_bytes:
        import sys
        print(f"WARN: cache size {size} bytes exceeds cap {max_bytes} (single entry too large)", file=sys.stderr)

def _cache_lookup(cache: dict[str, CacheEntry], url: str, ttl: int) -> list[UrlIssue] | None:
    entry = cache.get(url)
    if entry is None:
        return None
    if time.time() - entry.timestamp > ttl:
        return None
    return [UrlIssue(rule=i["rule"], severity=i["severity"], message=i["message"]) for i in entry.issues]


def check_url_with_cache(
    url: str,
    *,
    cache_file: Path | None = None,
    ttl_seconds: int = DEFAULT_CACHE_TTL_SECONDS,
    timeout: float = 10.0,
    user_agent: str = "workflow-kit-url-validity/0.7.35",
    max_retries: int = 3,
    max_bytes: int = DEFAULT_CACHE_MAX_BYTES,
    max_entries: int = DEFAULT_CACHE_MAX_ENTRIES,
) -> list[UrlIssue]:
    """V-R10 online HEAD with disk cache (ADR-013) + size cap + LRU (ADR-014) + file lock (ADR-015)."""
    cache_file = cache_file or DEFAULT_CACHE_FILE
    with _CacheLock(cache_file):
        cache = _load_cache(cache_file)
        cached = _cache_lookup(cache, url, ttl_seconds)
        if cached is not None:
            return cached
        issues = check_url_online(url, timeout=timeout, user_agent=user_agent, max_retries=max_retries)
        cache[url] = CacheEntry(
            url=url,
            timestamp=time.time(),
            issues=tuple({"rule": i.rule, "severity": i.severity, "message": i.message} for i in issues),
        )
        _save_cache(cache_file, cache, max_bytes=max_bytes, max_entries=max_entries)
    return issues


def cache_clear(cache_file: Path | None = None) -> None:
    cache_file = cache_file or DEFAULT_CACHE_FILE
    if cache_file.exists():
        cache_file.unlink()


def cache_stats(cache_file: Path | None = None) -> dict[str, int]:
    cache_file = cache_file or DEFAULT_CACHE_FILE
    cache = _load_cache(cache_file)
    now = time.time()
    total = len(cache)
    fresh = sum(1 for e in cache.values() if now - e.timestamp < DEFAULT_CACHE_TTL_SECONDS)
    # bytes: actual on-disk size of the cache file (after possible LRU eviction in last save)
    bytes_size = cache_file.stat().st_size if cache_file.exists() else 0
    return {
        "total": total,
        "fresh": fresh,
        "expired": total - fresh,
        "bytes": bytes_size,
        "evictions_total": _evictions_total,
    }

def _build_arg_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="workflow_kit.url_validity", description="V-R10 URL validity check")
    p.add_argument("urls", nargs="*", help="URLs to check")
    p.add_argument("--mode", choices=["strict", "loose"], default="strict", help="lint mode")
    p.add_argument("--online", action="store_true", help="run online HEAD check (ADR-012)")
    p.add_argument("--cache", action="store_true", help="use 24h disk cache (ADR-013)")
    p.add_argument("--ttl", type=int, default=DEFAULT_CACHE_TTL_SECONDS, help=f"cache TTL (default: {DEFAULT_CACHE_TTL_SECONDS})")
    p.add_argument("--max-retries", type=int, default=3, help="max retries for 5xx/429/timeout (default: 3)")
    p.add_argument("--timeout", type=float, default=10.0, help="online HEAD/body timeout in seconds (default: 10.0)")
    p.add_argument("--max-entries", type=int, default=DEFAULT_CACHE_MAX_ENTRIES, help=f"cache entry count cap (default: {DEFAULT_CACHE_MAX_ENTRIES})")
    p.add_argument("--cache-stats", action="store_true", help="print cache statistics and exit")
    p.add_argument("--cache-clear", action="store_true", help="clear disk cache and exit")
    p.add_argument("--body", action="store_true", help="run V-R11 body content audit (ADR-017, opt-in)")
    p.add_argument("--max-body-bytes", type=int, default=DEFAULT_BODY_MAX_BYTES, help=f"body size cap in bytes (default: {DEFAULT_BODY_MAX_BYTES})")
    return p


def main(argv: list[str] | None = None) -> int:
    args = _build_arg_parser().parse_args(argv)

    if args.cache_stats:
        print(f"Cache stats: {cache_stats()}")
        return 0
    if args.cache_clear:
        cache_clear()
        print("Cache cleared.")
        return 0

    if not args.urls:
        print("ERROR: no URLs provided (use positional args, or --cache-stats / --cache-clear)", file=sys.stderr)
        return 2

    failed = 0
    for url in args.urls:
        issues = check_url(url)
        if args.online:
            if args.cache:
                online_issues = check_url_with_cache(url, ttl_seconds=args.ttl, timeout=args.timeout, max_retries=args.max_retries, max_bytes=args.max_bytes, max_entries=args.max_entries)
            else:
                online_issues = check_url_online(url, timeout=args.timeout, max_retries=args.max_retries)
            issues.extend(online_issues)
        if args.body:
            body_issues = check_url_body(url, timeout=args.timeout, max_body_bytes=args.max_body_bytes)
            issues.extend(body_issues)
        if args.mode == "loose":
            print(f"  PASS  {url}")
        else:
            for issue in issues:
                print(f"  [{issue.severity.upper()}] {issue.rule} {url}: {issue.message}")
                if issue.severity == "error":
                    failed += 1
    return 1 if failed > 0 else 0


if __name__ == "__main__":
    sys.exit(main())
