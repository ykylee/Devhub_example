"use client";

import { useEffect, useState } from "react";
import { gardenerService, Suggestion } from "@/lib/services/gardener.service";
import { motion, AnimatePresence } from "framer-motion";
import { Sparkles, ArrowRight, CheckCircle2, ShieldAlert, TrendingUp, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";

export function GardenerFeed() {
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { addToast } = useStore();

  useEffect(() => {
    const fetchSuggestions = async () => {
      try {
        const data = await gardenerService.getSuggestions();
        setSuggestions(data);
      } catch {
        console.error("Failed to fetch suggestions");
      } finally {
        setIsLoading(false);
      }
    };

    fetchSuggestions();
  }, []);

  const handleApply = async (id: string, title: string) => {
    try {
      addToast(`Applying AI suggestion: ${title}`, "info");
      await gardenerService.applySuggestion(id);
      setSuggestions(prev => prev.filter(s => s.id !== id));
      addToast("Suggestion applied successfully.", "success");
    } catch {
      addToast("Failed to apply suggestion.", "error");
    }
  };

  const getIcon = (type: string) => {
    switch (type) {
      case "optimization": return <TrendingUp className="w-5 h-5 text-emerald-400" />;
      case "security": return <ShieldAlert className="w-5 h-5 text-rose-400" />;
      case "scaling": return <Zap className="w-5 h-5 text-amber-400" />;
      default: return <Sparkles className="w-5 h-5 text-primary" />;
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2].map(i => (
          <div key={i} className="glass rounded-2xl p-6 h-32 animate-pulse bg-white/5" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-primary/20 border border-primary/30">
            <Sparkles className="w-5 h-5 text-primary animate-pulse" />
          </div>
          <h3 className="text-xl font-black text-white uppercase tracking-tight">AI Gardener <span className="text-primary">Advisor</span></h3>
        </div>
        <span className="text-[10px] font-black text-muted-foreground uppercase tracking-widest bg-white/5 px-3 py-1 rounded-full border border-white/10">
          {suggestions.length} Active Insights
        </span>
      </div>

      <AnimatePresence mode="popLayout">
        {suggestions.length === 0 ? (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="glass rounded-2xl p-10 text-center border-dashed border-white/10"
          >
            <CheckCircle2 className="w-12 h-12 text-emerald-500/30 mx-auto mb-4" />
            <p className="text-muted-foreground text-sm font-medium italic">All systems are optimized. No suggestions at this time.</p>
          </motion.div>
        ) : (
          <div className="grid grid-cols-1 gap-4">
            {suggestions.map((suggestion, index) => (
              <motion.div
                key={suggestion.id}
                layout
                initial={{ opacity: 0, y: 20, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9, x: 20 }}
                transition={{ duration: 0.4, delay: index * 0.1 }}
                className="group relative glass rounded-2xl p-6 border-white/10 hover:border-primary/30 transition-all duration-500 overflow-hidden"
              >
                {/* Glow effect */}
                <div className="absolute -inset-1 bg-gradient-to-r from-primary/10 to-accent/10 rounded-2xl blur opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                
                <div className="relative flex items-start gap-5">
                  <div className={cn(
                    "p-3 rounded-2xl bg-white/5 border border-white/10",
                    suggestion.impact === 'high' ? "border-rose-500/30 bg-rose-500/5" : "border-white/10"
                  )}>
                    {getIcon(suggestion.type)}
                  </div>
                  
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-1">
                      <h4 className="text-lg font-bold text-white group-hover:text-primary transition-colors">
                        {suggestion.title}
                      </h4>
                      <div className={cn(
                        "text-[9px] font-black uppercase tracking-[0.2em] px-2 py-0.5 rounded border",
                        suggestion.impact === 'high' ? "text-rose-400 border-rose-500/20 bg-rose-500/10" :
                        suggestion.impact === 'medium' ? "text-amber-400 border-amber-500/20 bg-amber-500/10" :
                        "text-emerald-400 border-emerald-500/20 bg-emerald-500/10"
                      )}>
                        {suggestion.impact} Impact
                      </div>
                    </div>
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-4 leading-relaxed">
                      {suggestion.description}
                    </p>
                    
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-mono text-white/30 uppercase">
                        {new Date(suggestion.created_at).toLocaleTimeString()}
                      </span>
                      
                      <button 
                        onClick={() => handleApply(suggestion.id, suggestion.title)}
                        className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-primary hover:text-accent transition-all group/btn"
                      >
                        Apply Optimization <ArrowRight className="w-3 h-3 group-hover/btn:translate-x-1 transition-transform" />
                      </button>
                    </div>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
