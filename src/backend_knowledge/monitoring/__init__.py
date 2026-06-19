"""Monitoring + metrics (umbrella doc §3.6.6.3 + §11.3 + §17.5 정합).

28 metrics M-v0.2.3+ production:
- 5 base metrics (§11.3)
- 13 governance metrics (§3.6.6.3 — 4 layer × 4 metric type + 1 event type)
- 5 Pi LLM metrics (§3.5.7.5, M-v0.2.3+ scope)
- 1 false positive metric (§3.5.8.3, M-v0.2.3+ scope)
- 4 API versioning metrics (§16.4, M-v0.3.0+ scope)

PoC (M-v0.2.0): 5 base + 13 governance = 18 metrics implemented.
Remaining 10 are stubbed (deferred to M-v0.2.3+ / M-v0.3.0+).
"""
