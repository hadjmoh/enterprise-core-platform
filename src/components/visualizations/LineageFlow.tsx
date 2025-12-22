import React from 'react';
import { Shield, CheckCircle, Clock, Zap, Target, Database } from 'lucide-react';
import { cn } from '../../lib/utils';

interface LineageStep {
    stage: string;
    node_id: string;
    timestamp: string;
    action: string;
    signature: string;
    verified?: boolean;
}

interface LineageFlowProps {
    steps: LineageStep[];
    isVerified?: boolean;
}

const STAGE_ICONS: Record<string, React.ReactNode> = {
    ingestion: <Zap className="w-4 h-4" />,
    enrichment: <Shield className="w-4 h-4" />,
    correlation: <CheckCircle className="w-4 h-4" />,
    detection: <Target className="w-4 h-4" />,
    storage: <Database className="w-4 h-4" />,
};

export function LineageFlow({ steps, isVerified = true }: LineageFlowProps) {
    return (
        <div className="relative">
            <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-slate-800" />

            <div className="space-y-8">
                {steps.map((step, idx) => (
                    <div key={idx} className="relative flex items-start gap-6 group">
                        {/* Status Icon */}
                        <div className={cn(
                            "relative z-10 w-12 h-12 rounded-xl flex items-center justify-center border-2 transition-all duration-300",
                            isVerified ? "bg-slate-900 border-indigo-500 shadow-[0_0_15px_rgba(99,102,241,0.2)]" : "bg-slate-900 border-red-500/50"
                        )}>
                            {STAGE_ICONS[step.stage] || <Clock className="w-4 h-4" />}
                        </div>

                        {/* Step Content */}
                        <div className="flex-1 bg-slate-800/40 border border-slate-700/50 rounded-xl p-4 group-hover:bg-slate-800/60 transition-colors">
                            <div className="flex justify-between items-start mb-2">
                                <div>
                                    <h4 className="text-sm font-bold text-white capitalize">{step.stage}</h4>
                                    <p className="text-xs text-slate-400 mt-0.5">{step.action}</p>
                                </div>
                                <div className="text-right">
                                    <div className="text-[10px] font-mono text-slate-500">{new Date(step.timestamp).toLocaleString()}</div>
                                    <div className="mt-1 flex items-center justify-end gap-1">
                                        <span className="text-[10px] bg-slate-900 text-indigo-300 px-1.5 py-0.5 rounded border border-indigo-500/30">
                                            {step.node_id}
                                        </span>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-4 pt-3 border-t border-slate-700/50 flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <Shield className={cn("w-3 h-3", isVerified ? "text-emerald-400" : "text-red-400")} />
                                    <span className="text-[10px] font-mono text-slate-500 truncate max-w-[150px]">
                                        {step.signature}
                                    </span>
                                </div>
                                {isVerified && (
                                    <span className="text-[10px] text-emerald-400 font-medium flex items-center gap-1">
                                        <CheckCircle className="w-3 h-3" />
                                        Verified
                                    </span>
                                )}
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {isVerified && (
                <div className="mt-8 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-emerald-500/20 flex items-center justify-center">
                        <Shield className="w-4 h-4 text-emerald-400" />
                    </div>
                    <div>
                        <p className="text-sm font-bold text-emerald-400">Immutable Chain-of-Custody</p>
                        <p className="text-xs text-emerald-500/70">Cryptographic audit trail successfully verified across all hops.</p>
                    </div>
                </div>
            )}
        </div>
    );
}
