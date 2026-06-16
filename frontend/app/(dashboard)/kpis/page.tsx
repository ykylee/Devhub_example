"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Plus, 
  Trash2, 
  ArrowLeft, 
  ArrowRight, 
  HelpCircle, 
  Percent, 
  Hash, 
  Star,
  Settings,
  Terminal,
  Terminal as CodeIcon
} from "lucide-react";
import { DashboardHeader } from "@/shared/ui-foundation/components/DashboardHeader";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { DomainPicker, type DomainEntity } from "@/shared/ui-foundation/components/DomainPicker";
import { repositoryService } from "@/domain/repository-integration/service/repository.service";

interface KPIItem {
  id: string;
  name: string;
  description: string;
  type: "percentage" | "count" | "score" | "python";
  numerator?: keyof typeof DATA_SOURCES;
  denominator?: keyof typeof DATA_SOURCES;
  singleSource?: keyof typeof DATA_SOURCES;
  pythonCode?: string;
}

const DATA_SOURCES = {
  total_platforms: { label: "전체 플랫폼 수", value: 12 },
  total_projects: { label: "전체 프로젝트 수", value: 25 },
  total_dev_requests: { label: "전체 DREQ 수", value: 34 },
  total_repositories: { label: "전체 리포지토리 수", value: 8 },
  total_builds: { label: "전체 빌드 수", value: 150 },
  active_projects: { label: "활성 프로젝트 수", value: 18 },
  passed_quality_gates: { label: "품질 게이트 통과 수", value: 10 },
  failed_quality_gates: { label: "품질 게이트 실패 수", value: 2 },
  open_pull_requests: { label: "오픈된 PR 수", value: 42 },
  successful_builds: { label: "성공한 빌드 수", value: 125 },
  avg_quality_score: { label: "평균 품질 점수", value: 4.2 },
};

const SEED_KPIS: KPIItem[] = [
  {
    id: "kpi-1",
    name: "품질 게이트 통과율",
    description: "전체 플랫폼 대비 정밀 품질 게이트를 만족하는 비율",
    type: "percentage",
    numerator: "passed_quality_gates",
    denominator: "total_platforms",
  },
  {
    id: "kpi-2",
    name: "프로젝트 활성화율",
    description: "전체 등록 프로젝트 중 활성화되어 실행 중인 비율",
    type: "percentage",
    numerator: "active_projects",
    denominator: "total_projects",
  },
  {
    id: "kpi-3",
    name: "활성 풀 리퀘스트 건수",
    description: "개발자 간의 현재 열려 있는 협업 리뷰 단위 수",
    type: "count",
    singleSource: "open_pull_requests",
  },
  {
    id: "kpi-4",
    name: "시스템 평균 품질 점수",
    description: "각 소스 코드 리포지토리의 종합 품질 평가 평균치",
    type: "score",
    singleSource: "avg_quality_score",
  },
  {
    id: "kpi-5",
    name: "빌드 가중 품질 효율 점수 (Python)",
    description: "파이썬 스크립트를 통해 성공 빌드 비율과 평균 품질 점수를 조합한 지표를 계산합니다.",
    type: "python",
    pythonCode: `# 빌드 품질 효율 계산 스크립트
score = avg_quality_score
success = successful_builds
total_b = total_builds

# 성공 빌드 비율 계산
build_ratio = success / total_b if total_b > 0 else 0

# 품질 점수에 성공 빌드율을 곱해 최종 효율 점수 도출
return score * build_ratio
`
  }
];

