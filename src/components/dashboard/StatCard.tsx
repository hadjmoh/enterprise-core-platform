import { Card } from '../ui/Card';
import { cn } from '../../lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface StatCardProps {
    title: string;
    value: string;
    trend?: 'up' | 'down' | 'neutral';
    trendValue?: string;
    icon: React.ReactNode;
    variant?: 'default' | 'glass' | 'neon';
}

export function StatCard({ title, value, trend, trendValue, icon, variant = 'default' }: StatCardProps) {
    return (
        <Card variant={variant} className="relative overflow-hidden">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <p className="text-sm font-medium text-slate-400">{title}</p>
                    <h3 className="text-2xl font-bold text-white mt-1">{value}</h3>
                </div>
                <div className={cn(
                    "p-2 rounded-lg",
                    variant === 'neon' ? "bg-brand-500/20 text-brand-400" : "bg-slate-800 text-slate-300"
                )}>
                    {icon}
                </div>
            </div>

            {trend && (
                <div className="flex items-center gap-2">
                    {trend === 'up' && <TrendingUp className="h-4 w-4 text-emerald-400" />}
                    {trend === 'down' && <TrendingDown className="h-4 w-4 text-red-400" />}
                    {trend === 'neutral' && <Minus className="h-4 w-4 text-slate-400" />}

                    <span className={cn(
                        "text-sm font-medium",
                        trend === 'up' && "text-emerald-400",
                        trend === 'down' && "text-red-400",
                        trend === 'neutral' && "text-slate-400",
                    )}>
                        {trendValue}
                    </span>
                    <span className="text-slate-500 text-sm">vs last hour</span>
                </div>
            )}
        </Card>
    );
}
