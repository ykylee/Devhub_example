"use client";

import { useEffect, useState, useCallback } from "react";
import { Sun, Moon } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export function ThemeToggle() {
  const [theme, setTheme] = useState<"light" | "dark">("light");
  const [mounted, setMounted] = useState(false);

  const applyTheme = useCallback((t: "light" | "dark") => {
    if (t === "dark") {
      document.documentElement.classList.add("theme-dark");
    } else {
      document.documentElement.classList.remove("theme-dark");
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 0);
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (mounted) {
      const savedTheme = localStorage.getItem("devhub-theme") as "light" | "dark";
      if (savedTheme) {
        setTimeout(() => {
          setTheme(savedTheme);
          applyTheme(savedTheme);
        }, 0);
      }
    }
  }, [mounted, applyTheme]);

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    applyTheme(newTheme);
    localStorage.setItem("devhub-theme", newTheme);
  };

  if (!mounted) return null;

  return (
    <motion.button
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={toggleTheme}
      className="p-2.5 rounded-xl hover:bg-muted/30 text-muted-foreground hover:text-primary-foreground transition-all flex items-center justify-center"
      title={theme === "light" ? "Switch to Dark Mode (Purple)" : "Switch to Light Mode (Blue)"}
    >
      <AnimatePresence mode="wait">
        <motion.div
          key={theme}
          initial={{ opacity: 0, rotate: -90, scale: 0.8 }}
          animate={{ opacity: 1, rotate: 0, scale: 1 }}
          exit={{ opacity: 0, rotate: 90, scale: 0.8 }}
          transition={{ duration: 0.2 }}
        >
          {theme === "light" ? (
            <Sun className="h-5 w-5 text-amber-500" />
          ) : (
            <Moon className="h-5 w-5 text-indigo-400" />
          )}
        </motion.div>
      </AnimatePresence>
    </motion.button>
  );
}
