"use client";

import { useState, FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { authService } from "@/lib/services/auth.service";
import { ApiError } from "@/lib/services/api-client";

// /signup — self-service Sign Up form (RM-M3-01, sprint claude/work_260513-m).
// Posts to /api/v1/auth/signup (API-23, backend_api_contract §11.5.2). The
// backend validates the (name, system_id, employee_id) triple against the
// HR DB; on success the user can sign in via /login (OIDC flow).
//
// This page is public — AuthGuard does not gate /signup. The four input
// fields mirror the backend payload exactly; client-side validation is
// minimal (required) so server error codes surface to the user.

export default function SignUpPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    name: "",
    system_id: "",
    employee_id: "",
    password: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<{ user_id: string; department: string } | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const res = await authService.signup(form);
      setSuccess({ user_id: res.data.user_id, department: res.data.department });
    } catch (err) {
      if (err instanceof ApiError) {
        // Surface server-side codes verbatim (§11.5.2 error matrix).
        if (err.status === 403) {
          setError("HR DB 조회 실패 — 이름/사내 ID/사번이 인사 DB 와 일치하는지 확인하세요.");
        } else if (err.status === 400) {
          setError("입력값을 다시 확인하세요.");
        } else {
          setError(err.message || "가입 처리 중 오류가 발생했습니다.");
        }
      } else {
        setError("네트워크 오류로 가입을 완료하지 못했습니다.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4">
        <div className="w-full max-w-md text-center">
          <h1 className="text-3xl font-black text-white tracking-tight mb-4">
            가입 완료
          </h1>
          <p className="text-muted-foreground mb-2">
            <span className="text-primary font-bold">{success.user_id}</span> ({success.department})
          </p>
          <p className="text-muted-foreground mb-8">
            계정이 생성됐습니다. 로그인 페이지에서 발급받은 비밀번호로 로그인하세요.
          </p>
          <Link
            href="/login"
            className="inline-block px-6 py-3 rounded-xl bg-primary/20 hover:bg-primary/30 text-white font-bold transition-all"
          >
            로그인하기
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <h1 className="text-3xl font-black text-white tracking-tight mb-6">
          Sign Up
        </h1>
        <p className="text-sm text-muted-foreground mb-6">
          인사 DB 에 등록된 이름, 사내 ID, 사번이 일치해야 가입 가능합니다.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Field
            label="이름"
            name="name"
            value={form.name}
            onChange={(v) => setForm({ ...form, name: v })}
            required
          />
          <Field
            label="사내 ID"
            name="system_id"
            value={form.system_id}
            onChange={(v) => setForm({ ...form, system_id: v })}
            required
          />
          <Field
            label="사번"
            name="employee_id"
            value={form.employee_id}
            onChange={(v) => setForm({ ...form, employee_id: v })}
            required
          />
          <Field
            label="비밀번호"
            name="password"
            type="password"
            value={form.password}
            onChange={(v) => setForm({ ...form, password: v })}
            required
          />

          {error && (
            <div data-testid="signup-error" className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-200 text-sm">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="w-full py-3 rounded-xl bg-primary hover:bg-primary/90 disabled:opacity-50 text-white font-bold transition-all"
          >
            {submitting ? "처리 중..." : "가입"}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          이미 계정이 있으신가요? <Link href="/login" className="text-primary hover:underline">로그인</Link>
        </p>
      </div>
    </div>
  );
}

interface FieldProps {
  label: string;
  name: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  required?: boolean;
}

function Field({ label, name, value, onChange, type = "text", required }: FieldProps) {
  return (
    <label className="block">
      <span className="block text-xs font-bold text-muted-foreground uppercase tracking-wider mb-2">
        {label}
      </span>
      <input
        name={name}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
      />
    </label>
  );
}
