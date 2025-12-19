import { createContext } from 'react';
import type { TimezoneMode } from '../types/ui';

export interface UIContextType {
    timezone: TimezoneMode;
    setTimezone: (mode: TimezoneMode) => void;
    isLocked: boolean;
    setIsLocked: (locked: boolean) => void;
}

export const UIContext = createContext<UIContextType | undefined>(undefined);
