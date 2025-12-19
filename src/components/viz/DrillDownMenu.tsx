import { Filter, XCircle, Globe, Bell, Maximize2, Layers } from 'lucide-react';

interface DrillDownMenuProps {
    x: number;
    y: number;
    filters: Record<string, unknown>;
    onAction: (mode: 'include' | 'exclude' | 'pivot' | 'global' | 'alert' | 'timezoom') => void;
    onClose: () => void;
}

export const DrillDownMenu: React.FC<DrillDownMenuProps> = ({ x, y, filters, onAction, onClose }) => {
    // Ensure menu stays within viewport
    const style: React.CSSProperties = {
        position: 'fixed',
        left: Math.min(x, window.innerWidth - 240),
        top: Math.min(y, window.innerHeight - 300),
        zIndex: 100,
    };

    const filterEntries = Object.entries(filters).filter(([k]) => !k.startsWith('_'));
    const isTimeDrill = '_time' in filters;

    return (
        <>
            <div
                className="fixed inset-0 z-40 bg-transparent"
                onClick={onClose}
            />
            <div
                style={style}
                className="w-64 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-1 animate-in zoom-in-95 duration-150 backdrop-blur-xl"
            >
                <div className="px-3 py-2 border-b border-slate-800/50">
                    <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Tactical Actions</p>
                    <div className="mt-1.5 space-y-1">
                        {filterEntries.map(([k, v]) => (
                            <p key={k} className="text-xs text-slate-200 truncate">
                                {k}: <span className="text-brand-400 font-mono">"{String(v)}"</span>
                            </p>
                        ))}
                        {isTimeDrill && (
                            <p className="text-[10px] text-slate-500 font-mono">
                                Time: {new Date(filters._time as string).toLocaleTimeString()}
                            </p>
                        )}
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
                        <span className="text-[9px] text-slate-600 font-bold group-hover:text-slate-400">+</span>
                    </button>

                    <button
                        onClick={() => onAction('exclude')}
                        className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors group"
                    >
                        <div className="flex items-center gap-3">
                            <XCircle className="h-3.5 w-3.5 text-red-400" />
                            <span>Exclude Filter</span>
                        </div>
                        <span className="text-[9px] text-slate-600 font-bold group-hover:text-slate-400">-</span>
                    </button>

                    {isTimeDrill && (
                        <button
                            onClick={() => onAction('timezoom')}
                            className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                        >
                            <Maximize2 className="h-3.5 w-3.5 text-amber-400" />
                            <span>Zoom into Time Range</span>
                        </button>
                    )}

                    <div className="h-[1px] bg-slate-800/50 my-1 mx-2" />

                    <button
                        onClick={() => onAction('pivot')}
                        className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                    >
                        <Layers className="h-3.5 w-3.5 text-indigo-400" />
                        <span>Pivot by Context</span>
                    </button>

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

