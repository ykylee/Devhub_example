import { create } from "zustand";
import { persist } from "zustand/middleware";

export type ToastType = "info" | "success" | "warning" | "error";
export type UserRole = "Developer" | "Manager" | "System Admin";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

interface AppState {
  role: UserRole;
  setRole: (role: UserRole) => void;
  isDeepFocus: boolean;
  setDeepFocus: (active: boolean) => void;
  notifications: number;
  clearNotifications: () => void;
  incrementNotifications: () => void;
  toasts: Toast[];
  addToast: (message: string, type?: ToastType) => void;
  removeToast: (id: string) => void;
}

export const useStore = create<AppState>()(
  persist(
    (set) => ({
      role: "Developer",
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
    }),
    {
      name: "devhub-storage",
    }
  )
);
