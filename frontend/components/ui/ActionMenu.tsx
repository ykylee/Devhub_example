"use client";

import { ReactNode, useEffect, useMemo, useRef, useState } from "react";
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
  const touchedRef = useRef(false);
  const touchedResetTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!open) return;
    const onEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", onEsc);
    return () => window.removeEventListener("keydown", onEsc);
  }, [open]);

  useEffect(() => {
    return () => {
      if (touchedResetTimer.current) clearTimeout(touchedResetTimer.current);
    };
  }, []);

  const markTouched = () => {
    touchedRef.current = true;
    // Auto-clear so a touch sequence that never fires onClick (touchcancel,
    // unmount mid-tap, certain webviews) cannot silently swallow the next
    // mouse click. 350ms covers the typical touch->click delay window.
    if (touchedResetTimer.current) clearTimeout(touchedResetTimer.current);
    touchedResetTimer.current = setTimeout(() => {
      touchedRef.current = false;
    }, 350);
  };

  const visibleItems = useMemo(() => items.filter(Boolean), [items]);
  const toggleFromTarget = (target: HTMLButtonElement) => {
    if (open) {
      setOpen(false);
      return;
    }
    const rect = target.getBoundingClientRect();
    const left = Math.max(12, Math.min(rect.right - widthPx, window.innerWidth - widthPx - 12));
    setPos({ top: rect.bottom + 8, left });
    setOpen(true);
  };

  const runItemAction = (item: ActionMenuItem) => {
    setOpen(false);
    item.onClick();
  };

  return (
    <>
      <button
        type="button"
        onTouchEnd={(e) => {
          e.preventDefault();
          e.stopPropagation();
          markTouched();
          toggleFromTarget(e.currentTarget as HTMLButtonElement);
        }}
        onClick={(e) => {
          if (touchedRef.current) {
            touchedRef.current = false;
            if (touchedResetTimer.current) {
              clearTimeout(touchedResetTimer.current);
              touchedResetTimer.current = null;
            }
            return;
          }
          e.preventDefault();
          e.stopPropagation();
          toggleFromTarget(e.currentTarget as HTMLButtonElement);
        }}
        className={cn("p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-colors cursor-pointer touch-manipulation", triggerClassName)}
        aria-haspopup="menu"
        aria-expanded={open}
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
            role="menu"
            aria-label={title}
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
                  role="menuitem"
                  // onPointerDown drives the touch/mouse path (avoids the 300ms
                  // tap-to-click delay on iOS). onClick is the keyboard path —
                  // Enter/Space fire click but not pointer events.
                  onPointerDown={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    runItemAction(item);
                  }}
                  onClick={(e) => {
                    // pointerType is empty when the click was synthesized from
                    // keyboard activation; in that case onPointerDown didn't
                    // run, so we still need to drive the action here.
                    const nativeEvent = e.nativeEvent as PointerEvent;
                    if (nativeEvent && typeof nativeEvent.pointerType === "string" && nativeEvent.pointerType !== "") {
                      // already handled by onPointerDown
                      return;
                    }
                    e.preventDefault();
                    e.stopPropagation();
                    runItemAction(item);
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
