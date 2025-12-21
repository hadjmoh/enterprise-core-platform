import { useEffect } from 'react';
import { Filter, XCircle, Globe, Bell, Maximize2, Layers, Copy, Shield, ExternalLink, Zap } from 'lucide-react';

interface DrillDownMenuProps {
    x: number;
    y: number;
    filters: Record<string, unknown>;
    onAction: (mode: 'include' | 'exclude' | 'pivot' | 'global' | 'alert' | 'timezoom' | 'copy' | 'forensic') => void;
    onClose: () => void;
}

const KeyCap = ({ children }: { children: React.ReactNode }) => (
    <kbd className="px-1.5 py-0.5 rounded bg-slate-800 border-b-2 border-slate-950 text-[9px] text-slate-500 font-bold group-hover:text-slate-300 transition-colors">
        {children}
    </kbd>
);

export const DrillDownMenu: React.FC<DrillDownMenuProps> = ({ x, y, filters, onAction, onClose }) => {
    // Keyboard Listeners for Speed-of-Thought Analysis
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === '+') onAction('include');
            if (e.key === '-') onAction('exclude');
            if (e.key === 'p' || e.key === 'P') onAction('pivot');
            if (e.key === 'z' || e.key === 'Z') onAction('timezoom');
            if (e.key === 'c' || e.key === 'C') onAction('copy');
            if (e.key === 'Escape') onClose();
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [onAction, onClose]);

    // Ensure menu stays within viewport
    const style: React.CSSProperties = {
        position: 'fixed',
        left: Math.min(x, window.innerWidth - 240),
        top: Math.min(y, window.innerHeight - 350),
        zIndex: 100,
    };

    const filterEntries = Object.entries(filters).filter(([k]) => !k.startsWith('_'));
    const isTimeDrill = '_time' in filters;

    // Smart Detection: IPs, Domains, etc.
    const isForensicMatch = filterEntries.some(([k, v]) => {
        const val = String(v);
        return k.toLowerCase().includes('ip') ||
            /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/.test(val) ||
            k.toLowerCase().includes('host') ||
            val.includes('.com') || val.includes('.org');
    });

    return (
        <>
            <button
                className="fixed inset-0 z-40 bg-transparent w-full h-full border-none outline-none cursor-default"
                onClick={onClose}
                aria-label="Close Analysis Menu"
                tabIndex={-1}
            />
            <div
                style={style}
                className="w-64 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-1 animate-in zoom-in-95 duration-150 backdrop-blur-xl ring-1 ring-white/10"
                role="menu"
                aria-label="Tactical Analysis Menu"
            >
                <div className="px-3 py-2 border-b border-slate-800/50 flex items-center justify-between">
                    <div>
                        <p className="text-[10px] font-bold text-brand-500 uppercase tracking-widest flex items-center gap-1.5">
                            <Zap className="h-2.5 w-2.5 fill-brand-500" />
                            Tactical Core
                        </p>
                        <div className="mt-1.5 space-y-1">
                            {filterEntries.map(([k, v]) => (
                                <p key={k} className="text-xs text-slate-200 truncate max-w-[180px]">
                                    {k}: <span className="text-brand-400 font-mono">"{String(v)}"</span>
                                </p>
                            ))}
                        </div>
                    </div>
                </div>

                <div className="p-1 space-y-0.5 mt-1">
                    <button
                        onClick={() => onAction('include')}
                        className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <Filter className="h-3.5 w-3.5 text-brand-400" />
                            <span>Include Filter</span>
                        </div>
                        <KeyCap>+</KeyCap>
                    </button>

                    <button
                        onClick={() => onAction('exclude')}
                        className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <XCircle className="h-3.5 w-3.5 text-red-400" />
                            <span>Exclude Filter</span>
                        </div>
                        <KeyCap>-</KeyCap>
                    </button>

                    {isTimeDrill && (
                        <button
                            onClick={() => onAction('timezoom')}
                            className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                        >
                            <div className="flex items-center gap-3">
                                <Maximize2 className="h-3.5 w-3.5 text-amber-400" />
                                <span>Zoom into Time</span>
                            </div>
                            <KeyCap>Z</KeyCap>
                        </button>
                    )}

                    <div className="h-[1px] bg-slate-800/50 my-1 mx-2" />

                    <button
                        onClick={() => onAction('pivot')}
                        className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <Layers className="h-3.5 w-3.5 text-indigo-400" />
                            <span>Pivot by Context</span>
                        </div>
                        <KeyCap>P</KeyCap>
                    </button>

                    <button
                        onClick={() => onAction('copy')}
                        className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <Copy className="h-3.5 w-3.5 text-emerald-400" />
                            <span>Copy SPL Filter</span>
                        </div>
                        <KeyCap>C</KeyCap>
                    </button>

                    {isForensicMatch && (
                        <button
                            onClick={() => onAction('forensic')}
                            className="w-full flex items-center gap-3 px-3 py-2 text-xs text-rose-400 bg-rose-500/5 hover:bg-rose-500/10 hover:text-rose-300 rounded-lg transition-colors border border-rose-500/10"
                        >
                            <Shield className="h-3.5 w-3.5" />
                            <span>Forensic Intel Lookup</span>
                            <ExternalLink className="h-2.5 w-2.5 opacity-50 ml-auto" />
                        </button>
                    )}

                    <button
                        onClick={() => onAction('alert')}
                        className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                    >
                        <Bell className="h-3.5 w-3.5 text-rose-400" />
                        <span>Create Alert for this</span>
                    </button>

                    <button
                        onClick={() => onAction('global')}
                        className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-brand-500/10 hover:text-brand-400 rounded-lg transition-colors border-t border-slate-800/50 mt-1 pt-2"
                    >
                        <Globe className="h-3.5 w-3.5 text-brand-400" />
                        Apply to Dashboard
                    </button>
                </div>
            </div>
        </>
    );
};
