import { cn } from "@/lib/utils";

interface BadgeProps {
  children: React.ReactNode;
  variant?: "primary" | "secondary" | "accent" | "success" | "warning" | "danger" | "glass";
  className?: string;
  dot?: boolean;
}

const variants = {
  primary: "bg-primary/10 text-primary border-primary/20",
  secondary: "bg-white/5 text-white/70 border-white/10",
  accent: "bg-accent/10 text-accent border-accent/20",
  success: "bg-green-500/10 text-green-500 border-green-500/20",
  warning: "bg-amber-500/10 text-amber-500 border-amber-500/20",
  danger: "bg-rose-500/10 text-rose-500 border-rose-500/20",
  glass: "glass border-white/10 text-white/80",
};

export function Badge({ 
  children, 
  variant = "glass", 
  className,
  dot = false
}: BadgeProps) {
  return (
    <span className={cn(
      "inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-widest border",
      variants[variant],
      className
    )}>
      {dot && <span className={cn("w-1 h-1 rounded-full", variant === "glass" ? "bg-primary" : "bg-current")} />}
      {children}
    </span>
  );
}
