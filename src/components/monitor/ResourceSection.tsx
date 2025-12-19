import { Card } from '../ui/Card';
import { ProgressBar } from '../ui/ProgressBar';
import type { LucideIcon } from 'lucide-react';

interface ResourceSectionProps {
    title: string;
    icon: LucideIcon;
    children: React.ReactNode;
    mainMetric?: {
        label: string;
        value: string;
        percent: number;
        color?: string;
    };
    variant?: 'default' | 'glass' | 'neon';
}

export function ResourceSection({ title, icon: Icon, children, mainMetric, variant = 'default' }: ResourceSectionProps) {
    return (
        <Card variant={variant} className="flex flex-col h-full">
            <div className="flex items-center gap-2 mb-4">
                <Icon className="h-5 w-5 text-slate-400" />
                <h3 className="font-semibold text-white">{title}</h3>
            </div>

            {mainMetric && (
                <div className="mb-6 p-4 rounded-xl bg-slate-900/50 border border-slate-800">
                    <div className="flex justify-between items-end mb-2">
                        <span className="text-slate-400 text-sm">{mainMetric.label}</span>
                        <span className="text-2xl font-bold text-white">{mainMetric.value}</span>
                    </div>
                    <ProgressBar
                        value={mainMetric.percent}
                        color={mainMetric.color}
                    />
                </div>
            )}

            <div className="flex-1 space-y-1">
                {children}
            </div>
        </Card>
    );
}
