# backend-knowledge frontend (SvelteKit) — Architecture Design

- 문서 목적: `backend-knowledge` v0.2.0+ standalone backend 의 관리용 frontend 의 main design doc. tech stack (Svelte 5 + SvelteKit 2 + TypeScript 5) + 5 page scope (umbrella doc §12) + Path Y dev fixture 정공법 + tier 분리.
- 범위: §3 tech stack / §4 아키텍처 (routes + lib + components) / §5 5 page scope / §6 Path Y dev fixture / §7 API client / §8 build + dev workflow / §9 CI / §10 forward-looking.
- 대상 독자: M-v0.2.0+ frontend sprint 진입자, 후속 contributor, PR reviewer, 운영자.
- 상태: **accepted** (2026-06-19, M-v0.2.0 PoC + frontend 1차 release, retro-design recovery)
- 최종 수정일: 2026-06-19
- 관련 문서: [`../architecture.md`](./architecture.md) (main design) / [`../tech-stack.md`](./tech-stack.md) (runtime + frontend stack) / [`../README.md`](./README.md) / [`../../planning/release_v0-2_roadmap.md`](../../planning/release_v0-2_roadmap.md) §12 (frontend page 상세화) / [ADR-0035 backend-knowledge 신설](../../adr/0035-backend-knowledge-creation.md)

## 1. 개요

### 1.1 컨셉
- **§1.2 G7 standalone 정공법** (umbrella doc) + **§12 frontend page 상세화** (M-v0.2.1+): backend-knowledge 의 **관리/조회용 standalone frontend**.
- **위치**: `backend-knowledge/web/` 별도 standalone web (devhub frontend 와 **완전 분리**), umbrella doc §12.1 정합.
- **독립성**: 다른 frontend (devhub Next.js 16) 와 code / state / build 공유 ❌. backend-knowledge backend 만 호출 (§1.2 G7).
- **M-v0.2.0+ frontend 0 page** → **M-v0.2.0 PoC + frontend 1차 release** (사용자 2026-06-19 결정): 5 page 점진 도입.

### 1.2 위치 + tier
- **layer**: 3 group (routes / lib / components) + 1 utility (config)
- **directory**: `backend-knowledge/web/` (sibling to `backend-knowledge/src/`, ADR-0035 §3 정합)
- **tier**: **사외** (standalone 정공법, 사내 한정 정보 0, gateway 3-step orchestration 정공법, umbrella doc §12.3)

## 2. tech stack (자세한 내용: [`tech-stack.md`](./tech-stack.md) §1.2)

- **SvelteKit 2.x** (meta-framework, file-based routing + server-side load + form action)
- **Svelte 5.x** (runes: `$state` / `$derived` / `$effect`, 2024-10 stable)
- **TypeScript 5.x** (backend 정합)
- **Vite 5.x** (SvelteKit default bundler)
- **Node.js 20 LTS** (dev/prod runtime)
- **Vitest** (Vite native test runner, M-v0.2.0+ scope)
- **Playwright** (E2E, M-v0.2.1+ scope)
- **styling**: plain CSS (PoC scope) — Tailwind CSS optional (M-v0.2.1+)
- **http client**: native `fetch` + `src/lib/api.ts` wrapper (Path Y header auto-inject)

## 3. 아키텍처

### 3.1 3 layer (routes / lib / components) + 1 utility

| Layer | Path | Import 가능 | Import 금지 |
|---|---|---|---|
| **routes** | `src/routes/**/*.svelte`, `+page.ts`, `+page.server.ts`, `+layout.svelte` | lib, components, types | 다른 route (cross-route ❌) |
| **lib** | `src/lib/{api,path-y,types}.ts` | (독립) | routes, components (단독) |
| **components** | `src/lib/components/{Sidebar,Header,PathYDevFixture,...}.svelte` | lib, types | routes (단독) |
| **config** (utility) | `svelte.config.js`, `vite.config.ts`, `package.json` | (utility 끼리) | 다른 layer ❌ |

### 3.2 File-based routing (SvelteKit 2)
- `src/routes/+page.svelte` → `/`
- `src/routes/concepts/+page.svelte` → `/concepts`
- `src/routes/concepts/[type]/[name]/+page.svelte` → `/concepts/{type}/{name}` (dynamic param)
- `src/routes/bundles/[name]/+page.svelte` → `/bundles/{name}` (dynamic param)
- `+layout.svelte` → 모든 page 의 공통 layout (sidebar + header + Path Y dev fixture)

