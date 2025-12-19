import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface SingleValuePanelProps {
    value: string | number;
    label?: string;
    trend?: number; // Optional percentage change
    unit?: string;
}

export const SingleValuePanel: React.FC<SingleValuePanelProps> = ({ value, label, trend, unit }) => {
    return (
        <div className="flex flex-col items-center justify-center p-8 bg-slate-900/40 rounded-2xl border border-slate-800 shadow-inner group transition-all hover:border-indigo-500/30">
            {label && (
                <span className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-2 group-hover:text-indigo-400">
                    {label}
                </span>
            )}

            <div className="flex items-baseline gap-2">
                <span className="text-6xl font-black text-white tracking-tighter drop-shadow-lg">
                    {typeof value === 'number' ? value.toLocaleString() : value}
                </span>
                {unit && <span className="text-xl font-bold text-slate-500">{unit}</span>}
            </div>

            {trend !== undefined && (
                <div className={`mt-4 flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold ${trend > 0 ? 'bg-emerald-500/10 text-emerald-400' :
                        trend < 0 ? 'bg-red-500/10 text-red-400' :
                            'bg-slate-500/10 text-slate-400'
                    }`}>
                    {trend > 0 ? <TrendingUp className="w-3 h-3" /> :
                        trend < 0 ? <TrendingDown className="w-3 h-3" /> :
                            <Minus className="w-3 h-3" />}
                    {Math.abs(trend)}%
                </div>
            )}
        </div>
    );
};
