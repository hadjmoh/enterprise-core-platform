import { Search, Bell, Menu, Clock } from 'lucide-react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { useAlerts } from '../../contexts/AlertContextInstance';
import { useUI } from '../../hooks/useUI';

interface HeaderProps {
    toggleSidebar: () => void;
    openAlerts: () => void;
}

export function Header({ toggleSidebar, openAlerts }: HeaderProps) {
    const { alerts } = useAlerts();
    const { timezone, setTimezone } = useUI();
    const unreadCount = alerts.filter(a => !a.isRead).length;

    return (
        <header className="sticky top-0 z-30 flex h-16 w-full items-center justify-between border-b border-slate-800 bg-slate-950/50 px-4 backdrop-blur-xl transition-all">
            <div className="flex items-center gap-4">
                <Button variant="ghost" size="sm" className="lg:hidden" onClick={toggleSidebar}>
                    <Menu className="h-5 w-5" />
                </Button>
                <div className="hidden md:block w-96">
                    <Input
                        placeholder="Search resources, logs, or settings..."
                        icon={<Search className="h-4 w-4" />}
                    />
                </div>
            </div>

            <div className="flex items-center gap-4">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setTimezone(timezone === 'UTC' ? 'Local' : 'UTC')}
                    className="flex items-center gap-2 text-xs font-mono text-slate-400 hover:text-brand"
                >
                    <Clock className="h-4 w-4" />
                    {timezone}
                </Button>

                <Button variant="ghost" size="sm" className="relative" onClick={openAlerts}>
                    <Bell className="h-5 w-5 text-slate-400" />
                    {unreadCount > 0 && (
                        <span className="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-lg ring-2 ring-slate-950">
                            <span className="absolute inset-0 rounded-full bg-red-500 animate-ping opacity-75" />
                            <span className="relative">{unreadCount > 9 ? '9+' : unreadCount}</span>
                        </span>
                    )}
                </Button>
                <div className="flex items-center gap-3 pl-4 border-l border-slate-800">
                    <div className="text-right hidden md:block">
                        <p className="text-sm font-medium text-white">Admin User</p>
                        <p className="text-xs text-slate-400">System Administrator</p>
                    </div>
                    <div className="h-9 w-9 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-sm font-bold text-slate-300">
                        AD
                    </div>
                </div>
            </div>
        </header>
    );
}