### 3.3 SSR vs CSR
- **SSR (default)**: SvelteKit 의 default rendering. `+page.server.ts` 의 `load` function 으로 backend API 호출 (서버 측 fetch, X-DevHub-User-Context header server-side 주입).
- **CSR (선택)**: `+page.ts` (universal load) 또는 component 내 `fetch` (client-side).
- **M-v0.2.0 PoC default**: SSR (CSR 필요 시 component 내 `onMount` + `fetch`).
- **Trade-off**: SSR = SEO + initial load 빠름, CSR = interactive UI 단순.

## 4. 5 page scope (umbrella doc §12.1~§12.5)

| # | Path | Page | 기능 |
|---|---|---|---|
| 1 | `/` | **Dashboard** | 30 endpoint status (health) + recent audit event (10 row) + sync trigger shortcut (per source dropdown) |
| 2 | `/concepts` | **Concept list** | 8 type (dataset/metric/api_endpoint/runbook/integration/event/reference/decision) × 4 source (gitea 4 sub-plugin) + substring search + bundle filter |
| 3 | `/concepts/{type}/{name}` | **Concept detail** | frontmatter 12 field + body (Markdown render) + cross-link graph (in-link + out-link + count) + change history (per audit.concept.archive / publish event) |
| 4 | `/ingest` | **Ingest trigger** | 5 source dropdown (homelab_mock / gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action) + sync button + dry_run toggle + since datetime + result display (synced/failed/raw_ids) |
| 5 | `/bundles` | **Bundle list** | 5 bundle dropdown (homelab / gitea 4) + create button (owner_org_id + visibility) + rebuild button (per bundle) + concept_count + viz.html link |
| 6 | `/bundles/{name}` | **Bundle detail** | bundle metadata (owner_org_id / visibility / created_at) + concept list (per type) + rebuild button + viz.html iframe (full-screen toggle) |
| 7 | `/raw` | **Raw list** | source filter (5 source) + since datetime + 50 row page (raw_id / type / name / sha256 / size / registered_at / visibility) + register raw form (PoC: read-only) |
| 8 | `/audit` | **Audit viewer** | event_type dropdown (7 type) + user_id input + from/to datetime + 100 row page (envelope + data) + per concept / per user sub-page |

**총 8 route** (5 page group, 1차 release). 5 page (umbrella §12) + 추가 (dashboard + raw + audit) = 8.

## 5. Path Y dev fixture (§12.3 정공법)

### 5.1 정공법
- **M-v0.2.0 PoC default** (gateway 없이 standalone): `src/lib/components/PathYDevFixture.svelte` — sidebar 하단의 input box 에 user context (base64url JSON 7 field) 수동 입력 → localStorage 저장 → 모든 `fetch` 호출 시 `X-DevHub-User-Context` header 자동 주입
- **M-v0.2.1+** (gateway 통합): 별도 gateway 가 Path Y 주입, frontend 는 session cookie 만 받음. dev fixture 제거.

### 5.2 8 field schema
```typescript
interface PathYUserContext {
  version: 'v0';
  user_id: string;
  org_id: string;
  org_unit_ids: string[];
  project_ids: string[];
  roles: string[];
  request_id: string;
  issued_at: string;  // ISO 8601
}
```

### 5.3 Default dev fixture
- user_id: `u_dev_001`
- org_id: `ou_dev_dept_a`
- roles: `['developer', 'project_leader:prj_dev']`
- issued_at: 현재 시각 (만료 5분 자동)
- **submit** 클릭 시 base64url encode → localStorage `path_y_context` 저장

## 6. API client

### 6.1 `src/lib/api.ts` wrapper
```typescript
import type { PathYUserContext } from './path-y';

export interface ApiOptions {
  pathY: PathYUserContext | null;
  baseUrl: string;  // default 'http://localhost:8000'
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit & { api: ApiOptions },
): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.api.pathY) {
    const encoded = btoa(JSON.stringify(options.api.pathY));
    headers.set('X-DevHub-User-Context', encoded);
  }
  const res = await fetch(`${options.api.baseUrl}${path}`, { ...options, headers });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.detail?.code || 'E_UNKNOWN', body.detail?.message || res.statusText);
  }
  return res.json();
}

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(`${status} ${code}: ${message}`);
  }
}
```

