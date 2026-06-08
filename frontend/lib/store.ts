import { create } from "zustand";
import { persist, subscribeWithSelector } from "zustand/middleware";

export type ToastType = "info" | "success" | "warning" | "error";
export type UserRole = "Developer" | "Manager" | "System Admin";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

export interface AuthenticatedActor {
  login: string;
  user_id?: string;
  subject?: string;
  role: UserRole;
  source?: string;
  display_name?: string;
  email?: string;
  primary_unit_id?: string | null;
  onboarding_required?: boolean;
  onboarding_completed_at?: string | null;
  review_status?: "pending_review" | "reviewed" | null;
}

interface AppState {
  role: UserRole | null;
  actor: AuthenticatedActor | null;
  setActor: (actor: AuthenticatedActor) => void;
  clearActor: () => void;
  setRole: (role: UserRole | null) => void;
  isDeepFocus: boolean;
  setDeepFocus: (active: boolean) => void;
  notifications: number;
  clearNotifications: () => void;
  incrementNotifications: () => void;
  toasts: Toast[];
  addToast: (message: string, type?: ToastType) => void;
  removeToast: (id: string) => void;
  isLoggingOut: boolean;
  setIsLoggingOut: (active: boolean) => void;
  isSidebarOpen: boolean;
  setSidebarOpen: (active: boolean) => void;
  isSidebarCollapsed: boolean;
  setSidebarCollapsed: (active: boolean) => void;
}
 
export const useStore = create<AppState>()(
  subscribeWithSelector(
    persist(
      (set, get) => ({
        role: null,
        actor: null,
        // setActor 가 logout 진행 중에 외부에서 호출되어도 stale session 을
        // 다시 박지 않도록 단일 gate 로 차단. (issue #488 spec 정합 502 분기.)
        //
        // 동기:
        //   backend POST /api/v1/auth/logout 이 502 (Keycloak unreachable) 를
        //   정상 응답하는 환경 — CI E2E 컨테이너 flake 가 대표 사례 — 에서
        //   backend 는 access token revoke 를 못 하고 logout 다음 200/401
        //   가 race 한다. frontend 의 AuthGuard 가 pathname 변경에 반응해
        //   whoAmI() 를 재호출하면, backend 가 200 (revoke 못 했으므로) 으로
        //   답해 actor store 에 stale session 이 다시 박힌다. 그 사이
        //   `window.location.assign("/login")` 이 발사돼도 router 가 next
        //   render 에서 AuthGuard 가 막은 `/developer` 또는 `/login` 진입
        //   직전에 stale actor 가 살아있어 protected 경로로 다시 redirect
        //   되는 deadlock.
        //
        //   isLoggingOut 이 true 인 동안 setActor 호출은 모두 no-op 처리.
        //   단, login-success path (auth.service.ts:241) 와 onboarding/profile
        //   update path 에서 의도적으로 호출하는 경우는 isLoggingOut=false
        //   상태이므로 영향 없음. (logout 자체도 line 206 에서 setIsLoggingOut
        //   true → setActor skip → setIsLoggingOut(false) 는 navigation 완료
        //   시점이다.)
        setActor: (actor) => {
          if (get().isLoggingOut) return;
          set({ actor, role: actor.role });
        },
        clearActor: () => set({ actor: null, role: null }),
        setRole: (role) => set({ role }),
        isDeepFocus: false,
        setDeepFocus: (active) => set({ isDeepFocus: active }),
        notifications: 3,
        clearNotifications: () => set({ notifications: 0 }),
        incrementNotifications: () => set((state) => ({ notifications: state.notifications + 1 })),
        toasts: [],
        addToast: (message, type = "info") => {
          const id = Math.random().toString(36).substring(2, 9);
          set((state) => ({ 
            toasts: [...state.toasts, { id, message, type }] 
          }));
          setTimeout(() => {
            set((state) => ({ 
              toasts: state.toasts.filter((t) => t.id !== id) 
            }));
          }, 5000);
        },
        removeToast: (id) => set((state) => ({ 
          toasts: state.toasts.filter((t) => t.id !== id) 
        })),
        isLoggingOut: false,
        setIsLoggingOut: (active) => set({ isLoggingOut: active }),
        isSidebarOpen: false,
        setSidebarOpen: (active) => set({ isSidebarOpen: active }),
        isSidebarCollapsed: false,
        setSidebarCollapsed: (active) => set({ isSidebarCollapsed: active }),
      }),
      {
        name: "devhub-storage",
        partialize: (state) => {
          const { isLoggingOut, toasts, isSidebarOpen, ...rest } = state;
          void isLoggingOut;
          void toasts;
          void isSidebarOpen;
          return rest;
        },
      }
    )
  )
);
