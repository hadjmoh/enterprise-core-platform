import { useState, type ReactNode, useEffect } from 'react';
import type { TimezoneMode } from '../types/ui';
import { UIContext } from './UIContextInstance';

export function UIProvider({ children }: { children: ReactNode }) {
    const [timezone, setTimezone] = useState<TimezoneMode>(() => {
        const saved = localStorage.getItem('enterprise_core_timezone');
        return (saved as TimezoneMode) || 'Local';
    });

    const [isLocked, setIsLocked] = useState<boolean>(() => {
        const saved = localStorage.getItem('enterprise_core_layout_locked');
        return saved === 'true';
    });

    useEffect(() => {
        localStorage.setItem('enterprise_core_timezone', timezone);
    }, [timezone]);

    useEffect(() => {
        localStorage.setItem('enterprise_core_layout_locked', String(isLocked));
    }, [isLocked]);

    return (
        <UIContext.Provider value={{ timezone, setTimezone, isLocked, setIsLocked }}>
            {children}
        </UIContext.Provider>
    );
}
