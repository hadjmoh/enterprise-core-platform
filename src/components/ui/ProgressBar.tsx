import { cn } from '../../lib/utils';

interface ProgressBarProps {
    value: number; // 0 to 100
    max?: number;
    color?: string; // Tailwind color class e.g., "bg-blue-500"
    showValue?: boolean;
    className?: string;
}

export function ProgressBar({ value, max = 100, color = "bg-brand-500", showValue = false, className }: ProgressBarProps) {
    const percentage = Math.min(100, Math.max(0, (value / max) * 100));

    return (
        <div className={cn("w-full flex items-center gap-3", className)}>
            <div className="h-2 flex-1 rounded-full bg-slate-800 overflow-hidden">
                <div
                    className={cn("h-full rounded-full transition-all duration-500", color)}
                    style={{ width: `${percentage}%` }}
                />
            </div>
            {showValue && (
                <span className="text-xs font-medium text-slate-400 w-8 text-right">
                    {Math.round(percentage)}%
                </span>
            )}
        </div>
    );
}
