import type { AlertContextType } from '../types/alert';
import { createContext, useContext } from 'react';

export const AlertContext = createContext<AlertContextType | undefined>(undefined);

export function useAlerts() {
    const context = useContext(AlertContext);
    if (context === undefined) {
        throw new Error('useAlerts must be used within an AlertProvider');
    }
    return context;
}