export default function KPIDashboardPage() {
  const [kpis, setKpis] = useState<KPIItem[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  
  // Form state
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [type, setType] = useState<"percentage" | "count" | "score" | "python">("percentage");
  const [numerator, setNumerator] = useState<keyof typeof DATA_SOURCES>("passed_quality_gates");
  const [denominator, setDenominator] = useState<keyof typeof DATA_SOURCES>("total_platforms");
  const [singleSource, setSingleSource] = useState<keyof typeof DATA_SOURCES>("open_pull_requests");
  const [pythonCode, setPythonCode] = useState("");

  // Sprint D — DomainPicker entity fetch (kpi-tests-per-domain-scope.md §6.4).
  // Repository scope 만 실제 fetch (Sprint A 활성화). Platform/Project 는 Sprint
  // B/C 와 함께 fetch 추가. 실패 시 picker 가 에러 표시.
  const [repositories, setRepositories] = useState<DomainEntity[]>([]);
  const [reposLoading, setReposLoading] = useState(true);
  const [reposError, setReposError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setReposLoading(true);
    setReposError(null);
    repositoryService.listRepositories()
      .then((repos) => {
        if (cancelled) return;
        setRepositories(
          repos.map((r) => ({ id: String(r.id), name: r.full_name, description: r.clone_url })),
        );
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setReposError(err instanceof Error ? err.message : "Failed to load repositories");
      })
      .finally(() => {
        if (cancelled) return;
        setReposLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const stored = localStorage.getItem("devhub-custom-kpis");
    if (stored) {
      try {
        setKpis(JSON.parse(stored) as KPIItem[]);
      } catch (e) {
        setKpis(SEED_KPIS);
      }
    } else {
      setKpis(SEED_KPIS);
    }
  }, []);

  const saveKPIs = (updated: KPIItem[]) => {
    setKpis(updated);
    localStorage.setItem("devhub-custom-kpis", JSON.stringify(updated));
  };

  const handleAddKPI = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    const newKpi: KPIItem = {
      id: "kpi-" + Date.now(),
      name: name.trim(),
      description: description.trim(),
      type,
      ...(type === "percentage" ? { numerator, denominator } : 
         type === "python" ? { pythonCode } : { singleSource }),
    };

    const updated = [...kpis, newKpi];
    saveKPIs(updated);
    setIsModalOpen(false);
    
    // reset form
    setName("");
    setDescription("");
    setType("percentage");
    setPythonCode("");
  };

  const handleDeleteKPI = (id: string) => {
    const updated = kpis.filter((kpi) => kpi.id !== id);
    saveKPIs(updated);
  };

  const moveKPI = (index: number, direction: "left" | "right") => {
    const nextIndex = direction === "left" ? index - 1 : index + 1;
    if (nextIndex < 0 || nextIndex >= kpis.length) return;

    const updated = [...kpis];
    const temp = updated[index];
    updated[index] = updated[nextIndex];
    updated[nextIndex] = temp;
    saveKPIs(updated);
  };

  // 가상 Python 스크립트 실행기
  const executePythonMetric = (code: string): number => {
    try {
      // 1. 변수 바인딩 컨텍스트 구성
      const ctx = {
        total_platforms: DATA_SOURCES.total_platforms.value,
        total_projects: DATA_SOURCES.total_projects.value,
        total_dev_requests: DATA_SOURCES.total_dev_requests.value,
        total_repositories: DATA_SOURCES.total_repositories.value,
        total_builds: DATA_SOURCES.total_builds.value,
        active_projects: DATA_SOURCES.active_projects.value,
        passed_quality_gates: DATA_SOURCES.passed_quality_gates.value,
        failed_quality_gates: DATA_SOURCES.failed_quality_gates.value,
        open_pull_requests: DATA_SOURCES.open_pull_requests.value,
        successful_builds: DATA_SOURCES.successful_builds.value,
        avg_quality_score: DATA_SOURCES.avg_quality_score.value,
      };

      // 2. Python 주석(#) 제거
      let jsCode = code.replace(/#.*/g, "");

      // 3. 변수 치환
      Object.entries(ctx).forEach(([key, val]) => {
        const regex = new RegExp(`\\b${key}\\b`, "g");
        jsCode = jsCode.replace(regex, String(val));
      });

      // 3.5 Python 삼항 연산자 (A if COND else B) -> JS 삼항 연산자 (COND ? A : B) 변환
      const lines = jsCode.split("\n");
      const processedLines = lines.map(line => {
        let text = line;
        let prefix = "";
        let target = text;
        if (text.includes("=")) {
          const idx = text.indexOf("=");
          prefix = text.substring(0, idx + 1);
          target = text.substring(idx + 1);
        }

        const ternaryRegex = /(.+?)\s+if\s+(.+?)\s+else\s+(.+)/;
        const match = target.match(ternaryRegex);
        if (match) {
          const [, expr1, cond, expr2] = match;
          let returnPrefix = "";
          let cleanExpr1 = expr1;
          if (expr1.trim().startsWith("return ")) {
            returnPrefix = "return ";
            cleanExpr1 = expr1.replace("return ", "");
          }
          return `${prefix}${returnPrefix}(${cond.trim()}) ? (${cleanExpr1.trim()}) : (${expr2.trim()})`;
        }
        return line;
      });
      jsCode = processedLines.join("\n");

      // 4. 간단한 dynamic evaluate
      if (jsCode.includes("return")) {
        const fnBody = jsCode.replace(/def .*\n/g, ""); // def 선언이 있는 경우 제거
        const tempFn = new Function(fnBody);
        return Number(tempFn());
      } else {
        const lines = jsCode.trim().split("\n");
        const lastLine = lines[lines.length - 1];
        if (lastLine.includes("=")) {
          const parts = lastLine.split("=");
          const expr = parts[parts.length - 1];
          const tempFn = new Function(`return (${expr});`);
          return Number(tempFn());
        } else {
          const tempFn = new Function(`return (${jsCode});`);
          return Number(tempFn());
        }
      }
    } catch (e) {
      console.warn("Python execution simulation failed:", e);
      return NaN;
    }
  };

  const calculateKPIValue = (kpi: KPIItem) => {
    if (kpi.type === "percentage" && kpi.numerator && kpi.denominator) {
      const numVal = DATA_SOURCES[kpi.numerator]?.value ?? 0;
      const denVal = DATA_SOURCES[kpi.denominator]?.value ?? 1;
      return denVal === 0 ? 0 : (numVal / denVal) * 100;
    }
    if (kpi.type === "count" && kpi.singleSource) {
      return DATA_SOURCES[kpi.singleSource]?.value ?? 0;
    }
    if (kpi.type === "score" && kpi.singleSource) {
      return DATA_SOURCES[kpi.singleSource]?.value ?? 0;
    }
    if (kpi.type === "python" && kpi.pythonCode) {
      return executePythonMetric(kpi.pythonCode);
    }
    return 0;
  };

  return (
    <div className="space-y-10 pb-20 px-4 md:px-8">
      {/* Sprint D — kpi-tests-per-domain-scope.md §6.4 Domain picker. Repository
          scope 만 실제 entity list fetch (Sprint A 활성화). Platform/Project 는
          Sprint B/C 와 함께 fetch 추가. */}
      <DomainPicker
        defaultScope="repository"
        repositories={repositories}
        loading={reposLoading}
        error={reposError}
      />

      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
        <DashboardHeader
          titlePrefix="KPI"
          titleGradient="Performance Dashboard"
          subtitle="실시간 품질 및 생산성 핵심 성과 지표(KPI)를 정의하고 자유롭게 배치합니다."
        />
        <button 
          onClick={() => {
            setType("percentage");
            setIsModalOpen(true);
          }}
          className="flex items-center gap-2 px-5 py-3 rounded-2xl bg-gradient-to-r from-primary to-accent hover:opacity-90 text-primary-foreground font-bold text-sm shadow-lg shadow-primary/20 transition-all self-start md:self-auto shrink-0"
        >
          <Plus className="w-4 h-4" /> KPI 지표 추가
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {kpis.map((kpi, index) => {
          const value = calculateKPIValue(kpi);
          
          return (
            <motion.div
              layout
              key={kpi.id}
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.3 }}
              className="glass-card p-8 flex flex-col justify-between min-h-[240px] relative group border border-white/10 dark:border-white/5"
            >
              {/* Card Controls */}
              <div className="absolute top-6 right-6 flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <button 
                  type="button"
                  onClick={() => moveKPI(index, "left")} 
                  disabled={index === 0}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-muted-foreground hover:text-foreground disabled:opacity-30 transition-all"
                  title="Move Left"
                >
                  <ArrowLeft className="w-3.5 h-3.5" />
                </button>
                <button 
                  type="button"
                  onClick={() => moveKPI(index, "right")} 
                  disabled={index === kpis.length - 1}
                  className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-muted-foreground hover:text-foreground disabled:opacity-30 transition-all"
                  title="Move Right"
                >
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
                <button 
                  type="button"
                  onClick={() => handleDeleteKPI(kpi.id)}
                  className="p-2 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 transition-all"
                  title="Delete KPI"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>

              <div>
                <div className="flex items-center gap-2 mb-3">
                  {kpi.type === "percentage" && <Percent className="w-4 h-4 text-emerald-500" />}
                  {kpi.type === "count" && <Hash className="w-4 h-4 text-cyan-500" />}
                  {kpi.type === "score" && <Star className="w-4 h-4 text-amber-500" />}
                  {kpi.type === "python" && <Terminal className="w-4 h-4 text-primary" />}
                  <span className="text-[10px] font-black uppercase tracking-wider text-muted-foreground">
                    {kpi.type === "percentage" ? "Percentage" : 
                     kpi.type === "count" ? "Count" : 
                     kpi.type === "score" ? "Score" : "Python Script"}
                  </span>
                </div>
                <h3 className="text-xl font-bold text-foreground mb-2 pr-24">{kpi.name}</h3>
                <p className="text-xs text-muted-foreground line-clamp-2 pr-6 leading-relaxed mb-6">{kpi.description}</p>
              </div>

              <div className="flex items-end justify-between border-t border-white/5 pt-4">
                <div className="text-xs text-muted-foreground flex flex-col gap-0.5 max-w-[65%]">
                  {kpi.type === "percentage" && kpi.numerator && kpi.denominator && (
                    <>
                      <span className="truncate">분자: {DATA_SOURCES[kpi.numerator]?.label} ({DATA_SOURCES[kpi.numerator]?.value})</span>
                      <span className="truncate">분모: {DATA_SOURCES[kpi.denominator]?.label} ({DATA_SOURCES[kpi.denominator]?.value})</span>
                    </>
                  )}
                  {kpi.type !== "percentage" && kpi.type !== "python" && kpi.singleSource && (
                    <span className="truncate">연동원천: {DATA_SOURCES[kpi.singleSource]?.label}</span>
                  )}
                  {kpi.type === "python" && (
                    <span className="font-mono text-[9px] text-zinc-400 bg-white/5 p-1.5 rounded border border-white/5 truncate max-h-16 overflow-y-auto block whitespace-pre">
                      {kpi.pythonCode}
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-4 shrink-0">
                  {kpi.type === "percentage" && (
                    <div className="relative w-16 h-16 shrink-0">
                      <svg className="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
                        <path
                          className="text-white/10 dark:text-white/5 stroke-current"
                          strokeWidth="3"
                          fill="none"
                          d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        />
                        <motion.path
                          initial={{ strokeDasharray: "0, 100" }}
                          animate={{ strokeDasharray: `${value}, 100` }}
                          transition={{ duration: 0.8, ease: "easeOut" }}
                          className="text-emerald-500 stroke-current"
                          strokeWidth="3.2"
                          strokeLinecap="round"
                          fill="none"
                          d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        />
                      </svg>
                      <div className="absolute inset-0 flex items-center justify-center">
                        <span className="text-[10px] font-mono font-bold text-foreground">{isNaN(value) ? "ERR" : `${value.toFixed(0)}%`}</span>
                      </div>
                    </div>
                  )}

                  {kpi.type === "count" && (
                    <div className="text-right">
                      <span className="text-4xl font-black font-mono text-cyan-500">{value}</span>
                      <span className="text-xs text-muted-foreground ml-1">건</span>
                    </div>
                  )}

                  {kpi.type === "score" && (
                    <div className="text-right">
                      <span className="text-4xl font-black font-mono text-amber-500">{value.toFixed(1)}</span>
                      <span className="text-xs text-muted-foreground ml-1">/ 5.0</span>
                    </div>
                  )}

                  {kpi.type === "python" && (
                    <div className="text-right">
                      <span className="text-4xl font-black font-mono text-primary">
                        {isNaN(value) ? "ERR" : value.toFixed(1)}
                      </span>
                      <span className="text-[10px] font-bold text-muted-foreground block uppercase tracking-wider mt-0.5">Score</span>
                    </div>
                  )}
                </div>
              </div>
            </motion.div>
          );
        })}

        {kpis.length === 0 && (
          <div className="col-span-full py-20 text-center glass-card border border-dashed border-white/10">
            <HelpCircle className="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-50" />
            <p className="text-sm font-bold text-muted-foreground">정의된 KPI가 없습니다.</p>
            <p className="text-xs text-muted-foreground/60 mt-1">상단의 "KPI 지표 추가" 버튼을 통해 새로운 KPI를 추가해보세요.</p>
          </div>
        )}
      </div>

      {/* KPI Add Modal */}
      <AnimatePresence>
        {isModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 bg-black/60 backdrop-blur-md"
              onClick={() => setIsModalOpen(false)}
            />
            <motion.div 
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              className="relative w-full max-w-lg p-8 rounded-3xl glass border border-white/10 bg-card dark:bg-zinc-900/90 shadow-2xl z-10 space-y-6 max-h-[90vh] overflow-y-auto"
            >
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-black text-foreground flex items-center gap-2">
                  <Settings className="w-5 h-5 text-primary" /> KPI 지표 추가 정의
                </h3>
                <button 
                  onClick={() => setIsModalOpen(false)}
                  className="p-2 rounded-xl hover:bg-white/10 text-muted-foreground transition-all"
                >
                  ✕
                </button>
              </div>

              <form onSubmit={handleAddKPI} className="space-y-4">
                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">KPI 이름</label>
                  <input 
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="예: 품질 통과 지표"
                    required
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                  />
                </div>

                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">KPI 설명</label>
                  <textarea 
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="KPI에 대한 상세 목적과 역할을 기술합니다."
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all h-16 resize-none"
                  />
                </div>

                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-2">지표 기준 유형</label>
                  <div className="grid grid-cols-4 gap-2">
                    {[
                      { key: "percentage", label: "백분율", icon: Percent },
                      { key: "count", label: "건수", icon: Hash },
                      { key: "score", label: "점수", icon: Star },
                      { key: "python", label: "Python", icon: Terminal }
                    ].map((item) => (
                      <button
                        type="button"
                        key={item.key}
                        onClick={() => setType(item.key as any)}
                        className={`p-2.5 rounded-xl border flex flex-col items-center gap-1.5 text-[10px] font-bold transition-all ${
                          type === item.key 
                            ? "bg-primary/20 border-primary text-primary" 
                            : "bg-white/5 border-white/10 text-muted-foreground hover:bg-white/10"
                        }`}
                      >
                        <item.icon className="w-3.5 h-3.5" />
                        {item.label}
                      </button>
                    ))}
                  </div>
                </div>

                {type === "percentage" && (
                  <div className="grid grid-cols-2 gap-4 p-4 rounded-2xl bg-white/5 border border-white/5">
                    <div>
                      <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1.5">분자 (Numerator)</label>
                      <select
                        value={numerator}
                        onChange={(e) => setNumerator(e.target.value as keyof typeof DATA_SOURCES)}
                        className="w-full px-3 py-2.5 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary transition-all"
                      >
                        {Object.entries(DATA_SOURCES).map(([key, item]) => (
                          <option key={key} value={key}>{item.label} ({item.value})</option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1.5">분모 (Denominator)</label>
                      <select
                        value={denominator}
                        onChange={(e) => setDenominator(e.target.value as keyof typeof DATA_SOURCES)}
                        className="w-full px-3 py-2.5 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary transition-all"
                      >
                        {Object.entries(DATA_SOURCES).map(([key, item]) => (
                          <option key={key} value={key}>{item.label} ({item.value})</option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {type !== "percentage" && type !== "python" && (
                  <div className="p-4 rounded-2xl bg-white/5 border border-white/5">
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1.5">연동 데이터 원천</label>
                    <select
                      value={singleSource}
                      onChange={(e) => setSingleSource(e.target.value as keyof typeof DATA_SOURCES)}
                      className="w-full px-3 py-2.5 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary transition-all"
                    >
                      {Object.entries(DATA_SOURCES).map(([key, item]) => (
                        <option key={key} value={key}>{item.label} ({item.value})</option>
                      ))}
                    </select>
                  </div>
                )}

                {type === "python" && (
                  <div className="space-y-4">
                    <div>
                      <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">Python 메트릭 스크립트</label>
                      <p className="text-[10px] text-muted-foreground mb-2 leading-relaxed">
                        하단의 사용 가능한 시스템 변수들을 자유롭게 활용하여 <code>return</code> 값을 계산하는 파이썬 코드를 작성하세요.
                      </p>
                      <textarea
                        value={pythonCode}
                        onChange={(e) => setPythonCode(e.target.value)}
                        placeholder={`# 예시 스크립트\npassed = passed_quality_gates\ntotal = total_platforms\nreturn (passed / total) * 100 if total > 0 else 0`}
                        className="w-full px-4 py-3 rounded-xl border border-white/10 bg-black/30 text-emerald-400 font-mono text-xs focus:outline-none focus:border-primary transition-all h-40 resize-none"
                      />
                    </div>
                    
                    <div>
                      <p className="text-[9px] font-bold text-muted-foreground uppercase mb-1.5 tracking-wider">사용 가능 시스템 변수 (클릭하여 스크립트에 삽입)</p>
                      <div className="flex flex-wrap gap-1.5 max-h-24 overflow-y-auto p-2 rounded-xl bg-white/5 border border-white/5">
                        {Object.keys(DATA_SOURCES).map((varName) => (
                          <button
                            type="button"
                            key={varName}
                            onClick={() => setPythonCode(prev => prev + (prev.endsWith(" ") || prev === "" ? "" : " ") + varName)}
                            className="px-2 py-1 rounded bg-zinc-800 hover:bg-zinc-700 border border-zinc-700 text-[9px] font-mono text-zinc-300 transition-all"
                          >
                            {varName}
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
                
                <div className="flex justify-end gap-3 pt-4 border-t border-white/5">
                  <button 
                    type="button" 
                    onClick={() => setIsModalOpen(false)}
                    className="px-5 py-2.5 rounded-xl border border-white/10 text-sm text-foreground hover:bg-white/5 transition-all"
                  >
                    취소
                  </button>
                  <button 
                    type="submit" 
                    className="px-5 py-2.5 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-all"
                  >
                    KPI 지표 추가 🚀
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