### 6.2 Endpoint wrapper (type-safe)
```typescript
// src/lib/api.ts (continued)
export const api = {
  health: () => apiFetch<{ status: string; version: string }>('/health', { method: 'GET', api: { ... } }),
  listConcepts: (params: { q?: string; bundle?: string; type?: string }) =>
    apiFetch<{ items: Concept[]; total: number }>(`/api/v0-2/search?${new URLSearchParams(params)}`, { method: 'GET', api: { ... } }),
  getConcept: (bundle: string, type: string, name: string) =>
    apiFetch<{ data: Concept }>(`/api/v0-2/concepts/${type}/${name}?bundle=${bundle}`, { method: 'GET', api: { ... } }),
  // ... 30 endpoint wrapper
};
```

### 6.3 `src/lib/types.ts` — Pydantic v2 schema mirror
```typescript
// 1:1 mirror of backend src/backend_knowledge/okf/frontmatter.py
export type ConceptType = 'dataset' | 'metric' | 'api_endpoint' | 'runbook' | 'integration' | 'event' | 'reference' | 'decision';
export type Visibility = 'public' | 'org' | 'personal' | 'project';

export interface ConceptFrontmatter {
  type: ConceptType;
  title?: string;
  description?: string;
  resource?: string;
  tags?: string[];
  timestamp?: string;
  x_devhub_source?: string;
  x_devhub_bundle?: string;
  x_devhub_version: number;
  x_devhub_curator: 'rule-based' | 'llm-system_admin' | 'human-self-user' | 'human-org-head' | 'human-system-admin';
  x_devhub_owner_org_id?: string;
  x_devhub_owner_user_id?: string;
  x_devhub_owner_org_unit_ids?: string[];
  x_devhub_owner_project_ids?: string[];
  x_devhub_visibility: Visibility;
}

export interface Concept {
  concept_id: string;
  type: string;
  name: string;
  bundle: string;
  frontmatter: ConceptFrontmatter;
  body: string;
  // ...
}
```

## 7. Build + dev workflow

### 7.1 dev server
```bash
cd backend-knowledge/web
npm install
npm run dev  # http://localhost:5173
```

### 7.2 build
```bash
npm run build  # SvelteKit build → .svelte-kit/output + build/
npm run preview  # preview built bundle
```

### 7.3 type check
```bash
npm run check  # svelte-kit sync + svelte-check (TypeScript + Svelte type)
```

### 7.4 test (M-v0.2.0 PoC)
```bash
npm run test:unit  # Vitest (5 page component test, 30+ test)
npm run test:e2e  # Playwright (5 page navigation + form interaction, M-v0.2.1+)
```

## 8. CI (umbrella doc §6.5.4 E2E smoke 정합)

### 8.1 M-v0.2.0 PoC default
- **Frontend build check** (`npm ci && npm run build`) — CI pre-merge
- **Vitest** (component test) — CI pre-merge
- (Playwright E2E 는 M-v0.2.1+ scope)

### 8.2 GitHub Actions workflow
- `frontend-unit-tests` job — `npm ci && npm test:unit`
- `frontend-build-artifacts` job — `npm ci && npm run build` (artifacts upload)
- `frontend-e2e-smoke` job — `npx playwright test` (M-v0.2.1+)

## 9. Forward-looking (M-v0.2.1+ scope)

| Milestone | scope | 영향 |
|---|---|---|
| **M-v0.2.1** | 1차 frontend 완성 (이 doc 정합) + 5 page + Vitest + Playwright E2E + build CI | `backend-knowledge/web/` 디렉터리 + 8 route + Path Y dev fixture + Vitest 30+ test |
| **M-v0.2.1+** | gateway 통합 (M-v0.2.0 PoC standalone → M-v0.2.1 gateway) | Path Y dev fixture 제거 + session cookie + OIDC |
| **M-v0.2.1+** | 시각화 강화 (viz.html incoming edge + D3.js graph + bundle graph) | `viz.html` iframe + Cytoscape.js client-side |
| **M-v0.2.2** | Tailwind CSS + 디자인 시스템 | `tailwind.config.js` + theme token |
| **M-v0.2.3+** | multi-vendor LLM UI (Pi SDK 활용) + RAG chat | `/chat` route + LLM streaming + citation |
| **M-v0.3.0+** | Pi (pi.dev) LLM UI (M-v0.2.3+ scope) + cross-link 자동 resolution UI | `/resolve-links` + confidence slider |

## 10. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-19 | 1차 작성 (M-v0.2.0 PoC + frontend 1차 release, retro-design recovery). §1-§9 9 section + 5 page scope + Path Y dev fixture + 3 layer architecture. |
