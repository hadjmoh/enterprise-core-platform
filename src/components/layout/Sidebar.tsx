import { cn } from '../../lib/utils';
import { Button } from '../ui/Button';
import { NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import {
    LayoutDashboard,
    ShieldCheck,
    Users,
    Server,
    Settings,
    ShieldAlert,
    LogOut,
    ChevronLeft,
    ChevronRight,
    Database,
    FileText,
    Globe,
    Layout,
    Zap,
    Cloud,
    Fingerprint,
    Network,
    FlaskConical,
    type LucideIcon
} from 'lucide-react';

interface SidebarProps {
    isOpen: boolean;
    setIsOpen: (value: boolean) => void;
    isMobile: boolean;
}

const NavItem = ({ icon: Icon, label, to, isOpen, isMobile }: { icon: LucideIcon, label: string, to: string, isOpen: boolean, isMobile: boolean }) => (
    <NavLink
        to={to}
        className={({ isActive }) => cn(
            "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
            isActive
                ? "bg-brand-500/10 text-brand-400"
                : "text-slate-400 hover:bg-slate-800 hover:text-slate-100"
        )}
    >
        <Icon className="h-5 w-5" />
        {(isOpen || isMobile) && <span>{label}</span>}
    </NavLink>
);

export function Sidebar({ isOpen, setIsOpen, isMobile }: SidebarProps) {
    const { logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <aside
            className={cn(
                "fixed left-0 top-0 z-40 h-screen border-r border-slate-800 bg-slate-950 transition-all duration-300",
                isOpen ? "w-64" : "w-20",
                isMobile && !isOpen && "-translate-x-full"
            )}
        >
            <div className="flex h-16 items-center justify-between px-4 border-b border-slate-800">
                <div className="flex items-center gap-2 font-bold text-white">
                    <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-brand-500 to-cyan-400 p-[1px]">
                        <div className="h-full w-full rounded-lg bg-slate-950 flex items-center justify-center">
                            <span className="text-brand-400">E</span>
                        </div>
                    </div>
                    {(isOpen || isMobile) && <span className="bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">Enterprise</span>}
                </div>
                {!isMobile && (
                    <Button
                        variant="ghost"
                        size="sm"
                        className="hidden lg:flex"
                        onClick={() => setIsOpen(!isOpen)}
                    >
                        {isOpen ? <ChevronLeft className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                    </Button>
                )}
            </div>

            <div className="flex h-[calc(100vh-64px)] flex-col justify-between p-4">
                <nav className="space-y-1">
                    <NavItem icon={LayoutDashboard} label="Dashboard" to="/" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Users} label="User Management" to="/users" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Server} label="Infrastructure" to="/infrastructure" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Database} label="Data Inputs" to="/data-inputs" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={ShieldAlert} label="Security" to="/security" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Layout} label="Security Posture" to="/security/dashboard" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={ShieldAlert} label="Incident Review" to="/security/incidents" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Globe} label="Threat Intelligence" to="/threat-intel" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Zap} label="Detection Rules" to="/security/rules" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Cloud} label="Cloud Assets" to="/cloud/assets" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Fingerprint} label="Trust Graph" to="/trust-graph" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={Network} label="Cluster Management" to="/cluster" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={FlaskConical} label="Simulation Lab" to="/simulation-lab" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={ShieldCheck} label="Control Center" to="/control-center" isOpen={isOpen} isMobile={isMobile} />
                    <NavItem icon={FileText} label="Reporting Hub" to="/reports" isOpen={isOpen} isMobile={isMobile} />
                </nav>

                <div className="space-y-1">
                    <NavItem icon={Settings} label="Settings" to="/settings" isOpen={isOpen} isMobile={isMobile} />
                    <button
                        onClick={handleLogout}
                        className={cn(
                            "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors text-slate-400 hover:bg-red-500/10 hover:text-red-400"
                        )}
                    >
                        <LogOut className="h-5 w-5" />
                        {(isOpen || isMobile) && <span>Logout</span>}
                    </button>
                </div>
            </div>
        </aside>
    );
}
