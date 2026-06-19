"""v0.2.0 PoC backend-knowledge standalone tool.

umbrella doc §1.2 G7 + ADR-0035 정합:
- 완전 standalone (다른 backend 연결 ❌, OIDC ❌)
- 외부 시스템 7종 source 만 단방향 (Gitea 4 + homelab + metrics + hrdb, M-v0.2.0+ 확장)
- Path Y caller-provided user context (auth 자체 안 함, X-DevHub-User-Context header 로 governance)
- OKF concept format (Google OKF v0.1, ADR-0034)
- 봉투 암호화 (AES-256-GCM, scope = raw + .env/KEK 만, ADR-0025)
"""

__version__ = "0.2.0"
