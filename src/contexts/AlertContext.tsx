import { useState, useCallback, useEffect, type ReactNode } from 'react';
import type { Alert } from '../types/alert';
import { AlertContext } from './AlertContextInstance';

const STORAGE_KEY = 'enterprise_core_alerts';

export function AlertProvider({ children }: { children: ReactNode }) {
    const [alerts, setAlerts] = useState<Alert[]>(() => {
        const saved = localStorage.getItem(STORAGE_KEY);
        const data = saved ? JSON.parse(saved) : [];
        // Filter out any corrupted data on load
        return Array.isArray(data) ? data.filter(a => a && a.id && a.timestamp) : [];
    });

    // Persist to localStorage
    useEffect(() => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(alerts.slice(0, 50)));
    }, [alerts]);

    const addAlert = useCallback((newAlert: Omit<Alert, 'id' | 'timestamp' | 'isRead'>) => {
        const alert: Alert = {
            ...newAlert,
            id: Math.random().toString(36).substring(2, 9),
            timestamp: new Date().toISOString(),
            isRead: false,
        };
        setAlerts(prev => [alert, ...prev].slice(0, 50));
    }, []);

    const dismissAlert = useCallback((id: string) => {
        setAlerts(prev => prev.filter(a => a.id !== id));
    }, []);

    const markAsRead = useCallback((id: string) => {
        setAlerts(prev => prev.map(a => a.id === id ? { ...a, isRead: true } : a));
    }, []);

    const clearAll = useCallback(() => {
        setAlerts([]);
    }, []);

    return (
        <AlertContext.Provider value={{ alerts, addAlert, dismissAlert, markAsRead, clearAll }}>
            {children}
        </AlertContext.Provider>
    );
}
