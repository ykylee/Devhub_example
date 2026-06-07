"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Plus, 
  Search, 
  Calendar, 
  ListTodo, 
  PlayCircle, 
  CheckCircle2, 
  XCircle, 
  AlertCircle, 
  HelpCircle,
  GitPullRequest,
  Check,
  ChevronRight,
  TrendingUp
} from "lucide-react";
import { DashboardHeader } from "@/shared/ui-foundation/components/DashboardHeader";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip } from "recharts";

interface TestSchedule {
  id: string;
  title: string;
  stage: string;
  startDate: string;
  endDate: string;
  progress: number;
}

interface TestCase {
  id: string;
  title: string;
  component: string;
  type: "E2E" | "Unit" | "Integration";
  priority: "High" | "Medium" | "Low";
  status: "Pass" | "Fail" | "Blocked" | "Untested";
}

interface TestCycle {
  id: string;
  name: string;
  targetVersion: string;
  results: Record<string, TestCase["status"]>; // testcase.id -> status
}

const SEED_SCHEDULES: TestSchedule[] = [
  { id: "sched-1", title: "v1.0 Regression 테스트", stage: "배포 회귀 검증", startDate: "2026-06-01", endDate: "2026-06-05", progress: 100 },
  { id: "sched-2", title: "Staging 환경 E2E 시나리오 테스트", stage: "통합 보안 검증", startDate: "2026-06-08", endDate: "2026-06-12", progress: 40 },
  { id: "sched-3", title: "보안 취약점 및 부하 테스트", stage: "인프라 진단", startDate: "2026-06-15", endDate: "2026-06-19", progress: 0 },
];

const SEED_TESTCASES: TestCase[] = [
  { id: "TC-001", title: "Keycloak 인증 로그인 흐름 검증", component: "Auth", type: "E2E", priority: "High", status: "Pass" },
  { id: "TC-002", title: "DREQ 토큰 만료 정책 백엔드 정합성", component: "DREQ", type: "Unit", priority: "High", status: "Pass" },
  { id: "TC-003", title: "HomeLab Puller 수집 주기 스케줄링", component: "Integration", type: "Integration", priority: "Medium", status: "Pass" },
  { id: "TC-004", title: "조직도 드래그 앤 드롭 위치 저장 영속성", component: "OrgChart", type: "E2E", priority: "Medium", status: "Fail" },
  { id: "TC-005", title: "역방향 프록시 Same-Origin 쿠키 전달 가드", component: "Proxy", type: "Integration", priority: "High", status: "Blocked" },
  { id: "TC-006", title: "사용자 프로필 패스워드 변경 검증", component: "Account", type: "Unit", priority: "Low", status: "Untested" },
  { id: "TC-007", title: "품질 스코어 7일치 그래프 추세 계산 모듈", component: "Platform", type: "Unit", priority: "High", status: "Untested" },
];

const SEED_CYCLE: TestCycle = {
  id: "cycle-1",
  name: "v1.0 Staging Release Cycle 1",
  targetVersion: "v1.0.0",
  results: {
    "TC-001": "Pass",
    "TC-002": "Pass",
    "TC-003": "Pass",
    "TC-004": "Fail",
    "TC-005": "Blocked",
    "TC-006": "Untested",
    "TC-007": "Untested",
  }
};

const PRIORITY_BADGES = {
  High: "danger",
  Medium: "warning",
  Low: "primary",
} as const;

const STATUS_ICONS = {
  Pass: { icon: CheckCircle2, color: "text-emerald-500", label: "Pass" },
  Fail: { icon: XCircle, color: "text-rose-500", label: "Fail" },
  Blocked: { icon: AlertCircle, color: "text-amber-500", label: "Blocked" },
  Untested: { icon: HelpCircle, color: "text-muted-foreground", label: "Untested" },
};

const CHART_COLORS = {
  Pass: "var(--primary)", // emerald/primary
  Fail: "#f43f5e", // rose-500
  Blocked: "#f59e0b", // amber-500
  Untested: "#71717a", // zinc-500
};

