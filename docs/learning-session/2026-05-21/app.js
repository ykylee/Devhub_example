// DevHub Example 학습회 자료 — Chart.js 차트 정의
// main HEAD: d730fc6 (2026-05-21)

(function () {
  if (typeof Chart === "undefined") {
    console.error("[learning] Chart.js failed to load — check vendor/chart.umd.min.js");
    return;
  }

  // global theme
  Chart.defaults.color = "#cbd5e1";
  Chart.defaults.borderColor = "#334155";
  Chart.defaults.font.family = '-apple-system, "Segoe UI", "Noto Sans KR", sans-serif';
  Chart.defaults.plugins.legend.labels.boxWidth = 14;

  const palette = {
    accent: "#38bdf8",
    accent2: "#a78bfa",
    ok: "#22c55e",
    warn: "#eab308",
    gap: "#f97316",
    skip: "#64748b",
    pink: "#ec4899",
    cyan: "#06b6d4",
    indigo: "#6366f1",
    emerald: "#10b981",
  };

  // helper for transparent fills
  const alpha = (hex, a) => {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r},${g},${b},${a})`;
  };

  // ============================================================
  // Chart 1 — M0~M7 마일스톤 timeline (horizontal bar / Gantt-style)
  // ============================================================
  const ctx1 = document.getElementById("chart-milestone-timeline");
  if (ctx1) {
    // [start_day, end_day] from 2026-05-07 baseline
    const milestones = [
      { name: "M0 초기 스캐폴드", start: 0, end: 0, summary: "ADR-0001 Hydra+Kratos" },
      { name: "M1 인증/RBAC 기반", start: 1, end: 1, summary: "M1 PR-A..G, ADR-0002" },
      { name: "M2 인증/계정 완성", start: 1, end: 5, summary: "login/logout chain" },
      { name: "M3 User/Org 관리", start: 6, end: 7, summary: "ADR-0008/0009/0010" },
      { name: "M5 DREQ 종합", start: 8, end: 11, summary: "ADR-0012/13/14/17" },
      { name: "M6 Integration", start: 8, end: 11, summary: "ADR-0015/16 + topology" },
      { name: "M7 Onboarding ⚡", start: 14, end: 14, summary: "ADR-0021 + Carve A" },
      { name: "M4 Realtime (planned)", start: 14, end: 30, summary: "v1.1+ — WS/AI v2" },
    ];

    new Chart(ctx1, {
      type: "bar",
      data: {
        labels: milestones.map((m) => m.name),
        datasets: [
          {
            label: "기간 (day, 2026-05-07 기준)",
            data: milestones.map((m) => [m.start, m.end + 1]),
            backgroundColor: milestones.map((m, i) =>
              i === 6 ? alpha(palette.accent2, 0.85) : i === 7 ? alpha(palette.skip, 0.5) : alpha(palette.accent, 0.65)
            ),
            borderColor: milestones.map((m, i) =>
              i === 6 ? palette.accent2 : i === 7 ? palette.skip : palette.accent
            ),
            borderWidth: 1,
            borderRadius: 4,
          },
        ],
      },
      options: {
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: {
            beginAtZero: true,
            title: { display: true, text: "Day offset (D-0 = 2026-05-07)" },
            grid: { color: alpha(palette.skip, 0.2) },
          },
          y: {
            grid: { display: false },
          },
        },
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => {
                const m = milestones[ctx.dataIndex];
                return `${m.summary} (D+${m.start}~D+${m.end})`;
              },
            },
          },
        },
      },
    });
  }

  // ============================================================
  // Chart 2 — v1.0/v1.1/v2 잔여 carve 분포 (stacked bar)
  // ============================================================
  const ctx2 = document.getElementById("chart-carve-distribution");
  if (ctx2) {
    new Chart(ctx2, {
      type: "bar",
      data: {
        labels: ["v1.0 D-25", "v1.1 (M-v1.1)", "v2 P3"],
        datasets: [
          { label: "P0 (즉시 진입)", data: [0, 0, 0], backgroundColor: alpha(palette.gap, 0.85) },
          { label: "P1 (v1.0 안정성)", data: [1, 2, 0], backgroundColor: alpha(palette.warn, 0.85) },
          { label: "P2 (운영 + Onboarding carve)", data: [0, 11, 0], backgroundColor: alpha(palette.accent, 0.85) },
          { label: "P3 (v2 후순위)", data: [0, 1, 10], backgroundColor: alpha(palette.skip, 0.85) },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: { stacked: true, grid: { display: false } },
          y: { stacked: true, beginAtZero: true, title: { display: true, text: "잔여 carve 수" }, grid: { color: alpha(palette.skip, 0.2) } },
        },
        plugins: {
          tooltip: {
            callbacks: {
              afterLabel: (ctx) => {
                if (ctx.label === "v1.1 (M-v1.1)" && ctx.dataset.label.startsWith("P2")) {
                  return "포함: P2-8/9/10/11 Onboarding Carve A(✅)/B/C/D";
                }
                return "";
              },
            },
          },
        },
      },
    });
  }

  // ============================================================
  // Chart 3 — ADR 도메인 분포 (donut)
  // ============================================================
  const ctx3 = document.getElementById("chart-adr-domain");
  if (ctx3) {
    new Chart(ctx3, {
      type: "doughnut",
      data: {
        labels: [
          "인증/계정 (0001/0019/0020/0021)",
          "RBAC (0002/0007/0011)",
          "DREQ (0012/13/14/17)",
          "Integration (0015/16)",
          "조직 모델 (0008/9/10)",
          "정책/인프라 (0003/5/18)",
          "Legacy (0004/6)",
        ],
        datasets: [
          {
            data: [4, 3, 4, 2, 3, 3, 2],
            backgroundColor: [
              palette.accent,
              palette.indigo,
              palette.cyan,
              palette.emerald,
              palette.warn,
              palette.skip,
              palette.pink,
            ],
            borderColor: "#0f172a",
            borderWidth: 2,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: "right" },
          tooltip: {
            callbacks: {
              label: (ctx) => `${ctx.label}: ${ctx.parsed} ADR`,
            },
          },
        },
      },
    });
  }

  // ============================================================
  // Chart 4 — 단계별 ID 발급 수 (horizontal bar)
  // ============================================================
  const ctx4 = document.getElementById("chart-traceability-counts");
  if (ctx4) {
    new Chart(ctx4, {
      type: "bar",
      data: {
        labels: ["REQ-FR (~150)", "REQ-NFR (~55)", "UC (~75)", "ARCH (35)", "API (86)", "RM (~25)", "IMPL (~80)", "UT (~50)", "TC (~50)", "ADR (21)"],
        datasets: [
          {
            label: "ID 발급 수",
            data: [150, 55, 75, 35, 86, 25, 80, 50, 50, 21],
            backgroundColor: [
              palette.accent, palette.accent, palette.indigo, palette.cyan, palette.emerald,
              palette.warn, palette.gap, palette.pink, palette.accent2, palette.skip,
            ].map(c => alpha(c, 0.75)),
            borderColor: [
              palette.accent, palette.accent, palette.indigo, palette.cyan, palette.emerald,
              palette.warn, palette.gap, palette.pink, palette.accent2, palette.skip,
            ],
            borderWidth: 1,
            borderRadius: 4,
          },
        ],
      },
      options: {
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { beginAtZero: true, title: { display: true, text: "ID 누적 수" }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { grid: { display: false } },
        },
      },
    });
  }

  // ============================================================
  // Chart 5 — 2026-05-21 누적 PR (line + scatter)
  // ============================================================
  const ctx5 = document.getElementById("chart-pr-timeline");
  if (ctx5) {
    // PR 시간순 (오늘 작업 누적)
    const prs = [
      { num: 265, h: 7, type: "docs", domain: "Onboarding concept" },
      { num: 266, h: 8, type: "docs", domain: "Onboarding REQ" },
      { num: 267, h: 9, type: "docs", domain: "Onboarding ARCH/API" },
      { num: 269, h: 10, type: "docs", domain: "ADR-0021" },
      { num: 270, h: 11, type: "docs", domain: "codex hotfix #1" },
      { num: 271, h: 12, type: "docs", domain: "Carve plan" },
      { num: 276, h: 12.5, type: "docs", domain: "codex hotfix #2" },
      { num: 278, h: 13, type: "feat", domain: "Carve A backend" },
      { num: 277, h: 13.5, type: "codex", domain: "Deploy refactor" },
    ];

    new Chart(ctx5, {
      type: "line",
      data: {
        datasets: [
          {
            label: "Docs / Plan PR",
            data: prs.filter((p) => p.type === "docs").map((p, i, arr) => ({ x: p.h, y: arr.findIndex(x => x.num === p.num) + 1, pr: p })),
            backgroundColor: alpha(palette.accent2, 0.75),
            borderColor: palette.accent2,
            pointRadius: 8,
            pointHoverRadius: 11,
            showLine: false,
          },
          {
            label: "Backend feat PR",
            data: prs.filter((p) => p.type === "feat").map((p) => ({ x: p.h, y: 8, pr: p })),
            backgroundColor: alpha(palette.ok, 0.85),
            borderColor: palette.ok,
            pointRadius: 12,
            pointHoverRadius: 15,
            pointStyle: "star",
            showLine: false,
          },
          {
            label: "Codex deploy PR",
            data: prs.filter((p) => p.type === "codex").map((p) => ({ x: p.h, y: 9, pr: p })),
            backgroundColor: alpha(palette.gap, 0.85),
            borderColor: palette.gap,
            pointRadius: 10,
            pointHoverRadius: 13,
            pointStyle: "rectRot",
            showLine: false,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: {
            type: "linear",
            min: 6.5,
            max: 14,
            title: { display: true, text: "오늘의 시간 (시각, 24h)" },
            ticks: { stepSize: 1 },
            grid: { color: alpha(palette.skip, 0.2) },
          },
          y: {
            min: 0,
            max: 10,
            title: { display: true, text: "PR 누적 (순서)" },
            grid: { color: alpha(palette.skip, 0.2) },
          },
        },
        plugins: {
          tooltip: {
            callbacks: {
              label: (ctx) => {
                const p = ctx.raw.pr;
                return [`#${p.num} — ${p.domain}`, `${p.h}h`];
              },
            },
          },
        },
      },
    });
  }

  // ============================================================
  // Chart 6 — Migration 누적 (line)
  // ============================================================
  const ctx6 = document.getElementById("chart-migration-cumulative");
  if (ctx6) {
    // domain별로 누적 migration 수
    const data = [
      { day: "M0~M2", count: 11, label: "초기 schema" },
      { day: "M3", count: 18, label: "+app/proj (12~18)" },
      { day: "M5 DREQ", count: 27, label: "+ dreq + tokens + expires" },
      { day: "M6 INT", count: 30, label: "+ integration providers/bindings" },
      { day: "ADR-0020", count: 32, label: "+ event_cursors + audit dedup" },
      { day: "M7 Onboarding", count: 33, label: "+ users.onboarding_state ⚡" },
    ];

    new Chart(ctx6, {
      type: "line",
      data: {
        labels: data.map((d) => d.day),
        datasets: [
          {
            label: "Migration 누적",
            data: data.map((d) => d.count),
            backgroundColor: alpha(palette.accent, 0.3),
            borderColor: palette.accent,
            borderWidth: 3,
            fill: true,
            tension: 0.25,
            pointRadius: 7,
            pointHoverRadius: 10,
            pointBackgroundColor: palette.accent,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: { beginAtZero: true, title: { display: true, text: "Migration 누적 (000XXX)" }, grid: { color: alpha(palette.skip, 0.2) } },
          x: { grid: { display: false } },
        },
        plugins: {
          tooltip: {
            callbacks: {
              afterLabel: (ctx) => data[ctx.dataIndex].label,
            },
          },
        },
      },
    });
  }

  // ============================================================
  // Chart 7 — CI job 평균 시간 (horizontal bar)
  // ============================================================
  const ctx7 = document.getElementById("chart-ci-jobs");
  if (ctx7) {
    new Chart(ctx7, {
      type: "bar",
      data: {
        labels: [
          "Detect Changed Paths",
          "Workflow Lint (actionlint)",
          "Migration Prefix Uniqueness",
          "Backend Unit Tests",
          "Backend Integration Tests",
          "Frontend Unit Tests",
          "E2E shard 1/2",
          "E2E shard 2/2",
        ],
        datasets: [
          {
            label: "평균 시간 (초)",
            data: [8, 12, 6, 20, 55, 30, 200, 240],
            backgroundColor: [
              alpha(palette.skip, 0.7),
              alpha(palette.skip, 0.7),
              alpha(palette.skip, 0.7),
              alpha(palette.accent, 0.7),
              alpha(palette.accent, 0.85),
              alpha(palette.cyan, 0.7),
              alpha(palette.warn, 0.85),
              alpha(palette.warn, 0.85),
            ],
            borderColor: [
              palette.skip, palette.skip, palette.skip,
              palette.accent, palette.accent,
              palette.cyan,
              palette.warn, palette.warn,
            ],
            borderWidth: 1,
            borderRadius: 4,
          },
        ],
      },
      options: {
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { beginAtZero: true, title: { display: true, text: "초 (s)" }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { grid: { display: false } },
        },
      },
    });
  }

  console.log("[learning] charts initialized");
})();
