import { cn } from '../../lib/utils';

interface MetricRowProps {
    label: string;
    value: string | number;
    subValue?: string;
    status?: 'normal' | 'warning' | 'critical';
}

export function MetricRow({ label, value, subValue, status = 'normal' }: MetricRowProps) {
    return (
        <div className="flex items-center justify-between py-2 border-b border-slate-800/50 last:border-0">
            <span className="text-sm text-slate-400">{label}</span>
            <div className="text-right">
                <div className={cn(
                    "text-sm font-medium",
                    status === 'critical' ? "text-red-400" :
                        status === 'warning' ? "text-amber-400" :
                            "text-slate-200"
                )}>
                    {value}
                </div>
                {subValue && <div className="text-xs text-slate-500">{subValue}</div>}
            </div>
        </div>
    );
}
