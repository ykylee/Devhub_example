"use client";

import { Search, Filter, X } from "lucide-react";
import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";

interface FilterOption {
  label: string;
  value: string;
}

interface FilterBarProps {
  onSearch: (query: string) => void;
  onFilterChange: (value: string) => void;
  filterOptions: FilterOption[];
  placeholder?: string;
  activeFilter?: string;
  searchLabel?: string;
}

export function FilterBar({ 
  onSearch, 
  onFilterChange, 
  filterOptions, 
  placeholder = "Search resources...",
  activeFilter = "all",
  searchLabel
}: FilterBarProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      onSearch(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, onSearch]);

  return (
    <div className="flex flex-col md:flex-row items-center gap-4 w-full">
      <div className="relative flex-1 w-full">
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none">
          <Search className="w-4 h-4" />
        </div>
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder={placeholder}
          aria-label={searchLabel}
          className="w-full pl-11 pr-11 py-3 rounded-2xl glass border border-border/50 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all text-sm font-medium"
        />
        {searchQuery && (
          <button 
            onClick={() => setSearchQuery("")}
            className="absolute right-4 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      <div className="flex items-center gap-2 w-full md:w-auto">
        <div className="hidden lg:flex items-center gap-1 p-1 glass border border-border/50 rounded-xl">
          {filterOptions.map((option) => (
            <button
              key={option.value}
              onClick={() => onFilterChange(option.value)}
              className={`px-4 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${
                activeFilter === option.value 
                  ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20" 
                  : "text-muted-foreground hover:bg-muted/50"
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>

        {/* Mobile/Compact Filter Dropdown */}
        <div className="lg:hidden relative w-full md:w-auto">
          <button 
            onClick={() => setIsFilterOpen(!isFilterOpen)}
            className={`flex items-center justify-between gap-3 px-4 py-3 rounded-2xl glass border border-border/50 w-full md:w-48 text-sm font-medium ${isFilterOpen ? 'border-primary/50 ring-2 ring-primary/20' : ''}`}
          >
            <span className="flex items-center gap-2 text-muted-foreground">
              <Filter className="w-4 h-4" />
              {filterOptions.find(o => o.value === activeFilter)?.label || "Filter"}
            </span>
          </button>

          <AnimatePresence>
            {isFilterOpen && (
              <>
                <div 
                  className="fixed inset-0 z-40" 
                  onClick={() => setIsFilterOpen(false)}
                />
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  className="absolute right-0 top-full mt-2 w-full md:w-64 glass-card border border-border p-2 z-50 shadow-2xl"
                >
                  <div className="space-y-1">
                    {filterOptions.map((option) => (
                      <button
                        key={option.value}
                        onClick={() => {
                          onFilterChange(option.value);
                          setIsFilterOpen(false);
                        }}
                        className={`w-full text-left px-4 py-2.5 rounded-xl text-xs font-bold transition-all ${
                          activeFilter === option.value 
                            ? "bg-primary/10 text-primary" 
                            : "text-muted-foreground hover:bg-muted/50"
                        }`}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                </motion.div>
              </>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
