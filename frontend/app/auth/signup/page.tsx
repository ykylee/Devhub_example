"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { ShieldCheck, User, IdCard, Building2, Key, Loader2, ArrowRight, CheckCircle2, AlertCircle } from "lucide-react";
import Link from "next/link";
import type { LucideIcon } from "lucide-react";

export default function SignUpPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    name: "",
    system_id: "",
    employee_id: "",
    password: "",
    confirmPassword: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (formData.password !== formData.confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setSubmitting(true);
    try {
      const response = await fetch("/api/v1/auth/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: formData.name,
          system_id: formData.system_id,
          employee_id: formData.employee_id,
          password: formData.password,
        }),
      });

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.details || data.error || "Sign up failed.");
      }

      setSuccess(true);
      setTimeout(() => router.push("/login"), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An unexpected error occurred.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center p-6 relative overflow-hidden">
      {/* Background Glows */}
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/20 blur-[120px] rounded-full animate-pulse" />
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-accent/20 blur-[120px] rounded-full animate-pulse" />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-xl w-full"
      >
        <div className="glass border-white/10 rounded-[2.5rem] overflow-hidden shadow-2xl relative z-10">
          <div className="bg-white/5 px-10 py-12 border-b border-white/10 text-center space-y-4">
            <div className="w-16 h-16 bg-gradient-to-br from-primary to-accent rounded-2xl mx-auto flex items-center justify-center shadow-lg ring-4 ring-white/5">
              <ShieldCheck className="w-9 h-9 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-black text-white tracking-tighter uppercase">
                Join <span className="text-primary">DevHub</span>
              </h1>
              <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest mt-2">
                Self-Registration for Internal Personnel
              </p>
            </div>
          </div>

          <div className="p-10">
            <AnimatePresence mode="wait">
              {success ? (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="text-center space-y-6 py-8"
                >
                  <div className="w-20 h-20 bg-emerald-500/20 border border-emerald-500/20 rounded-full mx-auto flex items-center justify-center">
                    <CheckCircle2 className="w-10 h-10 text-emerald-400" />
                  </div>
                  <div className="space-y-2">
                    <h2 className="text-2xl font-bold text-white">Identity Verified!</h2>
                    <p className="text-muted-foreground">Your account has been created successfully. Redirecting to sign in...</p>
                  </div>
                  <div className="pt-4">
                    <Loader2 className="w-6 h-6 text-primary animate-spin mx-auto" />
                  </div>
                </motion.div>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <InputField
                      label="Full Name"
                      icon={User}
                      placeholder="e.g. YK Lee"
                      value={formData.name}
                      onChange={(v) => setFormData({ ...formData, name: v })}
                    />
                    <InputField
                      label="System ID"
                      icon={Building2}
                      placeholder="e.g. yklee"
                      value={formData.system_id}
                      onChange={(v) => setFormData({ ...formData, system_id: v })}
                    />
                  </div>

                  <InputField
                    label="Employee ID (사번)"
                    icon={IdCard}
                    placeholder="e.g. 1001"
                    value={formData.employee_id}
                    onChange={(v) => setFormData({ ...formData, employee_id: v })}
                  />

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <InputField
                      label="Password"
                      icon={Key}
                      type="password"
                      placeholder="••••••••"
                      value={formData.password}
                      onChange={(v) => setFormData({ ...formData, password: v })}
                    />
                    <InputField
                      label="Confirm Password"
                      icon={Key}
                      type="password"
                      placeholder="••••••••"
                      value={formData.confirmPassword}
                      onChange={(v) => setFormData({ ...formData, confirmPassword: v })}
                    />
                  </div>

                  {error && (
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: "auto" }}
                      className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl flex items-start gap-3"
                    >
                      <AlertCircle className="w-5 h-5 text-rose-400 shrink-0 mt-0.5" />
                      <p className="text-sm text-rose-400 font-medium leading-relaxed">{error}</p>
                    </motion.div>
                  )}

                  <button
                    type="submit"
                    disabled={submitting}
                    className="w-full bg-primary text-white font-black py-4 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-xl shadow-primary/20 disabled:opacity-50 uppercase tracking-widest text-sm"
                  >
                    {submitting ? (
                      <>
                        <Loader2 className="w-5 h-5 animate-spin" />
                        Verifying with HR...
                      </>
                    ) : (
                      <>
                        Register Account
                        <ArrowRight className="w-5 h-5" />
                      </>
                    )}
                  </button>

                  <div className="text-center pt-4">
                    <p className="text-sm text-muted-foreground">
                      Already have an account?{" "}
                      <Link href="/login" className="text-primary font-bold hover:underline">
                        Sign In
                      </Link>
                    </p>
                  </div>
                </form>
              )}
            </AnimatePresence>
          </div>
        </div>

        {/* Development Helper: Mock HR Data */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
          className="mt-8 p-6 glass border-white/5 rounded-3xl"
        >
          <div className="flex items-center gap-2 mb-4">
            <div className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
            <p className="text-[10px] font-black text-white/40 uppercase tracking-[0.2em]">Development Mode: Mock HR Database</p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {[
              { name: "YK Lee", id: "yklee", emp: "1001" },
              { name: "Alex Kim", id: "akim", emp: "1002" },
              { name: "Sam Jones", id: "sjones", emp: "1003" },
            ].map((p) => (
              <div key={p.id} className="p-3 bg-white/5 rounded-xl border border-white/5 space-y-1 hover:bg-white/10 transition-colors cursor-pointer"
                onClick={() => setFormData({ ...formData, name: p.name, system_id: p.id, employee_id: p.emp })}
              >
                <p className="text-[11px] font-bold text-white">{p.name}</p>
                <p className="text-[9px] text-muted-foreground">ID: <span className="text-white/60">{p.id}</span> • Emp: <span className="text-white/60">{p.emp}</span></p>
              </div>
            ))}
          </div>
          <p className="mt-4 text-[9px] text-white/20 italic text-center italic">Tip: Click a card to auto-fill the form for testing.</p>
        </motion.div>
      </motion.div>
    </div>
  );
}

type InputFieldProps = {
  label: string;
  icon: LucideIcon;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: string;
};

function InputField({ label, icon: Icon, value, onChange, placeholder, type = "text" }: InputFieldProps) {
  return (
    <div className="space-y-2">
      <label className="text-[10px] font-black text-white/40 uppercase tracking-widest px-1">
        {label}
      </label>
      <div className="relative group">
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/20 group-focus-within:text-primary transition-colors">
          <Icon className="w-4 h-4" />
        </div>
        <input
          type={type}
          required
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full bg-white/5 border border-white/10 rounded-2xl pl-12 pr-4 py-3.5 text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-white/10"
        />
      </div>
    </div>
  );
}
