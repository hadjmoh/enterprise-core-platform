export type AlertSeverity = 'critical' | 'warning' | 'info';

export interface Alert {
    id: string;
    title: string;
    message: string;
    timestamp: string;
    severity: AlertSeverity;
    isRead: boolean;
    action?: {
        label: string;
        url: string;
    };
}

export interface AlertContextType {
    alerts: Alert[];
    addAlert: (alert: Omit<Alert, 'id' | 'timestamp' | 'isRead'>) => void;
    dismissAlert: (id: string) => void;
    markAsRead: (id: string) => void;
    clearAll: () => void;
}
