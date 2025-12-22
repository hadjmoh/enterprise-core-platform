import { Button } from '../ui/Button';
import { AlertTriangle, XCircle, Info, CheckCircle, Database, Cpu, List } from 'lucide-react';
import type { CostEstimate } from '../../services/searchService';

interface CostPreviewProps {
    estimate: CostEstimate;
    allowed: boolean;
    reason: string;
    onConfirm: () => void;
    onCancel: () => void;
}

export const CostPreview: React.FC<CostPreviewProps> = ({ estimate, allowed, reason, onConfirm, onCancel }) => {
    const riskColor = {
        low: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20',
        medium: 'text-amber-400 bg-amber-500/10 border-amber-500/20',
        high: 'text-orange-400 bg-orange-500/10 border-orange-500/20',
        critical: 'text-red-400 bg-red-500/10 border-red-500/20',
    }[estimate.risk_level] || 'text-slate-400';

    const riskIcon = {
        low: <CheckCircle className="w-5 h-5" />,
        medium: <Info className="w-5 h-5" />,
        high: <AlertTriangle className="w-5 h-5" />,
        critical: <XCircle className="w-5 h-5" />,
    }[estimate.risk_level];

    return (
        <div className="absolute top-full left-0 right-0 mt-2 p-4 bg-slate-900 border border-slate-700 rounded-lg shadow-2xl z-50 animate-in fade-in slide-in-from-top-2">
            <div className={`flex items-start gap-3 p-3 rounded-lg border mb-4 ${riskColor}`}>
                <div className="mt-0.5">{riskIcon}</div>
                <div>
                    <h4 className="font-bold flex items-center gap-2">
                        {estimate.risk_level.toUpperCase()} Risk Query
                    </h4>
                    <p className="text-sm opacity-90 mt-1">{reason}</p>
                </div>
            </div>

            <div className="grid grid-cols-3 gap-4 mb-4 text-sm text-slate-300">
                <div className="flex flex-col gap-1 p-2 bg-slate-800/50 rounded">
                    <span className="text-xs text-slate-500 flex items-center gap-1">
                        <Database className="w-3 h-3" /> Data Scan
                    </span>
                    <span className="font-mono">{estimate.estimated_scan_gb.toFixed(2)} GB</span>
                </div>
                <div className="flex flex-col gap-1 p-2 bg-slate-800/50 rounded">
                    <span className="text-xs text-slate-500 flex items-center gap-1">
                        <List className="w-3 h-3" /> Rows
                    </span>
                    <span className="font-mono">{(estimate.estimated_rows / 1000000).toFixed(1)} M</span>
                </div>
                <div className="flex flex-col gap-1 p-2 bg-slate-800/50 rounded">
                    <span className="text-xs text-slate-500 flex items-center gap-1">
                        <Cpu className="w-3 h-3" /> CPU Est.
                    </span>
                    <span className="font-mono">{estimate.estimated_cpu_seconds.toFixed(0)} s</span>
                </div>
            </div>

            {estimate.risk_factors.length > 0 && (
                <div className="mb-4">
                    <p className="text-xs font-semibold text-slate-500 mb-2">RISK FACTORS</p>
                    <ul className="space-y-1">
                        {estimate.risk_factors.map((factor, i) => (
                            <li key={i} className="text-xs text-slate-400 flex items-center gap-2">
                                <span className="w-1 h-1 rounded-full bg-slate-500" />
                                {factor}
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            <div className="flex justify-end gap-2 pt-2 border-t border-slate-800">
                <Button variant="ghost" size="sm" onClick={onCancel}>
                    Cancel
                </Button>
                <Button
                    variant={allowed ? "primary" : "secondary"} // Secondary (gray) if blocked
                    size="sm"
                    onClick={onConfirm}
                    disabled={!allowed}
                    className={!allowed ? "opacity-50 cursor-not-allowed" : ""}
                >
                    {allowed ? "Run Query" : "Blocked by Policy"}
                </Button>
            </div>
        </div>
    );
};
