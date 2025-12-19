import { useEffect, useState } from 'react';
import { Card } from '../ui/Card';
import { X, GripVertical, RefreshCw, FileText, Database, Code, ChevronLeft, PlusCircle } from 'lucide-react';
import { Button } from '../ui/Button';
import { UniversalVisualizer } from '../viz/UniversalVisualizer';
import { useSearch } from '../../hooks/useSearch';
import { Badge } from '../ui/Badge';
import { Skeleton } from '../ui/Skeleton';
import { WidgetErrorBoundary } from './WidgetErrorBoundary';
import { exportToCSV, exportToPDF, exportToJSON } from '../../services/exportService';
import { DrillDownMenu } from '../viz/DrillDownMenu';

interface WidgetContainerProps {
    id: string;
    title: string;
    spl: string;
    isLocked: boolean;
    onRemove: (id: string) => void;
    onGlobalSearch?: (spl: string) => void;
    onAddWidget?: (widget: { title: string, type: string, spl: string }) => void;
}

export function WidgetContainer({ id, title, spl, isLocked, onRemove, onGlobalSearch, onAddWidget }: WidgetContainerProps) {
    const search = useSearch();
    const [history, setHistory] = useState<string[]>([spl]);
    const [activeDrillDown, setActiveDrillDown] = useState<{ x: number, y: number, filters: Record<string, unknown> } | null>(null);
    const currentSpl = history[history.length - 1];

    useEffect(() => {
        search.execute(currentSpl);
    }, [currentSpl, search]);

    const handleDrillDownEvent = (filters: Record<string, unknown>, x: number, y: number) => {
        setActiveDrillDown({ x, y, filters });
    };

    const executeDrillDown = (mode: 'include' | 'exclude' | 'pivot' | 'global' | 'alert' | 'timezoom' | 'copy' | 'forensic') => {
        if (!activeDrillDown) return;

        const { filters } = activeDrillDown;
        let refinements = '';

        if (mode === 'include') {
            refinements = Object.entries(filters)
                .filter(([k]) => !k.startsWith('_'))
                .map(([k, v]) => `| search ${k}="${v}"`)
                .join(' ');
            setHistory(prev => [...prev, `${currentSpl} ${refinements}`]);
        } else if (mode === 'exclude') {
            refinements = Object.entries(filters)
                .filter(([k]) => !k.startsWith('_'))
                .map(([k, v]) => `| search ${k}!="${v}"`)
                .join(' ');
            setHistory(prev => [...prev, `${currentSpl} ${refinements}`]);
        } else if (mode === 'pivot') {
            // Pivot by the first categorical key found
            const pivotKey = Object.keys(filters).find(k => !k.startsWith('_')) || 'host';
            setHistory(prev => [...prev, `${currentSpl} | stats count by ${pivotKey}`]);
        } else if (mode === 'timezoom' && filters._time) {
            const timestamp = new Date(filters._time as string).getTime();
            const start = new Date(timestamp - 300000).toISOString(); // -5m
            const end = new Date(timestamp + 300000).toISOString();   // +5m
            setHistory(prev => [...prev, `${currentSpl} | where _time > "${start}" AND _time < "${end}"`]);
        } else if (mode === 'alert') {
            alert('Opening Alert Creation Modal with filters: ' + JSON.stringify(filters));
        } else if (mode === 'copy') {
            const fragment = Object.entries(filters)
                .filter(([k]) => !k.startsWith('_'))
                .map(([k, v]) => `${k}="${v}"`)
                .join(' AND ');
            navigator.clipboard.writeText(`| search ${fragment}`);
            // Briefly show notification (simulated)
        } else if (mode === 'forensic') {
            const firstVal = String(Object.values(filters)[0]);
            window.open(`https://www.virustotal.com/gui/search/${encodeURIComponent(firstVal)}`, '_blank');
        } else if (mode === 'global' && onGlobalSearch) {
            const refinement = Object.entries(filters)
                .filter(([k]) => !k.startsWith('_'))
                .map(([k, v]) => `| search ${k}="${v}"`)
                .join(' ');
            onGlobalSearch(`${currentSpl} ${refinement}`);
        }

        setActiveDrillDown(null);
    };

    const handleBack = () => {
        if (history.length > 1) {
            setHistory(prev => prev.slice(0, -1));
        }
    };

    const handleSaveAsWidget = () => {
        if (onAddWidget) {
            onAddWidget({
                title: `${title} (Filtered)`,
                type: 'auto',
                spl: currentSpl
            });
        }
    };

    const getStatusBadge = () => {
        switch (search.status) {
            case 'streaming':
                return <Badge className="h-4 animate-pulse">Streaming</Badge>;
            case 'loading':
                return <Badge variant="outline" className="h-4">Loading</Badge>;
            case 'complete':
                return <Badge variant="success" className="h-4">Live</Badge>;
            case 'error':
                return <Badge variant="error" className="h-4">Error</Badge>;
            default:
                return null;
        }
    };

    const isLoading = search.status === 'loading' || (search.status === 'idle' && !search.results);

    return (
        <Card className={`h-full flex flex-col overflow-hidden bg-slate-900/50 border-slate-800 backdrop-blur-sm group shadow-2xl transition-all duration-300 ${!isLocked ? 'ring-1 ring-brand-500/20 shadow-brand-500/5' : ''}`}>
            <div className="flex items-center justify-between px-4 py-2 border-b border-slate-800/50 bg-slate-900/30">
                <div className="flex items-center gap-2 overflow-hidden text-nowrap">
                    {!isLocked && (
                        <div className="drag-handle cursor-grab active:cursor-grabbing text-slate-500 hover:text-slate-300 transition-colors flex-shrink-0">
                            <GripVertical className="h-4 w-4" />
                        </div>
                    )}
                    <div className="flex items-center gap-1 overflow-hidden">
                        {history.length > 1 && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-slate-500 hover:text-brand-400"
                                onClick={handleBack}
                                title="Go Back"
                            >
                                <ChevronLeft className="h-4 w-4" />
                            </Button>
                        )}
                        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest truncate">{title}</span>
                        {history.length > 1 && (
                            <Badge variant="outline" className="h-4 px-1 text-[8px] border-slate-700 text-slate-500">
                                Layer {history.length - 1}
                            </Badge>
                        )}
                        {getStatusBadge()}
                    </div>
                </div>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                    {history.length > 1 && onAddWidget && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 p-0 text-brand-500 hover:text-brand-400"
                            onClick={handleSaveAsWidget}
                            title="Save as New Widget"
                        >
                            <PlusCircle className="h-3.5 w-3.5" />
                        </Button>
                    )}

                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 text-slate-500 hover:text-slate-300"
                        onClick={() => search.execute(currentSpl)}
                        title="Refresh"
                    >
                        <RefreshCw className={`h-3 w-3 ${search.status === 'streaming' ? 'animate-spin' : ''}`} />
                    </Button>

                    {search.results && search.results.length > 0 && (
                        <div className="flex items-center gap-1 border-l border-slate-800 ml-1 pl-1">
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-slate-500 hover:text-indigo-400"
                                onClick={() => exportToCSV(search.results!, title, currentSpl)}
                                title="Export CSV"
                            >
                                <Database className="h-3 w-3" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-slate-500 hover:text-amber-400"
                                onClick={() => exportToPDF(`widget-content-${id}`, title, currentSpl)}
                                title="Export PDF"
                            >
                                <FileText className="h-3 w-3" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0 text-slate-500 hover:text-emerald-400"
                                onClick={() => exportToJSON(search.results!, title, currentSpl)}
                                title="Export JSON"
                            >
                                <Code className="h-3 w-3" />
                            </Button>
                        </div>
                    )}

                    {!isLocked && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 p-0 text-slate-500 hover:text-red-400"
                            onClick={() => onRemove(id)}
                        >
                            <X className="h-4 w-4" />
                        </Button>
                    )}
                </div>
            </div>
            <div id={`widget-content-${id}`} className="flex-1 min-h-0 p-4 relative bg-slate-900/50 rounded-b-2xl">
                <WidgetErrorBoundary title={title}>
                    {isLoading ? (
                        <div className="space-y-3 h-full">
                            <Skeleton className="h-[60%] w-full" />
                            <div className="flex gap-2 h-[20%]">
                                <Skeleton className="flex-1" />
                                <Skeleton className="flex-1" />
                                <Skeleton className="flex-1" />
                            </div>
                        </div>
                    ) : (
                        <UniversalVisualizer progress={search} onDrillDown={handleDrillDownEvent} />
                    )}
                </WidgetErrorBoundary>
            </div>

            {activeDrillDown && (
                <DrillDownMenu
                    x={activeDrillDown.x}
                    y={activeDrillDown.y}
                    filters={activeDrillDown.filters}
                    onAction={executeDrillDown}
                    onClose={() => setActiveDrillDown(null)}
                />
            )}
        </Card>
    );
}
