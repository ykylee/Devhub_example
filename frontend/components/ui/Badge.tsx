import { cn } from "@/lib/utils";

interface BadgeProps {
  children: React.ReactNode;
  variant?: "primary" | "secondary" | "accent" | "success" | "warning" | "danger" | "glass";
  className?: string;
  dot?: boolean;
}

const variants = {
  primary: "bg-primary/10 text-primary border-primary/20",
  secondary: "bg-muted text-muted-foreground border-border",
  accent: "bg-accent/10 text-accent border-accent/20",
  success: "bg-success/10 text-success border-success/20",
  warning: "bg-warning/10 text-warning border-warning/20",
  danger: "bg-destructive/10 text-destructive border-destructive/20",
  glass: "glass text-foreground/80 border-border/50",
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
