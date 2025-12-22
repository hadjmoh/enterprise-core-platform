import React from 'react';
import { cn } from '../../lib/utils';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
    variant?: 'default' | 'success' | 'warning' | 'error' | 'outline';
}

export const Badge = ({ className, variant = 'default', ...props }: BadgeProps) => {
    const variants = {
        default: "bg-brand-500/10 text-brand-400 border-brand-500/20",
        success: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
        warning: "bg-amber-500/10 text-amber-400 border-amber-500/20",
        error: "bg-red-500/10 text-red-400 border-red-500/20",
        outline: "border border-slate-600 text-slate-400",
    };

    return (
        <span
            className={cn(
                "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border transition-all",
                variants[variant],
                props.onClick && "cursor-pointer hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-brand-500/50 focus:ring-offset-1 focus:ring-offset-slate-900",
                className
            )}
            role={props.onClick ? "button" : props.role}
            tabIndex={props.onClick ? (props.tabIndex ?? 0) : props.tabIndex}
            onKeyDown={(e) => {
                if (props.onClick && (e.key === 'Enter' || e.key === ' ')) {
                    e.preventDefault();
                    (props.onClick as any)(e);
                }
            }}
            {...props}
        />
    );
};
