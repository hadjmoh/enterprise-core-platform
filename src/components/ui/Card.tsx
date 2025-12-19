import React from 'react';
import { cn } from '../../lib/utils';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: 'default' | 'glass' | 'neon';
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(
    ({ className, variant = 'default', children, ...props }, ref) => {
        const variants = {
            default: "bg-slate-800 border border-slate-700/50",
            glass: "bg-slate-900/40 backdrop-blur-md border border-slate-700/50 shadow-glass",
            neon: "bg-slate-900 border border-brand-500/30 shadow-glow",
        };

        return (
            <div
                ref={ref}
                className={cn(
                    "rounded-xl p-6 text-slate-200",
                    variants[variant],
                    className
                )}
                {...props}
            >
                {children}
            </div>
        );
    }
);
Card.displayName = "Card";
