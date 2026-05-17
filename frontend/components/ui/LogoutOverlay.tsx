"use client";

import { useStore } from "@/lib/store";
import { motion, AnimatePresence } from "framer-motion";
import { Loader2 } from "lucide-react";

export function LogoutOverlay() {
  const isLoggingOut = useStore((state) => state.isLoggingOut);

  return (
    <AnimatePresence>
      {isLoggingOut && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-[9999] bg-background/80 backdrop-blur-md flex flex-col items-center justify-center gap-4"
        >
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
          >
            <Loader2 className="w-12 h-12 text-primary" />
          </motion.div>
          <div className="flex flex-col items-center">
            <h2 className="text-2xl font-black text-foreground uppercase tracking-tighter">Logging out...</h2>
            <p className="text-muted-foreground font-medium">Securing your session and redirecting.</p>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
