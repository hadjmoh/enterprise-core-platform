import { Filter, XCircle, Search, Globe } from 'lucide-react';

interface DrillDownMenuProps {
    x: number;
    y: number;
    filters: Record<string, unknown>;
    onAction: (mode: 'include' | 'exclude' | 'pivot' | 'global') => void;
    onClose: () => void;
}

export const DrillDownMenu: React.FC<DrillDownMenuProps> = ({ x, y, filters, onAction, onClose }) => {
    // Ensure menu stays within viewport
    const style: React.CSSProperties = {
        position: 'fixed',
        left: Math.min(x, window.innerWidth - 240),
        top: Math.min(y, window.innerHeight - 300),
        zIndex: 50,
    };

    const firstField = Object.keys(filters)[0];
    const firstValue = filters[firstField];

    const includeSpl = `| search ${firstField}="${firstValue}"`;
    const excludeSpl = `| search ${firstField}!="${firstValue}"`;

    return (
        <>
            <div
                className="fixed inset-0 z-40 bg-transparent"
                onClick={onClose}
            />
            <div
                style={style}
                className="w-60 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-1 animate-in zoom-in-95 duration-150"
            >
                <div className="px-3 py-2 border-b border-slate-800/50">
                    <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Context Action</p>
                    <p className="text-xs text-slate-200 mt-1 truncate">
                        {firstField}: <span className="text-brand-400">"{String(firstValue)}"</span>
                    </p>
                </div>

                <div className="p-1 space-y-0.5 mt-1">
                    <div className="group relative">
                        <button
                            onClick={() => onAction('include')}
                            className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                        >
                            <div className="flex items-center gap-3">
                                <Filter className="h-3.5 w-3.5 text-brand-400" />
                                <span>Include Filter</span>
                            </div>
                        </button>
                        <div className="hidden group-hover:block absolute left-full ml-2 top-0 w-48 p-2 bg-slate-800 rounded-lg text-[10px] text-slate-400 font-mono border border-slate-700 shadow-xl z-50">
                            {includeSpl}
                        </div>
                    </div>

                    <div className="group relative">
                        <button
                            onClick={() => onAction('exclude')}
                            className="w-full flex items-center justify-between px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                        >
                            <div className="flex items-center gap-3">
                                <XCircle className="h-3.5 w-3.5 text-red-400" />
                                <span>Exclude Filter</span>
                            </div>
                        </button>
                        <div className="hidden group-hover:block absolute left-full ml-2 top-0 w-48 p-2 bg-slate-800 rounded-lg text-[10px] text-slate-400 font-mono border border-slate-700 shadow-xl z-50">
                            {excludeSpl}
                        </div>
                    </div>

                    <button
                        onClick={() => onAction('global')}
                        className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-brand-500/10 hover:text-brand-400 rounded-lg transition-colors border-t border-slate-800/50 mt-1 pt-2"
                    >
                        <Globe className="h-3.5 w-3.5 text-brand-400" />
                        Apply to Dashboard
                    </button>

                    <button
                        onClick={() => onAction('pivot')}
                        className="w-full flex items-center gap-3 px-3 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors"
                    >
                        <Search className="h-3.5 w-3.5 text-indigo-400" />
                        Pivot to Raw Logs
                    </button>
                </div>
            </div>
        </>
    );
};

