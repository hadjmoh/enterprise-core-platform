import { useEffect } from 'react';
import { useAlerts } from '../../contexts/AlertContextInstance';
import type { Alert } from '../../types/alert';
import { AlertCircle, X, Info, AlertTriangle } from 'lucide-react';

export function AlertToastContainer() {
    const { alerts, dismissAlert } = useAlerts();

    // Show top 3 unread alerts
    const visibleAlerts = alerts.filter((a: Alert) => !a.isRead).slice(0, 3);

    return (
        <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-3 pointer-events-none">
            {visibleAlerts.map((alert: Alert, index: number) => (
                <div
                    key={alert.id}
                    className="pointer-events-auto"
                    style={{
                        opacity: 1 - index * 0.2,
                        transform: `scale(${1 - index * 0.05})`,
                        zIndex: 100 - index
                    }}
                >
                    <AlertCard alert={alert} onDismiss={() => dismissAlert(alert.id)} />
                </div>
            ))}
        </div>
    );
}

function AlertCard({ alert, onDismiss }: { alert: Alert, onDismiss: () => void }) {
    useEffect(() => {
        const timer = setTimeout(onDismiss, 6000);
        return () => clearTimeout(timer);
    }, [onDismiss]);

    const getIcon = () => {
        switch (alert.severity) {
            case 'critical': return <AlertCircle className="h-5 w-5 text-red-400" />;
            case 'warning': return <AlertTriangle className="h-5 w-5 text-amber-400" />;
            default: return <Info className="h-5 w-5 text-blue-400" />;
        }
    };

    const getStyles = () => {
        switch (alert.severity) {
            case 'critical': return 'border-red-500/30 bg-red-950/40 shadow-red-500/10';
            case 'warning': return 'border-amber-500/30 bg-amber-950/40 shadow-amber-500/10';
            default: return 'border-blue-500/30 bg-blue-950/40 shadow-blue-500/10';
        }
    };

    return (
        <div className={`w-80 flex items-start gap-3 p-4 rounded-xl border backdrop-blur-xl shadow-2xl animate-in fade-in slide-in-from-bottom-5 duration-500 ${getStyles()}`}>
            <div className="flex-shrink-0 pt-0.5">
                {getIcon()}
            </div>
            <div className="flex-1 min-w-0 pr-2">
                <h4 className="text-xs font-black text-white uppercase tracking-widest">{alert.title}</h4>
                <p className="text-xs text-slate-300 mt-1.5 leading-relaxed">{alert.message}</p>
                <p className="text-[10px] text-slate-500 mt-2 font-mono">
                    {new Date(alert.timestamp).toLocaleTimeString()}
                </p>
                {alert.action && (
                    <div className="mt-3">
                        <button
                            onClick={() => window.location.href = alert.action!.url}
                            className={`px-3 py-1.5 rounded-lg text-[10px] font-bold uppercase tracking-wider transition-colors ${alert.severity === 'critical'
                                    ? 'bg-red-500 text-white hover:bg-red-400'
                                    : 'bg-brand text-white hover:bg-brand/80'
                                }`}
                        >
                            {alert.action.label}
                        </button>
                    </div>
                )}
            </div>
            <button
                onClick={onDismiss}
                className="text-slate-500 hover:text-white transition-colors"
            >
                <X className="h-4 w-4" />
            </button>
        </div>
    );
}
