"use client";

import { motion, AnimatePresence } from "framer-motion";
import { Info, CheckCircle2, AlertTriangle, XCircle, X, type LucideIcon } from "lucide-react";
import { useStore, type ToastType } from "@/lib/store";
import { cn } from "@/lib/utils";

const toastStyles: Record<ToastType, { icon: LucideIcon; color: string; bg: string; border: string }> = {
  info: { icon: Info, color: "text-blue-400", bg: "bg-blue-500/10", border: "border-blue-500/20" },
  success: { icon: CheckCircle2, color: "text-emerald-400", bg: "bg-emerald-500/10", border: "border-emerald-500/20" },
  warning: { icon: AlertTriangle, color: "text-amber-400", bg: "bg-amber-500/10", border: "border-amber-500/20" },
  error: { icon: XCircle, color: "text-rose-400", bg: "bg-rose-500/10", border: "border-rose-500/20" },
};

export function ToastContainer() {
  const { toasts, removeToast } = useStore();

  return (
    <div className="fixed bottom-6 right-6 z-[200] flex flex-col gap-3 w-full max-w-sm">
      <AnimatePresence>
        {toasts.map((toast) => {
          const style = toastStyles[toast.type];
          const Icon = style.icon;

          return (
            <motion.div
              key={toast.id}
              initial={{ opacity: 0, x: 50, scale: 0.9 }}
              animate={{ opacity: 1, x: 0, scale: 1 }}
              exit={{ opacity: 0, scale: 0.9, transition: { duration: 0.2 } }}
              className={cn(
                "glass p-4 rounded-2xl border flex items-start gap-3 shadow-2xl overflow-hidden relative group",
                style.bg,
                style.border
              )}
            >
              {/* Progress bar animation */}
              <motion.div 
                initial={{ width: "100%" }}
                animate={{ width: "0%" }}
                transition={{ duration: 5, ease: "linear" }}
                className={cn("absolute bottom-0 left-0 h-0.5 opacity-50", style.color.replace("text-", "bg-"))}
              />
              
              <Icon className={cn("w-5 h-5 shrink-0 mt-0.5", style.color)} />
              
              <div className="flex-1">
                <p className="text-sm font-bold text-foreground dark:text-primary-foreground leading-tight">{toast.message}</p>
              </div>
              
              <button 
                onClick={() => removeToast(toast.id)}
                className="p-1 rounded-lg hover:bg-muted/30 text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground transition-all"
              >
                <X className="w-4 h-4" />
              </button>
            </motion.div>
          );
        })}
      </AnimatePresence>
    </div>
  );
}

export function useToast() {
  const addToast = useStore((state) => state.addToast);
  
  return {
    toast: (message: string, type: ToastType = "info") => {
      addToast(message, type);
    },
    success: (message: string) => addToast(message, "success"),
    error: (message: string) => addToast(message, "error"),
    warning: (message: string) => addToast(message, "warning"),
    info: (message: string) => addToast(message, "info"),
  };
}
