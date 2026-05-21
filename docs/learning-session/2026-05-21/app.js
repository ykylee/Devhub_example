// DevHub Example 학습회 자료 — Slideshow + Chart.js
// main HEAD: d730fc6 (2026-05-21)

(function () {
  // ─────────────────────────── Slideshow framework ──
  const slidesEl = document.querySelector(".slides");
  const slideEls = Array.from(document.querySelectorAll(".slide"));
  const total = slideEls.length;
  const progressEl = document.querySelector(".progress-bar");
  const counterEl = document.querySelector(".slide-controls .counter");
  const prevBtn = document.querySelector(".slide-controls .prev");
  const nextBtn = document.querySelector(".slide-controls .next");
  const menuToggle = document.querySelector(".slide-menu-toggle");
  const menuEl = document.querySelector(".slide-menu");
  let current = 0;
  let chartsBuilt = new Set();

  function gotoSlide(idx) {
    idx = Math.max(0, Math.min(total - 1, idx));
    current = idx;
    if (slidesEl) slidesEl.style.transform = `translateX(-${idx * 100}vw)`;
    if (progressEl) progressEl.style.width = `${((idx + 1) / total) * 100}%`;
    if (counterEl) counterEl.textContent = `${idx + 1} / ${total}`;
    if (prevBtn) prevBtn.disabled = idx === 0;
    if (nextBtn) nextBtn.disabled = idx === total - 1;
    // update slide-num + menu active state
    slideEls.forEach((s, i) => {
      const numEl = s.querySelector(".slide-num");
      if (numEl) numEl.textContent = `${i + 1} / ${total}`;
    });
    document.querySelectorAll(".slide-menu a").forEach((a, i) => {
      a.classList.toggle("active", i === idx);
    });
    // lazy init chart on visible slide
    initChartsForSlide(idx);
    // hash sync
    if (slideEls[idx] && slideEls[idx].id) {
      history.replaceState(null, "", `#${slideEls[idx].id}`);
    }
  }

  function nextSlide() { gotoSlide(current + 1); }
  function prevSlide() { gotoSlide(current - 1); }

  document.addEventListener("keydown", (e) => {
    if (e.target.tagName === "INPUT" || e.target.tagName === "TEXTAREA") return;
    switch (e.key) {
      case "ArrowRight":
      case "PageDown":
      case " ":
        e.preventDefault();
        nextSlide();
        break;
      case "ArrowLeft":
      case "PageUp":
        e.preventDefault();
        prevSlide();
        break;
      case "Home":
        e.preventDefault();
        gotoSlide(0);
        break;
      case "End":
        e.preventDefault();
        gotoSlide(total - 1);
        break;
      case "Escape":
        if (menuEl) menuEl.classList.remove("open");
        break;
    }
  });

  if (prevBtn) prevBtn.addEventListener("click", prevSlide);
  if (nextBtn) nextBtn.addEventListener("click", nextSlide);

  if (menuToggle) {
    menuToggle.addEventListener("click", (e) => {
      e.stopPropagation();
      if (menuEl) menuEl.classList.toggle("open");
    });
  }
  document.addEventListener("click", (e) => {
    if (menuEl && !menuEl.contains(e.target) && e.target !== menuToggle) {
      menuEl.classList.remove("open");
    }
  });

  // menu link click → goto + close
  document.querySelectorAll(".slide-menu a").forEach((a, i) => {
    a.addEventListener("click", (e) => {
      e.preventDefault();
      gotoSlide(i);
      if (menuEl) menuEl.classList.remove("open");
    });
  });

  // initial hash
  const hash = window.location.hash.replace("#", "");
  if (hash) {
    const idx = slideEls.findIndex((s) => s.id === hash);
    if (idx >= 0) {
      // wait for first paint then jump
      requestAnimationFrame(() => gotoSlide(idx));
    } else {
      gotoSlide(0);
    }
  } else {
    gotoSlide(0);
  }

  // ─────────────────────────── Chart.js setup ───────
  if (typeof Chart === "undefined") {
    console.error("[learning] Chart.js failed to load — check vendor/chart.umd.min.js");
    return;
  }

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
  const alpha = (hex, a) => {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r},${g},${b},${a})`;
  };

  function initChartsForSlide(slideIdx) {
    const slide = slideEls[slideIdx];
    if (!slide) return;
    slide.querySelectorAll("canvas[data-chart]").forEach((canvas) => {
      const key = canvas.id;
      if (chartsBuilt.has(key)) return;
      chartsBuilt.add(key);
      buildChart(key, canvas);
    });
  }

  function buildChart(key, ctx) {
    switch (key) {
      case "chart-milestone-timeline":
        buildMilestoneTimeline(ctx); break;
      case "chart-carve-distribution":
        buildCarveDistribution(ctx); break;
      case "chart-adr-domain":
        buildAdrDomain(ctx); break;
      case "chart-traceability-counts":
        buildTraceabilityCounts(ctx); break;
      case "chart-pr-timeline":
        buildPrTimeline(ctx); break;
      case "chart-migration-cumulative":
        buildMigrationCumulative(ctx); break;
      case "chart-ci-jobs":
        buildCiJobs(ctx); break;
    }
  }

  function buildMilestoneTimeline(ctx) {
    const ms = [
      { name: "M0 초기 스캐폴드", s: 0, e: 0 },
      { name: "M1 인증/RBAC 기반", s: 1, e: 1 },
      { name: "M2 인증/계정 완성", s: 1, e: 5 },
      { name: "M3 User/Org 관리", s: 6, e: 7 },
      { name: "M5 DREQ 종합", s: 8, e: 11 },
      { name: "M6 Integration", s: 8, e: 11 },
      { name: "M7 Onboarding ⚡", s: 14, e: 14 },
      { name: "M4 Realtime (planned)", s: 14, e: 30 },
    ];
    new Chart(ctx, {
      type: "bar",
      data: {
        labels: ms.map(m => m.name),
        datasets: [{
          label: "기간 (day, D-0 = 2026-05-07)",
          data: ms.map(m => [m.s, m.e + 1]),
          backgroundColor: ms.map((m, i) =>
            i === 6 ? alpha(palette.accent2, 0.85) : i === 7 ? alpha(palette.skip, 0.5) : alpha(palette.accent, 0.65)),
          borderColor: ms.map((m, i) =>
            i === 6 ? palette.accent2 : i === 7 ? palette.skip : palette.accent),
          borderWidth: 1, borderRadius: 4,
        }],
      },
      options: {
        indexAxis: "y", responsive: true, maintainAspectRatio: false,
        scales: {
          x: { beginAtZero: true, title: { display: true, text: "Day offset" }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { grid: { display: false } },
        },
        plugins: { legend: { display: false } },
      },
    });
  }

  function buildCarveDistribution(ctx) {
    new Chart(ctx, {
      type: "bar",
      data: {
        labels: ["v1.0 D-25", "v1.1 (M-v1.1)", "v2 P3"],
        datasets: [
          { label: "P1 (안정성)", data: [1, 2, 0], backgroundColor: alpha(palette.warn, 0.85) },
          { label: "P2 (운영 + Onboarding)", data: [0, 11, 0], backgroundColor: alpha(palette.accent, 0.85) },
          { label: "P3 (후순위)", data: [0, 1, 10], backgroundColor: alpha(palette.skip, 0.85) },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        scales: {
          x: { stacked: true, grid: { display: false } },
          y: { stacked: true, beginAtZero: true, title: { display: true, text: "carve 수" }, grid: { color: alpha(palette.skip, 0.2) } },
        },
      },
    });
  }

  function buildAdrDomain(ctx) {
    new Chart(ctx, {
      type: "doughnut",
      data: {
        labels: ["인증/계정 (4)", "RBAC (3)", "DREQ (4)", "Integration (2)", "조직 모델 (3)", "정책/인프라 (3)", "Legacy (2)"],
        datasets: [{
          data: [4, 3, 4, 2, 3, 3, 2],
          backgroundColor: [palette.accent, palette.indigo, palette.cyan, palette.emerald, palette.warn, palette.skip, palette.pink],
          borderColor: "#0f172a", borderWidth: 2,
        }],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { position: "right" } },
      },
    });
  }

  function buildTraceabilityCounts(ctx) {
    new Chart(ctx, {
      type: "bar",
      data: {
        labels: ["REQ-FR", "REQ-NFR", "UC", "ARCH", "API", "RM", "IMPL", "UT", "TC", "ADR"],
        datasets: [{
          label: "ID 발급 수",
          data: [150, 55, 75, 35, 86, 25, 80, 50, 50, 21],
          backgroundColor: [palette.accent, palette.accent, palette.indigo, palette.cyan, palette.emerald,
                            palette.warn, palette.gap, palette.pink, palette.accent2, palette.skip].map(c => alpha(c, 0.75)),
          borderColor: [palette.accent, palette.accent, palette.indigo, palette.cyan, palette.emerald,
                        palette.warn, palette.gap, palette.pink, palette.accent2, palette.skip],
          borderWidth: 1, borderRadius: 4,
        }],
      },
      options: {
        indexAxis: "y", responsive: true, maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { beginAtZero: true, title: { display: true, text: "ID 누적 수" }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { grid: { display: false } },
        },
      },
    });
  }

  function buildPrTimeline(ctx) {
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
    new Chart(ctx, {
      type: "scatter",
      data: {
        datasets: [
          {
            label: "Docs / Plan PR",
            data: prs.filter(p => p.type === "docs").map((p, i) => ({ x: p.h, y: i + 1, pr: p })),
            backgroundColor: alpha(palette.accent2, 0.75),
            borderColor: palette.accent2,
            pointRadius: 9, pointHoverRadius: 12,
          },
          {
            label: "Backend feat PR",
            data: prs.filter(p => p.type === "feat").map(p => ({ x: p.h, y: 8, pr: p })),
            backgroundColor: alpha(palette.ok, 0.85),
            borderColor: palette.ok,
            pointRadius: 13, pointHoverRadius: 16, pointStyle: "star",
          },
          {
            label: "Codex deploy PR",
            data: prs.filter(p => p.type === "codex").map(p => ({ x: p.h, y: 9, pr: p })),
            backgroundColor: alpha(palette.gap, 0.85),
            borderColor: palette.gap,
            pointRadius: 11, pointHoverRadius: 14, pointStyle: "rectRot",
          },
        ],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        scales: {
          x: { type: "linear", min: 6.5, max: 14, title: { display: true, text: "시각 (24h)" }, ticks: { stepSize: 1 }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { min: 0, max: 10, title: { display: true, text: "PR 순서" }, grid: { color: alpha(palette.skip, 0.2) } },
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

  function buildMigrationCumulative(ctx) {
    const data = [
      { day: "M0~M2", count: 11 },
      { day: "M3", count: 18 },
      { day: "M5 DREQ", count: 27 },
      { day: "M6 INT", count: 30 },
      { day: "ADR-0020", count: 32 },
      { day: "M7 Onboarding", count: 33 },
    ];
    new Chart(ctx, {
      type: "line",
      data: {
        labels: data.map(d => d.day),
        datasets: [{
          label: "Migration 누적",
          data: data.map(d => d.count),
          backgroundColor: alpha(palette.accent, 0.3),
          borderColor: palette.accent, borderWidth: 3,
          fill: true, tension: 0.25, pointRadius: 7, pointHoverRadius: 10,
          pointBackgroundColor: palette.accent,
        }],
      },
      options: {
        responsive: true, maintainAspectRatio: false,
        scales: {
          y: { beginAtZero: true, title: { display: true, text: "누적 수" }, grid: { color: alpha(palette.skip, 0.2) } },
          x: { grid: { display: false } },
        },
      },
    });
  }

  function buildCiJobs(ctx) {
    new Chart(ctx, {
      type: "bar",
      data: {
        labels: ["Detect Path", "Workflow Lint", "Migration Unique", "Backend Unit", "Backend Integration", "Frontend Unit", "E2E shard 1/2", "E2E shard 2/2"],
        datasets: [{
          label: "평균 시간 (초)",
          data: [8, 12, 6, 20, 55, 30, 200, 240],
          backgroundColor: [alpha(palette.skip, 0.7), alpha(palette.skip, 0.7), alpha(palette.skip, 0.7),
                            alpha(palette.accent, 0.7), alpha(palette.accent, 0.85),
                            alpha(palette.cyan, 0.7), alpha(palette.warn, 0.85), alpha(palette.warn, 0.85)],
          borderColor: [palette.skip, palette.skip, palette.skip, palette.accent, palette.accent, palette.cyan, palette.warn, palette.warn],
          borderWidth: 1, borderRadius: 4,
        }],
      },
      options: {
        indexAxis: "y", responsive: true, maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { beginAtZero: true, title: { display: true, text: "초 (s)" }, grid: { color: alpha(palette.skip, 0.2) } },
          y: { grid: { display: false } },
        },
      },
    });
  }

  console.log(`[learning] slideshow ready (${total} slides), charts lazy-initialized`);
})();