export default function TestManagementPage() {
  const [activeTab, setActiveTab] = useState<"schedule" | "cases" | "cycle">("cycle");
  
  // Storage states
  const [schedules, setSchedules] = useState<TestSchedule[]>([]);
  const [testCases, setTestCases] = useState<TestCase[]>([]);
  const [cycle, setCycle] = useState<TestCycle | null>(null);

  // Filter & Search
  const [searchQuery, setSearchQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [priorityFilter, setPriorityFilter] = useState("all");

  // Modals state
  const [isScheduleModalOpen, setIsScheduleModalOpen] = useState(false);
  const [isCaseModalOpen, setIsCaseModalOpen] = useState(false);

  // Form states
  const [newSchedTitle, setNewSchedTitle] = useState("");
  const [newSchedStage, setNewSchedStage] = useState("");
  const [newSchedStart, setNewSchedStart] = useState("");
  const [newSchedEnd, setNewSchedEnd] = useState("");

  const [newCaseTitle, setNewCaseTitle] = useState("");
  const [newCaseComp, setNewCaseComp] = useState("");
  const [newCaseType, setNewCaseType] = useState<TestCase["type"]>("E2E");
  const [newCasePriority, setNewCasePriority] = useState<TestCase["priority"]>("High");

  useEffect(() => {
    // Load schedules
    const storedSched = localStorage.getItem("devhub-test-schedules");
    if (storedSched) setSchedules(JSON.parse(storedSched));
    else setSchedules(SEED_SCHEDULES);

    // Load cases
    const storedCases = localStorage.getItem("devhub-test-cases");
    let currentCases = SEED_TESTCASES;
    if (storedCases) {
      currentCases = JSON.parse(storedCases);
      setTestCases(currentCases);
    } else {
      setTestCases(SEED_TESTCASES);
    }

    // Load cycle
    const storedCycle = localStorage.getItem("devhub-test-cycle");
    if (storedCycle) {
      setCycle(JSON.parse(storedCycle));
    } else {
      // Seed cycle mapping
      const initialResults: Record<string, TestCase["status"]> = {};
      currentCases.forEach((tc) => {
        initialResults[tc.id] = tc.status;
      });
      setCycle({ ...SEED_CYCLE, results: initialResults });
    }
  }, []);

  const saveSchedules = (updated: TestSchedule[]) => {
    setSchedules(updated);
    localStorage.setItem("devhub-test-schedules", JSON.stringify(updated));
  };

  const saveCases = (updated: TestCase[]) => {
    setTestCases(updated);
    localStorage.setItem("devhub-test-cases", JSON.stringify(updated));

    // Update cycle mapping together
    if (cycle) {
      const updatedResults = { ...cycle.results };
      updated.forEach((tc) => {
        if (updatedResults[tc.id] === undefined) {
          updatedResults[tc.id] = "Untested";
        }
      });
      const updatedCycle = { ...cycle, results: updatedResults };
      setCycle(updatedCycle);
      localStorage.setItem("devhub-test-cycle", JSON.stringify(updatedCycle));
    }
  };

  const handleAddSchedule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSchedTitle.trim()) return;

    const newSched: TestSchedule = {
      id: "sched-" + Date.now(),
      title: newSchedTitle.trim(),
      stage: newSchedStage.trim() || "일반 테스트",
      startDate: newSchedStart || new Date().toISOString().split("T")[0],
      endDate: newSchedEnd || new Date().toISOString().split("T")[0],
      progress: 0,
    };

    saveSchedules([...schedules, newSched]);
    setIsScheduleModalOpen(false);
    setNewSchedTitle("");
    setNewSchedStage("");
    setNewSchedStart("");
    setNewSchedEnd("");
  };

  const handleAddCase = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCaseTitle.trim()) return;

    const newId = "TC-" + String(testCases.length + 1).padStart(3, "0");
    const newCase: TestCase = {
      id: newId,
      title: newCaseTitle.trim(),
      component: newCaseComp.trim() || "General",
      type: newCaseType,
      priority: newCasePriority,
      status: "Untested",
    };

    saveCases([...testCases, newCase]);
    setIsCaseModalOpen(false);
    setNewCaseTitle("");
    setNewCaseComp("");
    setNewCaseType("E2E");
    setNewCasePriority("High");
  };

  const updateTestCaseStatus = (caseId: string, status: TestCase["status"]) => {
    // 1. Update list
    const updatedCases = testCases.map((tc) => tc.id === caseId ? { ...tc, status } : tc);
    setTestCases(updatedCases);
    localStorage.setItem("devhub-test-cases", JSON.stringify(updatedCases));

    // 2. Update active cycle
    if (cycle) {
      const updatedCycle = {
        ...cycle,
        results: {
          ...cycle.results,
          [caseId]: status,
        }
      };
      setCycle(updatedCycle);
      localStorage.setItem("devhub-test-cycle", JSON.stringify(updatedCycle));
    }
  };

  // Compute stats for current cycle
  const getCycleStats = () => {
    if (!cycle) return { Pass: 0, Fail: 0, Blocked: 0, Untested: 0, total: 0, progress: 0 };
    const counts = { Pass: 0, Fail: 0, Blocked: 0, Untested: 0 };
    let total = 0;

    Object.values(cycle.results).forEach((status) => {
      if (counts[status] !== undefined) {
        counts[status]++;
        total++;
      }
    });

    const tested = counts.Pass + counts.Fail + counts.Blocked;
    const progress = total === 0 ? 0 : Math.round((tested / total) * 100);

    return { ...counts, total, progress };
  };

  const stats = getCycleStats();

  const chartData = [
    { name: "Pass", value: stats.Pass, color: CHART_COLORS.Pass },
    { name: "Fail", value: stats.Fail, color: CHART_COLORS.Fail },
    { name: "Blocked", value: stats.Blocked, color: CHART_COLORS.Blocked },
    { name: "Untested", value: stats.Untested, color: CHART_COLORS.Untested },
  ].filter((d) => d.value > 0);

  const filteredCases = testCases.filter((tc) => {
    const matchesSearch = tc.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
                          tc.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          tc.component.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = typeFilter === "all" || tc.type === typeFilter;
    const matchesPriority = priorityFilter === "all" || tc.priority === priorityFilter;
    return matchesSearch && matchesType && matchesPriority;
  });

  return (
    <div className="space-y-10 pb-20 px-4 md:px-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
        <DashboardHeader 
          titlePrefix="Test"
          titleGradient="Management Suite"
          subtitle="테스트 케이스 설계, 실행 주기 수립, 그리고 결과 분석 일정을 조율합니다."
        />
        <div className="flex items-center gap-3">
          {activeTab === "schedule" && (
            <button 
              onClick={() => setIsScheduleModalOpen(true)}
              className="flex items-center gap-2 px-5 py-3 rounded-2xl bg-gradient-to-r from-primary to-accent hover:opacity-90 text-primary-foreground font-bold text-sm shadow-lg shadow-primary/20 transition-all shrink-0"
            >
              <Plus className="w-4 h-4" /> 테스트 일정 추가
            </button>
          )}
          {activeTab === "cases" && (
            <button 
              onClick={() => setIsCaseModalOpen(true)}
              className="flex items-center gap-2 px-5 py-3 rounded-2xl bg-gradient-to-r from-primary to-accent hover:opacity-90 text-primary-foreground font-bold text-sm shadow-lg shadow-primary/20 transition-all shrink-0"
            >
              <Plus className="w-4 h-4" /> 테스트 케이스 추가
            </button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-white/10 gap-6">
        {[
          { key: "cycle", label: "테스트 사이클 & 통계", icon: PlayCircle },
          { key: "cases", label: "테스트 케이스 관리", icon: ListTodo },
          { key: "schedule", label: "테스트 일정 관리", icon: Calendar },
        ].map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key as any)}
              className={`pb-4 text-sm font-bold flex items-center gap-2 transition-all relative ${
                activeTab === tab.key 
                  ? "text-primary border-b-2 border-primary" 
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <Icon className="w-4 h-4" />
              {tab.label}
              {activeTab === tab.key && (
                <motion.div 
                  layoutId="active-test-tab" 
                  className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" 
                />
              )}
            </button>
          );
        })}
      </div>

      {/* Tab Panels */}
      <div className="min-h-[400px]">
        {activeTab === "cycle" && cycle && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            
            {/* Left/Middle: Test Cycle Queue */}
            <div className="lg:col-span-2 space-y-6">
              <div className="glass-card p-6 border border-white/10 dark:border-white/5">
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h3 className="text-lg font-bold text-foreground">{cycle.name}</h3>
                    <p className="text-xs text-muted-foreground mt-1">대상 빌드 버전: <Badge variant="secondary" className="font-mono">{cycle.targetVersion}</Badge></p>
                  </div>
                  <Badge variant={stats.progress === 100 ? "success" : "warning"}>진행률 {stats.progress}%</Badge>
                </div>
                
                {/* Progress bar */}
                <div className="w-full h-3 bg-white/10 dark:bg-black/20 rounded-full overflow-hidden mb-8">
                  <motion.div 
                    className="h-full bg-primary rounded-full" 
                    initial={{ width: 0 }}
                    animate={{ width: `${stats.progress}%` }}
                    transition={{ duration: 0.5 }}
                  />
                </div>

                <h4 className="text-xs font-black text-muted-foreground uppercase tracking-widest mb-4">테스트 실행 큐 (Execution Queue)</h4>
                <div className="space-y-4 max-h-[420px] overflow-y-auto pr-2">
                  {testCases.map((tc) => {
                    const currentStatus = cycle.results[tc.id] || "Untested";
                    
                    return (
                      <div key={tc.id} className="p-4 rounded-xl border border-white/5 bg-white/5 backdrop-blur-md flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <h5 className="text-xs font-bold text-foreground flex items-center gap-2">
                            <Badge variant="secondary" className="scale-90 font-mono">{tc.id}</Badge>
                            {tc.title}
                          </h5>
                          <p className="text-[10px] text-muted-foreground mt-1">{tc.component} • {tc.type} • Priority: {tc.priority}</p>
                        </div>
                        
                        {/* Status Quick Actions */}
                        <div className="flex items-center gap-1.5 shrink-0">
                          {[
                            { status: "Pass" as const, bg: "hover:bg-emerald-500/25", border: "border-emerald-500/30 text-emerald-500", activeBg: "bg-emerald-500 text-white border-emerald-500" },
                            { status: "Fail" as const, bg: "hover:bg-rose-500/25", border: "border-rose-500/30 text-rose-500", activeBg: "bg-rose-500 text-white border-rose-500" },
                            { status: "Blocked" as const, bg: "hover:bg-amber-500/25", border: "border-amber-500/30 text-amber-500", activeBg: "bg-amber-500 text-white border-amber-500" }
                          ].map((act) => (
                            <button
                              key={act.status}
                              onClick={() => updateTestCaseStatus(tc.id, act.status)}
                              className={`px-3 py-1.5 rounded-lg border text-[10px] font-bold transition-all ${
                                currentStatus === act.status 
                                  ? act.activeBg 
                                  : `bg-white/5 ${act.border} ${act.bg}`
                              }`}
                            >
                              {act.status}
                            </button>
                          ))}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>

            {/* Right: Pie Chart Stats & Metrics */}
            <div className="space-y-6">
              <div className="glass-card p-8 border border-white/10 dark:border-white/5 flex flex-col items-center">
                <h3 className="text-md font-bold text-foreground mb-6 self-start flex items-center gap-2">
                  <TrendingUp className="w-4 h-4 text-primary" /> 테스트 실행 통계
                </h3>
                
                {/* Donut Chart */}
                <div className="w-full h-56 relative flex items-center justify-center">
                  {chartData.length > 0 ? (
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={chartData}
                          cx="50%"
                          cy="50%"
                          innerRadius={65}
                          outerRadius={85}
                          paddingAngle={3}
                          dataKey="value"
                        >
                          {chartData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                        </Pie>
                        <Tooltip />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="text-xs text-muted-foreground">데이터 없음</div>
                  )}
                  <div className="absolute flex flex-col items-center justify-center">
                    <span className="text-3xl font-black font-mono text-foreground">{stats.Pass}</span>
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-0.5">Passed / {stats.total}</span>
                  </div>
                </div>

                {/* Legend Table */}
                <div className="w-full space-y-3 mt-6 border-t border-white/5 pt-6">
                  {Object.entries(STATUS_ICONS).map(([key, value]) => {
                    const count = stats[key as keyof typeof stats] ?? 0;
                    const pct = stats.total === 0 ? 0 : Math.round((count / stats.total) * 100);
                    const colorMap = {
                      Pass: "bg-primary",
                      Fail: "bg-rose-500",
                      Blocked: "bg-amber-500",
                      Untested: "bg-zinc-500",
                    };
                    
                    return (
                      <div key={key} className="flex items-center justify-between text-xs">
                        <div className="flex items-center gap-2 text-muted-foreground">
                          <div className={`w-2 h-2 rounded-full ${colorMap[key as keyof typeof colorMap]}`} />
                          <value.icon className={`w-3.5 h-3.5 ${value.color}`} />
                          <span>{value.label}</span>
                        </div>
                        <span className="font-mono font-bold text-foreground">{count}건 ({pct}%)</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>

          </div>
        )}

        {activeTab === "cases" && (
          <div className="glass-card p-6 border border-white/10 dark:border-white/5 space-y-6">
            
            {/* Filter Bar */}
            <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
              <div className="relative w-full md:max-w-xs">
                <Search className="w-4 h-4 text-muted-foreground absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  placeholder="ID, 제목, 컴포넌트 검색..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 text-sm rounded-xl border border-white/10 bg-white/5 text-foreground focus:outline-none focus:border-primary transition-all"
                />
              </div>

              <div className="flex gap-3 w-full md:w-auto">
                <select
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value)}
                  className="px-3 py-2 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary"
                >
                  <option value="all">모든 테스트 유형</option>
                  <option value="E2E">E2E</option>
                  <option value="Unit">Unit</option>
                  <option value="Integration">Integration</option>
                </select>

                <select
                  value={priorityFilter}
                  onChange={(e) => setPriorityFilter(e.target.value)}
                  className="px-3 py-2 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary"
                >
                  <option value="all">모든 우선순위</option>
                  <option value="High">High</option>
                  <option value="Medium">Medium</option>
                  <option value="Low">Low</option>
                </select>
              </div>
            </div>

            {/* Test Case Table */}
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-white/10 text-[10px] font-black text-muted-foreground uppercase tracking-wider">
                    <th className="py-3 px-4">TC ID</th>
                    <th className="py-3 px-4">제목</th>
                    <th className="py-3 px-4">컴포넌트</th>
                    <th className="py-3 px-4">유형</th>
                    <th className="py-3 px-4">우선순위</th>
                    <th className="py-3 px-4">최근 결과</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5 text-xs">
                  {filteredCases.map((tc) => {
                    const iconConfig = STATUS_ICONS[tc.status];
                    const Icon = iconConfig.icon;
                    
                    return (
                      <tr key={tc.id} className="hover:bg-white/5 transition-all">
                        <td className="py-3 px-4 font-mono font-bold text-foreground">{tc.id}</td>
                        <td className="py-3 px-4 font-bold text-foreground">{tc.title}</td>
                        <td className="py-3 px-4 text-muted-foreground">{tc.component}</td>
                        <td className="py-3 px-4">
                          <Badge variant="secondary" className="font-semibold">{tc.type}</Badge>
                        </td>
                        <td className="py-3 px-4">
                          <Badge variant={PRIORITY_BADGES[tc.priority]}>{tc.priority}</Badge>
                        </td>
                        <td className="py-3 px-4">
                          <span className={`flex items-center gap-1 font-bold ${iconConfig.color}`}>
                            <Icon className="w-3.5 h-3.5" />
                            {iconConfig.label}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                  {filteredCases.length === 0 && (
                    <tr>
                      <td colSpan={6} className="py-12 text-center text-muted-foreground">
                        검색 조건에 맞는 테스트 케이스가 없습니다.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === "schedule" && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {schedules.map((sched) => (
              <div key={sched.id} className="glass-card p-6 border border-white/10 dark:border-white/5 flex flex-col justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Calendar className="w-4 h-4 text-primary" />
                    <span className="text-[10px] font-black uppercase tracking-wider text-muted-foreground">{sched.stage}</span>
                  </div>
                  <h4 className="text-md font-bold text-foreground mb-4">{sched.title}</h4>
                </div>

                <div className="space-y-4">
                  <div className="flex justify-between text-[10px] font-black text-muted-foreground uppercase tracking-wider">
                    <span>진행 진척도</span>
                    <span>{sched.progress}%</span>
                  </div>
                  <div className="w-full h-2 bg-white/10 dark:bg-black/20 rounded-full overflow-hidden">
                    <div className="h-full bg-primary rounded-full" style={{ width: `${sched.progress}%` }} />
                  </div>
                  <div className="flex justify-between text-[10px] text-muted-foreground pt-2 border-t border-white/5">
                    <span>시작일: {sched.startDate}</span>
                    <span>종료일: {sched.endDate}</span>
                  </div>
                </div>
              </div>
            ))}
            {schedules.length === 0 && (
              <div className="col-span-full py-12 text-center text-muted-foreground glass-card border border-dashed border-white/10">
                등록된 테스트 일정이 없습니다.
              </div>
            )}
          </div>
        )}
      </div>

      {/* Schedule Create Modal */}
      <AnimatePresence>
        {isScheduleModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 bg-black/60 backdrop-blur-md"
              onClick={() => setIsScheduleModalOpen(false)}
            />
            <motion.div 
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              className="relative w-full max-w-lg p-8 rounded-3xl glass border border-white/10 bg-card dark:bg-zinc-900/90 shadow-2xl z-10 space-y-6"
            >
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-black text-foreground flex items-center gap-2">
                  <Calendar className="w-5 h-5 text-primary" /> 신규 테스트 일정 생성
                </h3>
                <button 
                  onClick={() => setIsScheduleModalOpen(false)}
                  className="p-2 rounded-xl hover:bg-white/10 text-muted-foreground transition-all"
                >
                  ✕
                </button>
              </div>

              <form onSubmit={handleAddSchedule} className="space-y-4">
                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">일정명</label>
                  <input 
                    type="text"
                    value={newSchedTitle}
                    onChange={(e) => setNewSchedTitle(e.target.value)}
                    placeholder="예: 배포 전 E2E 시나리오 최종 검증"
                    required
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                  />
                </div>

                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">테스트 단계 / 영역</label>
                  <input 
                    type="text"
                    value={newSchedStage}
                    onChange={(e) => setNewSchedStage(e.target.value)}
                    placeholder="예: 통합 회귀 진단"
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">시작일</label>
                    <input 
                      type="date"
                      value={newSchedStart}
                      onChange={(e) => setNewSchedStart(e.target.value)}
                      required
                      className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">종료일</label>
                    <input 
                      type="date"
                      value={newSchedEnd}
                      onChange={(e) => setNewSchedEnd(e.target.value)}
                      required
                      className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                    />
                  </div>
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-white/5">
                  <button 
                    type="button" 
                    onClick={() => setIsScheduleModalOpen(false)}
                    className="px-5 py-2.5 rounded-xl border border-white/10 text-sm text-foreground hover:bg-white/5 transition-all"
                  >
                    취소
                  </button>
                  <button 
                    type="submit" 
                    className="px-5 py-2.5 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-all"
                  >
                    일정 추가 🚀
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        )}

        {isCaseModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 bg-black/60 backdrop-blur-md"
              onClick={() => setIsCaseModalOpen(false)}
            />
            <motion.div 
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              className="relative w-full max-w-lg p-8 rounded-3xl glass border border-white/10 bg-card dark:bg-zinc-900/90 shadow-2xl z-10 space-y-6"
            >
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-black text-foreground flex items-center gap-2">
                  <ListTodo className="w-5 h-5 text-primary" /> 신규 테스트 케이스 설계
                </h3>
                <button 
                  onClick={() => setIsCaseModalOpen(false)}
                  className="p-2 rounded-xl hover:bg-white/10 text-muted-foreground transition-all"
                >
                  ✕
                </button>
              </div>

              <form onSubmit={handleAddCase} className="space-y-4">
                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">테스트 케이스 제목</label>
                  <input 
                    type="text"
                    value={newCaseTitle}
                    onChange={(e) => setNewCaseTitle(e.target.value)}
                    placeholder="예: API 인증 실패 및 fallback 캐시 작동 테스트"
                    required
                    className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1">컴포넌트 명</label>
                    <input 
                      type="text"
                      value={newCaseComp}
                      onChange={(e) => setNewCaseComp(e.target.value)}
                      placeholder="예: Auth / API"
                      required
                      className="w-full px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-foreground text-sm focus:outline-none focus:border-primary transition-all"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1.5">테스트 유형</label>
                    <select
                      value={newCaseType}
                      onChange={(e) => setNewCaseType(e.target.value as any)}
                      className="w-full px-3 py-2.5 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary"
                    >
                      <option value="E2E">E2E</option>
                      <option value="Unit">Unit</option>
                      <option value="Integration">Integration</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest block mb-1.5">우선순위</label>
                  <select
                    value={newCasePriority}
                    onChange={(e) => setNewCasePriority(e.target.value as any)}
                    className="w-full px-3 py-2.5 rounded-xl border border-white/10 bg-zinc-850 text-foreground text-xs focus:outline-none focus:border-primary"
                  >
                    <option value="High">High (최고 우선)</option>
                    <option value="Medium">Medium (일반)</option>
                    <option value="Low">Low (낮음)</option>
                  </select>
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-white/5">
                  <button 
                    type="button" 
                    onClick={() => setIsCaseModalOpen(false)}
                    className="px-5 py-2.5 rounded-xl border border-white/10 text-sm text-foreground hover:bg-white/5 transition-all"
                  >
                    취소
                  </button>
                  <button 
                    type="submit" 
                    className="px-5 py-2.5 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-all"
                  >
                    케이스 추가 🚀
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
