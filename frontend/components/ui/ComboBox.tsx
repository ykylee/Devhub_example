"use client";

import { useState, useRef, useEffect, useMemo } from "react";
import { Search, ChevronDown, X, Check } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";

interface Option {
  label: string;
  value: string;
  description?: string;
}

interface ComboBoxProps {
  options: Option[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  emptyText?: string;
  className?: string;
  disabled?: boolean;
}

export function ComboBox({
  options,
  value,
  onChange,
  placeholder = "Select option...",
  emptyText = "No results found.",
  className,
  disabled = false,
}: ComboBoxProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);

  const selectedOption = useMemo(() => options.find((o) => o.value === value), [options, value]);

  const filteredOptions = useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return options;
    return options.filter(
      (o) => o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q) || o.description?.toLowerCase().includes(q),
    );
  }, [options, search]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    if (open) setSearch("");
  }, [open]);

  return (
    <div className={cn("relative", className)} ref={containerRef}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen(!open)}
        className={cn(
          "flex items-center justify-between w-full px-4 py-3 rounded-xl glass border border-border/50 text-sm font-medium transition-all text-left",
          open && "border-primary/50 ring-2 ring-primary/20",
          disabled && "opacity-50 cursor-not-allowed",
        )}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span className={cn("truncate", !selectedOption && "text-muted-foreground")}>
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <ChevronDown className={cn("w-4 h-4 text-muted-foreground transition-transform", open && "rotate-180")} />
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 10, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 10, scale: 0.95 }}
            className="absolute z-[100] mt-2 w-full glass-card border border-border shadow-2xl rounded-2xl overflow-hidden flex flex-col max-h-72"
          >
            <div className="p-2 border-b border-border/60 bg-muted/20 flex items-center gap-2">
              <Search className="w-4 h-4 text-muted-foreground ml-2" />
              <input
                autoFocus
                type="text"
                placeholder="Search..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full bg-transparent border-none outline-none py-2 text-sm text-foreground placeholder:text-muted-foreground"
              />
              {search && (
                <button onClick={() => setSearch("")} className="p-1 rounded-lg hover:bg-muted/50">
                  <X className="w-3 h-3 text-muted-foreground" />
                </button>
              )}
            </div>

            <div className="overflow-y-auto flex-1 py-1">
              {filteredOptions.length === 0 ? (
                <p className="px-4 py-8 text-xs text-muted-foreground text-center italic">{emptyText}</p>
              ) : (
                filteredOptions.map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    role="option"
                    aria-selected={value === option.value}
                    onClick={() => {
                      onChange(option.value);
                      setOpen(false);
                    }}
                    className={cn(
                      "w-full text-left px-4 py-3 text-sm transition-colors flex items-center justify-between group hover:bg-primary/10",
                      value === option.value ? "bg-primary/10 text-primary font-bold" : "text-foreground",
                    )}
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span>{option.label}</span>
                        {value === option.value && <Check className="w-3 h-3" />}
                      </div>
                      {option.description && (
                        <p className="text-[10px] text-muted-foreground font-normal mt-0.5 group-hover:text-primary/70">
                          {option.description}
                        </p>
                      )}
                    </div>
                    <span className="text-[10px] font-mono text-muted-foreground opacity-40 group-hover:opacity-100">
                      {option.value}
                    </span>
                  </button>
                ))
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
