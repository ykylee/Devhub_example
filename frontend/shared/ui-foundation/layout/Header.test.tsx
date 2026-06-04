import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// 외부 의존 service / next 모듈은 모두 vi.mock 으로 격리.
const routerPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: routerPush }),
}));

const logoutMock = vi.fn();
vi.mock("@/domain/auth-session/service/auth.service", () => ({
  authService: {
    logout: () => logoutMock(),
  },
}));

type Subscriber = (event: { data: unknown }) => void;
const subscribers = new Map<string, Subscriber[]>();
function publish(type: string, data: unknown) {
  (subscribers.get(type) ?? []).forEach((cb) => cb({ data }));
}

vi.mock("@/domain/realtime/service/realtime.service", () => ({
  realtimeService: {
    isConnected: true,
    subscribe: (type: string, cb: Subscriber) => {
      const arr = subscribers.get(type) ?? [];
      arr.push(cb);
      subscribers.set(type, arr);
      return () => {
        const next = (subscribers.get(type) ?? []).filter((c) => c !== cb);
        subscribers.set(type, next);
      };
    },
  },
}));

const dreqListMock = vi.fn();
const dreqRegisterMock = vi.fn();
vi.mock("@/domain/dev-request/service/dev_request.service", () => ({
  devRequestService: {
    list: (params: unknown) => dreqListMock(params),
    register: (id: string, payload: unknown) => dreqRegisterMock(id, payload),
  },
}));

const repoListMock = vi.fn();
vi.mock("@/domain/repository-integration/service/repository.service", () => ({
  repositoryService: {
    listRepositories: () => repoListMock(),
  },
}));

// Header 가 import 하는 modal 컴포넌트들은 사이드 이펙트 무거우니 stub.
// onChanged / onPromote / onCreated 콜백을 외부에서 확인 가능하도록 ref 노출.
const detailHandlersRef: {
  current: {
    onClose?: () => void;
    onChanged?: () => void;
    onPromote?: (req: unknown) => void;
  };
} = { current: {} };
vi.mock("@/domain/dev-request/view/DevRequestDetailModal", () => ({
  DevRequestDetailModal: (props: {
    request: { id: string; title: string };
    onClose: () => void;
    onChanged: () => void;
    onPromote: (req: unknown) => void;
  }) => {
    detailHandlersRef.current = {
      onClose: props.onClose,
      onChanged: props.onChanged,
      onPromote: props.onPromote,
    };
    return (
      <div data-testid="dreq-detail-modal">
        <button onClick={props.onClose}>close</button>
      </div>
    );
  },
}));

const createHandlersRef: {
  current: {
    onClose?: () => void;
    onCreated?: (proj: { id: string }) => void;
  };
} = { current: {} };
vi.mock("@/domain/platform-lifecycle/view/ProjectCreationModal", () => ({
  ProjectCreationModal: (props: {
    onClose: () => void;
    onCreated: (proj: { id: string }) => void;
  }) => {
    createHandlersRef.current = {
      onClose: props.onClose,
      onCreated: props.onCreated,
    };
    return (
      <div data-testid="project-create-modal">
        <button onClick={props.onClose}>close-create</button>
      </div>
    );
  },
}));

vi.mock("framer-motion", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };
  const motion = new Proxy(
    {},
    {
      get: (_target, tag) =>
        ({ children, ...props }: AnyProps) =>
          React.createElement(tag as string, props, children),
    },
  );
  return {
    motion,
    AnimatePresence: ({ children }: AnyProps) =>
      React.createElement(React.Fragment, null, children),
  };
});

import { useStore, type AuthenticatedActor } from "@/lib/store";
import { Header } from "./Header";

function resetAll() {
  subscribers.clear();
  routerPush.mockClear();
  logoutMock.mockClear();
  dreqListMock.mockReset();
  dreqRegisterMock.mockReset();
  repoListMock.mockReset();
  act(() => {
    useStore.setState({
      role: null,
      actor: null,
      notifications: 0,
    });
  });
  // 기본 모킹: 빈 응답.
  dreqListMock.mockResolvedValue({ data: [], total: 0 });
  repoListMock.mockResolvedValue([]);
}

const devActor: AuthenticatedActor = { login: "alice", role: "Developer" };
const adminActor: AuthenticatedActor = { login: "admin", role: "System Admin" };

