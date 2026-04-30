"use client";

import { useState } from "react";
import { Search, Bell, User, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

interface HeaderProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Header({ className, ...props }: HeaderProps) {
  const [role, setRole] = useState("Developer");

  return (
    <header className={cn("sticky top-0 z-50 w-full border-b border-border bg-background/80 backdrop-blur-xl", className)} {...props}>
      <div className="flex h-16 items-center px-6 gap-4">
        <div className="flex-1 flex items-center">
          <div className="relative w-96 max-w-md hidden md:flex items-center">
            <Search className="absolute left-2.5 h-4 w-4 text-muted-foreground" />
            <input
              type="search"
              placeholder="Search wiki, components, or logs..."
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 pl-9 bg-card/50"
            />
          </div>
        </div>
        
        <div className="flex items-center gap-4">
          <button className="relative p-2 text-muted-foreground hover:text-foreground transition-colors">
            <Bell className="h-5 w-5" />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-indigo-500 rounded-full animate-pulse"></span>
          </button>
          
          <div className="h-8 w-px bg-border"></div>

          <div className="flex items-center gap-2 group cursor-pointer relative">
            <div className="w-8 h-8 rounded-full bg-accent flex items-center justify-center border border-border">
              <User className="w-4 h-4 text-muted-foreground" />
            </div>
            <div className="flex flex-col hidden sm:flex">
              <span className="text-sm font-medium leading-none">YK Lee</span>
              <span className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                {role} <ChevronDown className="w-3 h-3 group-hover:rotate-180 transition-transform" />
              </span>
            </div>
            
            {/* Simple Dropdown for Prototype */}
            <div className="absolute top-full right-0 mt-2 w-48 rounded-md border border-border bg-popover text-popover-foreground shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none group-hover:pointer-events-auto z-50">
              <div className="p-1">
                {["Developer", "Manager", "System Admin"].map((r) => (
                  <button
                    key={r}
                    onClick={() => setRole(r)}
                    className={cn(
                      "flex w-full items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground",
                      role === r && "bg-accent/50 text-accent-foreground font-medium"
                    )}
                  >
                    {r} View
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
