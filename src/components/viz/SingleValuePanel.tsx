import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import type { Threshold } from '../../types/viz';

interface SingleValuePanelProps {
    value: string | number;
    label?: string;
    trend?: number; // Optional percentage change
    unit?: string;
    thresholds?: Threshold[];
}

export const SingleValuePanel: React.FC<SingleValuePanelProps> = ({ value, label, trend, unit, thresholds }) => {
    const numericValue = typeof value === 'number' ? value : parseFloat(String(value));

    // Find active threshold
    const activeThreshold = thresholds
        ?.sort((a, b) => b.value - a.value)
        .find(t => numericValue >= t.value);

    const valueColor = activeThreshold?.color || '#ffffff';
    const isCritical = activeThreshold?.severity === 'critical';

    return (
        <div className={`flex flex-col items-center justify-center p-8 bg-slate-900/40 rounded-2xl border border-slate-800 shadow-inner group transition-all duration-500 ${isCritical ? 'ring-2 ring-red-500/50 shadow-2xl shadow-red-500/10' : 'hover:border-brand-500/30'}`}>
            {label && (
                <span className="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] mb-2 group-hover:text-brand-400 transition-colors">
                    {label}
                </span>
            )}

            <div className="flex items-baseline gap-2">
                <span
                    style={{ color: valueColor }}
                    className={`text-6xl font-black tracking-tighter drop-shadow-2xl transition-all duration-700 ${isCritical ? 'animate-pulse' : ''}`}
                >
                    {typeof value === 'number' ? value.toLocaleString() : value}
                </span>
                {unit && <span className="text-xl font-bold text-slate-500/50">{unit}</span>}
            </div>

            {trend !== undefined && (
                <div className={`mt-4 flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black ${trend > 0 ? 'bg-emerald-500/10 text-emerald-400' :
                    trend < 0 ? 'bg-red-500/10 text-red-400' :
                        'bg-slate-500/10 text-slate-400'
                    }`}>
                    {trend > 0 ? <TrendingUp className="w-3 h-3" /> :
                        trend < 0 ? <TrendingDown className="w-3 h-3" /> :
                            <Minus className="w-3 h-3" />}
                    {Math.abs(trend)}%
                </div>
            )}

            {activeThreshold?.label && (
                <div className="mt-2 text-[8px] font-bold text-slate-600 uppercase tracking-widest bg-slate-800/50 px-2 py-0.5 rounded">
                    Status: {activeThreshold.label}
                </div>
            )}
        </div>
    );
};