describe("Header (F-1)", () => {
  beforeEach(() => {
    resetAll();
  });

  it("Guest 상태에서 'Guest User' / 'No Role' 이 표기된다", async () => {
    render(<Header />);
    // mount-effect 의 dreqListMock / repoListMock flush 대기.
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
    expect(screen.getByText("Guest User")).toBeInTheDocument();
    expect(screen.getByText("No Role")).toBeInTheDocument();
  });

  it("actor 가 set 되어 있으면 actor.login 과 role 이 노출된다", async () => {
    act(() => {
      useStore.setState({ actor: devActor, role: "Developer" });
    });
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    expect(screen.getByText("alice")).toBeInTheDocument();
    expect(screen.getByText("Developer")).toBeInTheDocument();
  });

  it("초기 fetchDreqs 실패 시에도 헤더는 정상 렌더된다 (console.error 만)", async () => {
    const err = vi.spyOn(console, "error").mockImplementation(() => {});
    dreqListMock.mockRejectedValueOnce(new Error("network"));
    render(<Header />);
    await waitFor(() => expect(err).toHaveBeenCalled());
    expect(screen.getByLabelText(/notifications/i)).toBeInTheDocument();
    err.mockRestore();
  });

  it("초기 fetchRepos 실패 시에도 헤더는 정상 렌더된다", async () => {
    const err = vi.spyOn(console, "error").mockImplementation(() => {});
    repoListMock.mockRejectedValueOnce(new Error("repo-err"));
    render(<Header />);
    await waitFor(() => expect(err).toHaveBeenCalled());
    expect(screen.getByLabelText(/notifications/i)).toBeInTheDocument();
    err.mockRestore();
  });

  it("User menu 버튼 클릭 시 dropdown 이 열린다 (Account Profile / Sign Out 노출)", async () => {
    act(() => {
      useStore.setState({ actor: devActor, role: "Developer" });
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());

    await user.click(screen.getByRole("button", { name: /user menu/i }));
    expect(screen.getByRole("menuitem", { name: /account profile/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /sign out/i })).toBeInTheDocument();
  });

  it("System Admin 인 경우 dropdown 에 System Settings 가 추가된다", async () => {
    act(() => {
      useStore.setState({ actor: adminActor, role: "System Admin" });
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());

    await user.click(screen.getByRole("button", { name: /user menu/i }));
    expect(screen.getByRole("menuitem", { name: /system settings/i })).toBeInTheDocument();
  });

  it("Account Profile 클릭 시 /account 로 이동", async () => {
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /user menu/i }));
    await user.click(screen.getByRole("menuitem", { name: /account profile/i }));
    expect(routerPush).toHaveBeenCalledWith("/account");
  });

  it("Sign Out 클릭 시 authService.logout 이 호출된다", async () => {
    logoutMock.mockResolvedValue(undefined);
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /user menu/i }));
    await user.click(screen.getByRole("menuitem", { name: /sign out/i }));
    await waitFor(() => expect(logoutMock).toHaveBeenCalledTimes(1));
  });

  it("Theme 토글 버튼 클릭 시 theme-dark 클래스가 html 에 추가된다", async () => {
    document.documentElement.classList.remove("theme-dark");
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /user menu/i }));

    const themeBtn = screen.getByRole("menuitem", { name: /light mode|dark mode/i });
    await user.click(themeBtn);
    expect(document.documentElement.classList.contains("theme-dark")).toBe(true);
  });

  it("초기 theme-dark 클래스가 있는 상태에서 toggle 시 제거된다", async () => {
    document.documentElement.classList.add("theme-dark");
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByRole("button", { name: /user menu/i }));
    const themeBtn = screen.getByRole("menuitem", { name: /light mode|dark mode/i });
    await user.click(themeBtn);
    expect(document.documentElement.classList.contains("theme-dark")).toBe(false);
  });

  it("Notification bell 클릭 시 pending DREQ 목록이 노출된다", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [
        {
          id: "d1",
          external_ref: "JIRA-100",
          title: "Fix login bug",
          requester: "alice",
          received_at: "2026-05-29T00:00:00Z",
        },
      ],
      total: 1,
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());

    await user.click(screen.getByLabelText(/notifications/i));
    expect(screen.getByText("Fix login bug")).toBeInTheDocument();
    expect(screen.getByText("JIRA-100")).toBeInTheDocument();
  });

  it("Notification dropdown 의 'View All Dev Requests' 클릭 시 /dev-requests 로 이동", async () => {
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByRole("button", { name: /view all dev requests/i }));
    expect(routerPush).toHaveBeenCalledWith("/dev-requests");
  });

  it("Pending dreq 항목 클릭 시 DevRequestDetailModal 이 노출된다", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [
        {
          id: "d2",
          external_ref: "X-1",
          title: "Demo",
          requester: "bob",
          received_at: "2026-05-29T00:00:00Z",
        },
      ],
      total: 1,
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByText("Demo"));
    expect(screen.getByTestId("dreq-detail-modal")).toBeInTheDocument();
  });

  it("Mobile menu 버튼 클릭 시 store.setSidebarOpen(true) 가 호출된다", async () => {
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText("Open sidebar"));
    expect(useStore.getState().isSidebarOpen).toBe(true);
  });

  it("realtime status.changed 이벤트가 오면 fetchDreqs 가 다시 호출된다", async () => {
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
    dreqListMock.mockClear();
    dreqListMock.mockResolvedValue({ data: [], total: 0 });
    act(() => {
      publish("status.changed", { connected: false });
    });
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
  });

  it("realtime dev_request.created 이벤트가 오면 fetchDreqs 가 다시 호출된다", async () => {
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
    dreqListMock.mockClear();
    dreqListMock.mockResolvedValue({ data: [], total: 0 });
    act(() => {
      publish("dev_request.created", { id: "x" });
    });
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
  });

  it("realtime notification.created 이벤트가 오면 store.incrementNotifications + addToast", async () => {
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    const before = useStore.getState().notifications;
    act(() => {
      publish("notification.created", { message: "Hi" });
    });
    expect(useStore.getState().notifications).toBe(before + 1);
    expect(useStore.getState().toasts.find((t) => t.message === "Hi")).toBeDefined();
  });

  it("realtime risk.critical.created 이벤트가 오면 error toast 가 추가된다", async () => {
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    act(() => {
      publish("risk.critical.created", { message: "Boom" });
    });
    const toast = useStore.getState().toasts.find((t) => t.message.includes("Boom"));
    expect(toast).toBeDefined();
    expect(toast?.type).toBe("error");
  });

  it("DetailModal onPromote → ProjectCreationModal 노출 + onCreated 시 register 호출 + fetchDreqs 재호출", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [
        {
          id: "d3",
          external_ref: "JIRA-PROM",
          title: "promotion target",
          requester: "carol",
          received_at: "2026-05-29T00:00:00Z",
          details: "details body",
        },
      ],
      total: 1,
    });
    dreqRegisterMock.mockResolvedValue({});

    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByText("promotion target"));

    // DetailModal 의 onPromote 콜백 직접 호출 → ProjectCreationModal 노출.
    act(() => {
      detailHandlersRef.current.onPromote?.({
        id: "d3",
        external_ref: "JIRA-PROM",
        title: "promotion target",
        details: "details body",
      });
    });
    await waitFor(() => {
      expect(screen.getByTestId("project-create-modal")).toBeInTheDocument();
    });

    // ProjectCreationModal 의 onCreated 호출 → devRequestService.register 가 호출되어야 함.
    act(() => {
      createHandlersRef.current.onCreated?.({ id: "proj-99" });
    });
    await waitFor(() => {
      expect(dreqRegisterMock).toHaveBeenCalledWith("d3", {
        target_type: "project",
        target_id: "proj-99",
      });
    });
  });

  it("DetailModal onChanged 호출 시 fetchDreqs 가 재실행된다", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [
        {
          id: "d4",
          external_ref: "X-2",
          title: "alt",
          requester: "u",
          received_at: "2026-05-29T00:00:00Z",
        },
      ],
      total: 1,
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByText("alt"));
    dreqListMock.mockClear();
    dreqListMock.mockResolvedValue({ data: [], total: 0 });
    act(() => {
      detailHandlersRef.current.onChanged?.();
    });
    await waitFor(() => expect(dreqListMock).toHaveBeenCalledTimes(1));
  });

  it("DetailModal onClose 호출 시 detail modal 이 닫힌다", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [{ id: "d5", external_ref: "x", title: "z", requester: "u", received_at: "2026-05-29T00:00:00Z" }],
      total: 1,
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByText("z"));
    expect(screen.getByTestId("dreq-detail-modal")).toBeInTheDocument();
    act(() => detailHandlersRef.current.onClose?.());
    await waitFor(() => expect(screen.queryByTestId("dreq-detail-modal")).toBeNull());
  });

  it("ProjectCreationModal onClose 호출 시 project modal 이 닫힌다", async () => {
    dreqListMock.mockResolvedValueOnce({
      data: [{ id: "d6", external_ref: "x", title: "q", requester: "u", received_at: "2026-05-29T00:00:00Z" }],
      total: 1,
    });
    const user = userEvent.setup();
    render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    await user.click(screen.getByLabelText(/notifications/i));
    await user.click(screen.getByText("q"));
    act(() => {
      detailHandlersRef.current.onPromote?.({ id: "d6", external_ref: "x", title: "q", details: "" });
    });
    await waitFor(() => expect(screen.getByTestId("project-create-modal")).toBeInTheDocument());
    act(() => createHandlersRef.current.onClose?.());
    await waitFor(() => expect(screen.queryByTestId("project-create-modal")).toBeNull());
  });

  it("unmount 시 모든 realtime subscription 이 해제된다", async () => {
    const { unmount } = render(<Header />);
    await waitFor(() => expect(dreqListMock).toHaveBeenCalled());
    expect(subscribers.get("status.changed")?.length).toBeGreaterThanOrEqual(1);
    unmount();
    expect(subscribers.get("status.changed")?.length ?? 0).toBe(0);
    expect(subscribers.get("dev_request.created")?.length ?? 0).toBe(0);
  });
});
