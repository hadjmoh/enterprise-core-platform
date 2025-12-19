import React from 'react';
import { cn } from '../../lib/utils';

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: 'text' | 'rect' | 'circle';
}

export function Skeleton({ className, variant = 'rect', ...props }: SkeletonProps) {
    const variants = {
        text: "h-3 w-full rounded",
        rect: "h-full w-full rounded-lg",
        circle: "h-10 w-10 rounded-full",
    };

    return (
        <div
            className={cn(
                "animate-pulse bg-slate-800/50",
                variants[variant],
                className
            )}
            {...props}
        />
    );
}
