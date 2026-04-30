import Link from "next/link";
import { LayoutDashboard, Users, Server, Settings, Activity } from "lucide-react";
import { cn } from "@/lib/utils";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className, ...props }: SidebarProps) {
  return (
    <div className={cn("pb-12 w-64 border-r border-border bg-card/50 backdrop-blur-xl h-screen sticky top-0 flex flex-col", className)} {...props}>
      <div className="space-y-4 py-4 flex-1">
        <div className="px-3 py-2">
          <h2 className="mb-2 px-4 text-xl font-bold tracking-tight text-primary flex items-center gap-2">
            <div className="w-8 h-8 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-lg shadow-lg flex items-center justify-center">
              <Activity className="w-5 h-5 text-white" />
            </div>
            DevHub
          </h2>
          <div className="space-y-1 mt-6">
            <h3 className="px-4 text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
              Views
            </h3>
            <Link href="/developer" className="flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors">
              <LayoutDashboard className="h-4 w-4" />
              Developer
            </Link>
            <Link href="/manager" className="flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors">
              <Users className="h-4 w-4" />
              Manager
            </Link>
            <Link href="/admin" className="flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors">
              <Server className="h-4 w-4" />
              Sys Admin
            </Link>
          </div>
        </div>
      </div>
      <div className="p-4 border-t border-border mt-auto">
        <Link href="/settings" className="flex items-center gap-3 rounded-lg px-4 py-2 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors">
          <Settings className="h-4 w-4" />
          Settings
        </Link>
      </div>
    </div>
  );
}
