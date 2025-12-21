import { useState, useEffect } from 'react';
import { useAlerts } from '../../contexts/AlertContextInstance';
import type { Alert } from '../../types/alert';
import { X, Trash2, Bell, AlertCircle, AlertTriangle, Info, Clock, CheckCircle2, Eraser } from 'lucide-react';
import { Button } from '../ui/Button';

interface AlertSidebarProps {
    isOpen: boolean;
    onClose: () => void;
}

export function AlertSidebar({ isOpen, onClose }: AlertSidebarProps) {
    const { alerts, clearAll, markAsRead, dismissAlert } = useAlerts();
    const [showUnreadOnly, setShowUnreadOnly] = useState(false);

    useEffect(() => {
        const handleEsc = (e: KeyboardEvent) => {
            if (e.key === 'Escape' && isOpen) onClose();
        };
        window.addEventListener('keydown', handleEsc);
        return () => window.removeEventListener('keydown', handleEsc);
    }, [isOpen, onClose]);

    const clearRead = () => {
        alerts.filter(a => a.isRead).forEach(a => dismissAlert(a.id));
    };

    const filteredAlerts = showUnreadOnly ? alerts.filter(a => !a.isRead) : alerts;
    const hasReadAlerts = alerts.some(a => a.isRead);

    const getIcon = (severity: Alert['severity']) => {
        switch (severity) {
            case 'critical': return <AlertCircle className="h-4 w-4 text-red-500" />;
            case 'warning': return <AlertTriangle className="h-4 w-4 text-amber-500" />;
            default: return <Info className="h-4 w-4 text-blue-500" />;
        }
    };

    return (
        <>
            {/* Backdrop */}
            {isOpen && (
                <button
                    className="fixed inset-0 bg-slate-950/40 backdrop-blur-sm z-40 transition-opacity duration-300 w-full h-full border-none outline-none cursor-default"
                    onClick={onClose}
                    aria-label="Close Alert Sidebar"
                    tabIndex={-1}
                />
            )}

            {/* Panel */}
            <div
                className={`fixed top-0 right-0 h-full w-80 sm:w-96 bg-slate-900 border-l border-slate-800 shadow-2xl z-50 transform transition-transform duration-500 ease-out ${isOpen ? 'translate-x-0' : 'translate-x-full'}`}
                role="dialog"
                aria-modal="true"
                aria-labelledby="alert-sidebar-title"
            >
                <div className="flex flex-col h-full">
                    {/* Header */}
                    <div className="flex items-center justify-between p-6 border-b border-slate-800 bg-slate-900/50">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-brand/10 rounded-lg">
                                <Bell className="h-5 w-5 text-brand" />
                            </div>
                            <div>
                                <h3 id="alert-sidebar-title" className="text-lg font-bold text-white">Alert Intelligence</h3>
                                <p className="text-xs text-slate-500 uppercase tracking-tighter">Active Incident HUD</p>
                            </div>
                        </div>
                        <Button variant="ghost" size="sm" onClick={onClose} className="h-8 w-8 p-0" title="Close Panel" aria-label="Close Panel">
                            <X className="h-5 w-5" />
                        </Button>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center justify-between px-6 py-3 bg-slate-800/20 border-b border-slate-800/50">
                        <div className="flex items-center gap-4">
                            <span className="text-xs font-medium text-slate-400">
                                {filteredAlerts.length} Incident{filteredAlerts.length !== 1 ? 's' : ''}
                            </span>
                            <div className="h-3 w-px bg-slate-800" />
                            <label className="flex items-center gap-2 cursor-pointer group">
                                <input
                                    type="checkbox"
                                    checked={showUnreadOnly}
                                    onChange={(e) => setShowUnreadOnly(e.target.checked)}
                                    className="sr-only"
                                />
                                <div className={`w-6 h-3 rounded-full transition-colors relative ${showUnreadOnly ? 'bg-brand' : 'bg-slate-700'} ring-offset-slate-900 focus-within:ring-2 focus-within:ring-brand focus-within:ring-offset-2`}>
                                    <div className={`absolute top-0.5 w-2 h-2 rounded-full bg-white transition-transform ${showUnreadOnly ? 'translate-x-3.5' : 'translate-x-0.5'}`} />
                                </div>
                                <span className={`text-[10px] font-bold uppercase transition-colors ${showUnreadOnly ? 'text-brand' : 'text-slate-500 group-hover:text-slate-400'}`}>Unread</span>
                            </label>
                        </div>
                        {alerts.length > 0 && (
                            <div className="flex gap-2">
                                {hasReadAlerts && (
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={clearRead}
                                        title="Clear Read"
                                        aria-label="Clear Read Alerts"
                                        className="text-xs h-7 w-7 p-0 text-slate-500 hover:text-amber-400"
                                    >
                                        <Eraser className="h-3.5 w-3.5" />
                                    </Button>
                                )}
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={clearAll}
                                    title="Clear All"
                                    aria-label="Clear All Alerts"
                                    className="text-xs h-7 w-7 p-0 text-slate-500 hover:text-red-400"
                                >
                                    <Trash2 className="h-3.5 w-3.5" />
                                </Button>
                            </div>
                        )}
                    </div>

                    {/* Content */}
                    <div className="flex-1 overflow-y-auto p-4 space-y-3">
                        {filteredAlerts.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-48 text-center space-y-3 opacity-40">
                                <CheckCircle2 className="h-8 w-8 text-emerald-500" />
                                <p className="text-xs font-bold text-white uppercase tracking-tighter">
                                    {showUnreadOnly ? 'No New Incidents' : 'Zero Incidents'}
                                </p>
                            </div>
                        ) : (
                            filteredAlerts.map((alert: Alert) => (
                                <button
                                    key={alert.id}
                                    className={`w-full text-left group relative p-4 rounded-xl border transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-brand focus:ring-offset-2 focus:ring-offset-slate-900 ${alert.isRead ? 'bg-slate-900/30 border-slate-800/50' : 'bg-slate-800/40 border-slate-700 shadow-lg shadow-black/20'}`}
                                    onClick={() => markAsRead(alert.id)}
                                    aria-label={`Mark alert ${alert.title} as read`}
                                >
                                    <div className="flex items-start gap-3">
                                        <div className="mt-1">{getIcon(alert.severity)}</div>
                                        <div className="flex-1 space-y-1">
                                            <div className="flex items-center justify-between">
                                                <h4 className={`text-sm font-bold uppercase tracking-tight ${alert.isRead ? 'text-slate-400' : 'text-white'}`}>
                                                    {alert.title}
                                                </h4>
                                                {!alert.isRead && (
                                                    <span className="h-2 w-2 rounded-full bg-brand shadow-[0_0_8px_rgba(var(--brand-rgb),0.6)]" />
                                                )}
                                            </div>
                                            <p className={`text-xs leading-relaxed ${alert.isRead ? 'text-slate-500' : 'text-slate-300'}`}>
                                                {alert.message}
                                            </p>
                                            <div className="flex items-center gap-2 pt-2 text-[10px] text-slate-500 font-mono">
                                                <Clock className="h-3 w-3" />
                                                {new Date(alert.timestamp).toLocaleTimeString()}
                                            </div>
                                            {alert.action && (
                                                <div className="pt-3">
                                                    <button
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            window.location.href = alert.action!.url;
                                                        }}
                                                        className="px-3 py-1.5 rounded-lg bg-slate-800 text-white text-[10px] font-bold uppercase tracking-wider hover:bg-slate-700 transition-colors"
                                                    >
                                                        {alert.action.label}
                                                    </button>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                    <button
                                        className="absolute top-2 right-2 p-1 opacity-0 group-hover:opacity-100 text-slate-500 hover:text-red-400 transition-all"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            dismissAlert(alert.id);
                                        }}
                                    >
                                        <X className="h-3 w-3" />
                                    </button>
                                </button>
                            ))
                        )}
                    </div>

                    {/* Footer */}
                    <div className="p-6 border-t border-slate-800 bg-slate-900/50 mt-auto">
                        <Button variant="outline" className="w-full text-xs py-5 border-slate-700 hover:bg-slate-800">
                            Configure Alert Thresholds
                        </Button>
                    </div>
                </div>
            </div>
        </>
    );
}
