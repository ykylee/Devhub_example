"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { Lock, User, ArrowRight, ShieldCheck } from "lucide-react";
import { useRouter } from "next/navigation";
import { useStore } from "@/lib/store";

export default function LoginPage() {
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const setRole = useStore((state) => state.setRole);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    
    // MOCK LOGIN LOGIC - Will be replaced by Kratos flow in Phase 13
    setTimeout(() => {
      if (loginId === "admin") {
        setRole("System Admin");
        router.push("/organization");
      } else if (loginId === "manager") {
        setRole("Manager");
        router.push("/manager");
      } else {
        setRole("Developer");
        router.push("/developer");
      }
      setIsLoading(false);
    }, 1000);
  };

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4 selection:bg-primary/30">
      {/* Background Orbs */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -left-1/4 w-1/2 h-1/2 bg-primary/10 rounded-full blur-[120px]" />
        <div className="absolute bottom-1/4 -right-1/4 w-1/2 h-1/2 bg-accent/10 rounded-full blur-[120px]" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
      >
        <div className="text-center mb-10">
          <motion.div
            initial={{ scale: 0.5 }}
            animate={{ scale: 1 }}
            className="inline-flex p-4 rounded-3xl bg-gradient-to-br from-primary/20 to-accent/20 border border-white/10 mb-6 shadow-2xl"
          >
            <ShieldCheck className="w-12 h-12 text-white" />
          </motion.div>
          <h1 className="text-4xl font-black text-white tracking-tighter uppercase mb-2">
            DevHub <span className="text-primary">Identity</span>
          </h1>
          <p className="text-muted-foreground text-sm font-bold uppercase tracking-widest">
            Unified Authentication Gateway
          </p>
        </div>

        <div className="glass border-white/10 rounded-[2rem] p-10 shadow-2xl backdrop-blur-2xl">
          <form onSubmit={handleLogin} className="space-y-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest ml-1">Login ID</label>
              <div className="relative">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input
                  type="text"
                  value={loginId}
                  onChange={(e) => setLoginId(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-4 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/10"
                  placeholder="Enter your ID"
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest ml-1">Password</label>
              <div className="relative">
                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-4 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/10"
                  placeholder="••••••••"
                  required
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full bg-primary text-white font-black py-4 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 disabled:hover:scale-100 group uppercase tracking-widest text-xs"
            >
              {isLoading ? "Authenticating..." : (
                <>
                  Access System
                  <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                </>
              )}
            </button>
          </form>

          <div className="mt-8 pt-8 border-t border-white/5 text-center">
            <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest">
              Secured by Ory Hydra + Kratos
            </p>
          </div>
        </div>

        <div className="mt-8 flex justify-center gap-6">
          <p className="text-[9px] text-white/20 font-bold uppercase">Node: ASIA-01</p>
          <p className="text-[9px] text-white/20 font-bold uppercase">v0.5.0-BETA</p>
        </div>
      </motion.div>
    </div>
  );
}
