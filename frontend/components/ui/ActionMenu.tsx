"use client";

import { ReactNode, useEffect, useMemo, useState } from "react";
import { MoreHorizontal } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ActionMenuItem {
  key: string;
  label: string;
  onClick: () => void;
  icon?: ReactNode;
  tone?: "default" | "danger";
}

interface ActionMenuProps {
  title?: string;
  items: ActionMenuItem[];
  triggerClassName?: string;
  menuClassName?: string;
  widthPx?: number;
}

export function ActionMenu({
  title = "Actions",
  items,
  triggerClassName,
  menuClassName,
  widthPx = 192,
}: ActionMenuProps) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState({ top: 0, left: 0 });

  useEffect(() => {
    if (!open) return;
    const onEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", onEsc);
    return () => window.removeEventListener("keydown", onEsc);
  }, [open]);

  const visibleItems = useMemo(() => items.filter(Boolean), [items]);

  return (
    <>
      <button
        type="button"
        onPointerDown={(e) => {
          e.preventDefault();
          e.stopPropagation();
          if (open) {
            setOpen(false);
            return;
          }
          const rect = (e.currentTarget as HTMLButtonElement).getBoundingClientRect();
          const left = Math.max(12, Math.min(rect.right - widthPx, window.innerWidth - widthPx - 12));
          setPos({ top: rect.bottom + 8, left });
          setOpen(true);
        }}
        className={cn("p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-colors", triggerClassName)}
        aria-label={title}
      >
        <MoreHorizontal className="w-5 h-5" />
      </button>

      {open && (
        <>
          <div
            className="fixed inset-0 z-40"
            onPointerDown={(e) => {
              e.preventDefault();
              setOpen(false);
            }}
          />
          <div
            className={cn("fixed z-[60] glass bg-popover/90 border border-border rounded-xl overflow-hidden shadow-2xl py-1", menuClassName)}
            style={{ top: pos.top, left: pos.left, width: widthPx }}
            onPointerDown={(e) => e.stopPropagation()}
          >
            <div className="px-3 py-2 border-b border-border">
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest text-left">{title}</p>
            </div>
            <div className="py-1">
              {visibleItems.map((item) => (
                <button
                  key={item.key}
                  type="button"
                  onPointerDown={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setOpen(false);
                    item.onClick();
                  }}
                  className={cn(
                    "w-full flex items-center gap-2 px-3 py-2 text-xs transition-colors text-left",
                    item.tone === "danger"
                      ? "text-red-400 hover:bg-red-400/10"
                      : "text-foreground hover:bg-primary/10",
                  )}
                >
                  {item.icon}
                  {item.label}
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </>
  );
}
