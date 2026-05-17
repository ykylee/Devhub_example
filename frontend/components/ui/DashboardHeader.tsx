"use client";

import { motion } from "framer-motion";
import { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface DashboardHeaderProps {
  titlePrefix: string;
  titleGradient: string;
  subtitle: string | React.ReactNode;
  icon?: LucideIcon;
  actions?: React.ReactNode;
}

export function DashboardHeader({ 
  titlePrefix, 
  titleGradient, 
  subtitle, 
  icon: Icon,
  actions 
}: DashboardHeaderProps) {
  return (
    <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
      >
        <h1 className="text-4xl md:text-5xl font-black tracking-tight text-foreground dark:text-primary-foreground mb-3 leading-none">
          {titlePrefix} <span className="text-gradient hover:brightness-125 transition-all cursor-default">{titleGradient}</span>
        </h1>
        <div className="text-muted-foreground text-lg flex items-center gap-2 font-medium">
          {Icon && <Icon className="w-4 h-4 text-primary" />}
          {subtitle}
        </div>
      </motion.div>

      <motion.div 
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="flex items-center gap-4"
      >
        {actions}
      </motion.div>
    </div>
  );
}
