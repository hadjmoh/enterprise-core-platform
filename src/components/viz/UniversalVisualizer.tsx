import React from 'react';
import type { SearchProgress } from '../../hooks/useSearch';
import { AlertCircle, Loader2, Table, LineChart, Hash, PieChart, BarChart3 } from 'lucide-react';
import { TimeSeriesChart } from './TimeSeriesChart';
import { SingleValuePanel } from './SingleValuePanel';
import { DistributionChart } from './DistributionChart';

interface UniversalVisualizerProps {
    progress: SearchProgress;
    onDrillDown?: (filters: Record<string, unknown>, x: number, y: number) => void;
}

export const UniversalVisualizer: React.FC<UniversalVisualizerProps> = ({ progress, onDrillDown }) => {
    const { status, metadata, results, error } = progress;

    if (status === 'idle') {
        return (
            <div className="flex flex-col items-center justify-center p-12 bg-slate-900/50 rounded-xl border border-slate-800 border-dashed">
                <Loader2 className="w-8 h-8 text-slate-700 animate-spin mb-4" />
                <p className="text-slate-500 font-medium">Waiting for search...</p>
            </div>
        );
    }

    if (status === 'error' && error?.fatal) {
        return (
            <div className="p-6 bg-red-900/20 border border-red-900/50 rounded-xl">
                <div className="flex items-center gap-3 text-red-500 mb-2">
                    <AlertCircle className="w-5 h-5" />
                    <h3 className="font-bold">Search Error</h3>
                </div>
                <p className="text-red-400 text-sm font-mono">{error.message}</p>
            </div>
        );
    }

    // Smart Selection Logic
    const renderVisualization = () => {
        if (!metadata) return null;

        if (results.length === 0 && status === 'complete') {
            return (
                <div className="flex flex-col items-center justify-center p-12 bg-slate-900/30 rounded-xl border border-slate-800 border-dashed group/empty">
                    <div className="p-4 bg-slate-800/50 rounded-full mb-4 group-hover/empty:scale-110 transition-transform duration-500">
                        <Table className="w-8 h-8 text-slate-600" />
                    </div>
                    <p className="text-slate-400 font-bold uppercase tracking-widest text-xs">No Events Match Query</p>
                    <p className="text-slate-600 text-[10px] mt-2 max-w-xs text-center">Try adjusting your time range or filters to broaden your search results.</p>
                </div>
            );
        }

        switch (metadata._type) {
            case 'timechart':
                return (
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2 text-indigo-400">
                                <LineChart className="w-4 h-4" />
                                <span className="text-sm font-semibold uppercase tracking-wider">Time Series Analysis</span>
                                {results.length >= 500 && (
                                    <div className="ml-2 px-2 py-0.5 rounded bg-amber-500/20 border border-amber-500/30 text-[9px] font-bold text-amber-500 flex items-center gap-1">
                                        <AlertCircle className="w-2.5 h-2.5" />
                                        MAX SERIES REACHED
                                    </div>
                                )}
                            </div>
                            <span className="text-xs text-slate-500">Span: {metadata._span}s</span>
                        </div>
                        <div className="h-64 bg-slate-800/20 rounded-lg border border-slate-800 p-4 shadow-inner">
                            <TimeSeriesChart
                                data={results}
                                span={metadata._span}
                                onDrillDown={onDrillDown}
                                thresholds={[
                                    { value: 80, color: 'rgba(239, 68, 68, 0.5)', label: 'SLA BREACH' }
                                ]}
                            />
                        </div>
                    </div>
                );

            case 'stats':
            case 'top':
            case 'rare': {
                // Heuristic for chart type
                const row = results[0];
                const keys = row ? Object.keys(row).filter(k => !k.startsWith('_')) : [];
                const numericKeys = keys.filter(k => typeof row[k] === 'number');
                const categoricalKeys = keys.filter(k => typeof row[k] === 'string');

                // Single Value (e.g. stats count)
                if (results.length === 1 && numericKeys.length === 1 && categoricalKeys.length === 0) {
                    return (
                        <div className="space-y-4">
                            <div className="flex items-center gap-2 text-indigo-400">
                                <Hash className="w-4 h-4" />
                                <span className="text-sm font-semibold uppercase tracking-wider">KPI Metric</span>
                            </div>
                            <SingleValuePanel
                                value={row[numericKeys[0]] as number}
                                label={numericKeys[0]}
                                thresholds={[
                                    { value: 0, color: '#10b981', label: 'Healthy', severity: 'success' },
                                    { value: 50, color: '#fbbf24', label: 'Warning', severity: 'warning' },
                                    { value: 100, color: '#ef4444', label: 'Critical', severity: 'critical' }
                                ]}
                            />
                        </div>
                    );
                }

                // Distribution (e.g. stats count by host)
                if (results.length > 1 && categoricalKeys.length === 1 && numericKeys.length >= 1) {
                    const vizType = metadata._type === 'top' ? 'pie' : 'bar';
                    return (
                        <div className="space-y-4">
                            <div className="flex items-center gap-2 text-indigo-400">
                                {vizType === 'pie' ? <PieChart className="w-4 h-4" /> : <BarChart3 className="w-4 h-4" />}
                                <span className="text-sm font-semibold uppercase tracking-wider">
                                    {metadata._type === 'top' ? 'Top N Distribution' : 'Categorical Analysis'}
                                </span>
                                {results.length >= 500 && (
                                    <div className="ml-2 px-2 py-0.5 rounded bg-amber-500/20 border border-amber-500/30 text-[9px] font-bold text-amber-500 flex items-center gap-1">
                                        <AlertCircle className="w-2.5 h-2.5" />
                                        MAX SERIES REACHED
                                    </div>
                                )}
                            </div>
                            <div className="h-64 bg-slate-800/20 rounded-lg border border-slate-800 p-4">
                                <DistributionChart
                                    data={results}
                                    type={vizType}
                                    onDrillDown={onDrillDown}
                                    thresholds={[
                                        { value: 500, color: '#fbbf24' },
                                        { value: 1000, color: '#ef4444' }
                                    ]}
                                />
                            </div>
                        </div>
                    );
                }

                // Default to table
                return (
                    <div className="space-y-4">
                        <div className="flex items-center gap-2 text-emerald-400">
                            <Table className="w-4 h-4" />
                            <span className="text-sm font-semibold uppercase tracking-wider">Aggregation Results</span>
                        </div>
                        <div className="overflow-x-auto rounded-lg border border-slate-800 bg-slate-900/30">
                            <table className="w-full text-left text-sm border-collapse">
                                <thead className="bg-slate-800/80 text-slate-300 font-medium">
                                    <tr>
                                        {results[0] && Object.keys(results[0]).map(key => (
                                            <th key={key} className="px-4 py-2 capitalize">{key.replace('_', ' ')}</th>
                                        ))}
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-800 text-slate-400">
                                    {results.slice(0, 10).map((row: Record<string, unknown>, i: number) => (
                                        <tr
                                            key={i}
                                            className="hover:bg-slate-800/40 transition-colors cursor-pointer focus:outline-none focus:bg-slate-800 focus:ring-1 focus:ring-brand-500/50"
                                            onClick={(e) => onDrillDown?.(row, e.clientX, e.clientY)}
                                            role="button"
                                            tabIndex={0}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter' || e.key === ' ') {
                                                    e.preventDefault();
                                                    const rect = e.currentTarget.getBoundingClientRect();
                                                    onDrillDown?.(row, rect.left + rect.width / 2, rect.top + rect.height / 2);
                                                }
                                            }}
                                            aria-label="View drill-down options for this row"
                                        >
                                            {Object.entries(row).map(([key, val], j: number) => {
                                                const sVal = String(val);
                                                const cellStyle = "px-4 py-2 font-mono text-xs";
                                                let pillStyle = "";

                                                // Automatic Semantic Intelligence
                                                if (key === 'status') {
                                                    const code = parseInt(sVal);
                                                    if (code >= 500) pillStyle = "bg-red-500/10 text-red-500 border border-red-500/20 px-1.5 py-0.5 rounded text-[10px] font-bold";
                                                    else if (code >= 400) pillStyle = "bg-amber-500/10 text-amber-500 border border-amber-500/20 px-1.5 py-0.5 rounded text-[10px] font-bold";
                                                    else if (code >= 200) pillStyle = "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 px-1.5 py-0.5 rounded text-[10px] font-bold";
                                                } else if (key === 'latency' || key === 'duration') {
                                                    const lat = parseFloat(sVal);
                                                    if (lat > 1000) pillStyle = "text-red-400 font-bold underline decoration-red-500/50 underline-offset-4";
                                                    else if (lat > 500) pillStyle = "text-amber-400 font-bold";
                                                } else if (key.toLowerCase().includes('level')) {
                                                    if (sVal === 'ERROR') pillStyle = "text-red-500 font-black";
                                                    if (sVal === 'WARN') pillStyle = "text-amber-500 font-bold";
                                                }

                                                // Fuzzy Anomaly Detection (Visual Forensics)
                                                const lowerVal = sVal.toLowerCase();
                                                const isAnomaly = lowerVal.includes('error') ||
                                                    lowerVal.includes('fail') ||
                                                    lowerVal.includes('deny') ||
                                                    lowerVal.includes('critical') ||
                                                    lowerVal.includes('exception') ||
                                                    lowerVal.includes('unauthorized');

                                                if (isAnomaly && !pillStyle) {
                                                    pillStyle = "text-red-400 font-bold border-b border-red-500/50";
                                                }

                                                return (
                                                    <td key={j} className={cellStyle}>
                                                        <span className={pillStyle}>{sVal}</span>
                                                    </td>
                                                );
                                            })}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                            {results.length > 10 && (
                                <div className="p-2 text-center text-xs text-slate-500 bg-slate-900/50 border-t border-slate-800">
                                    Showing first 10 of {results.length} rows
                                </div>
                            )}
                        </div>
                    </div>
                );
            }

            default:
                return (
                    <div className="p-4 bg-slate-800/20 rounded-lg border border-slate-800">
                        <p className="text-slate-400 text-xs font-mono">
                            Visualizer for `{metadata._type}` not yet implemented. Falling back to JSON.
                        </p>
                        <pre className="mt-4 text-[10px] overflow-auto max-h-40 text-slate-500">
                            {JSON.stringify(results.slice(0, 3), null, 2)}
                        </pre>
                    </div>
                );
        }
    };

    return (
        <div className="bg-slate-900/80 backdrop-blur-xl p-6 rounded-2xl border border-slate-800 shadow-2xl">
            {renderVisualization()}

            {/* Status Footer */}
            <div className="mt-6 pt-4 border-t border-slate-800 flex items-center justify-between">
                <div className="flex items-center gap-4 text-[10px] font-bold tracking-widest uppercase">
                    <div className={`flex items-center gap-1.5 ${status === 'streaming' ? 'text-emerald-500' : 'text-slate-500'}`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${status === 'streaming' ? 'bg-emerald-500 animate-pulse ring-4 ring-emerald-500/20' : 'bg-slate-700'}`} />
                        {status === 'streaming' ? 'LIVE STREAMING' : status}
                    </div>
                    <div className="text-slate-500">
                        Events: <span className="text-slate-300">{progress.totalCount}</span>
                    </div>
                </div>
                {status === 'streaming' && (
                    <Loader2 className="w-3 h-3 text-slate-600 animate-spin" />
                )}
            </div>
        </div>
    );
};
