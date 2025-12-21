import React, { useState, useEffect } from 'react';
import {
    ShieldAlert,
    Settings,
    Play,
    Pause,
    Trash2,
    Plus,
    Zap
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { correlationService } from '../services/correlationService';
import type { CorrelationRule, DetectionSignal } from '../services/correlationService';
import { ingestionService } from '../services/ingestionService';

export const SecurityRulesPage: React.FC = () => {
    const [rules, setRules] = useState<CorrelationRule[]>([]);
    const [signals, setSignals] = useState<DetectionSignal[]>([]);
    const [isIngesting, setIsIngesting] = useState(false);

    useEffect(() => {
        refreshData();

        // Subscribe to new signals
        correlationService.onSignal((sid) => {
            setSignals(prev => [sid, ...prev].slice(0, 50));
        });

        const interval = setInterval(refreshData, 2000);
        return () => clearInterval(interval);
    }, []);

    const refreshData = () => {
        setRules([...correlationService.getRules()]);
        // ingestion status is private in service, we'll track local state for toggle
    };

    const toggleIngestion = () => {
        if (isIngesting) {
            ingestionService.stopMockIngestion();
            setIsIngesting(false);
        } else {
            ingestionService.startMockIngestion();
            setIsIngesting(true);
        }
    };

    const toggleRule = (id: string) => {
        correlationService.toggleRule(id);
        refreshData();
    };

    const deleteRule = (id: string) => {
        if (confirm('Delete this detection rule?')) {
            correlationService.deleteRule(id);
            refreshData();
        }
    };

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-slate-100 font-sans">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-orange-400 to-red-400 bg-clip-text text-transparent flex items-center gap-2">
                        <Zap className="h-8 w-8 text-orange-400" />
                        Detection Rules Engine
                    </h1>
                    <p className="text-slate-400 mt-1">Manage real-time correlation logic and signals.</p>
                </div>
                <div className="flex gap-3">
                    <Button variant={isIngesting ? 'secondary' : 'primary'} onClick={toggleIngestion}>
                        {isIngesting ? <><Pause className="mr-2 h-4 w-4" /> Stop Simulation</> : <><Play className="mr-2 h-4 w-4" /> Start Simulation</>}
                    </Button>
                    <Button variant="outline">
                        <Plus className="mr-2 h-4 w-4" /> New Rule
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Rules List */}
                <div className="lg:col-span-2 space-y-4">
                    <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                        <Settings className="h-5 w-5 text-slate-400" />
                        Active Correlation Rules
                    </h2>
                    {rules.map(rule => (
                        <Card key={rule.id} className={`p-4 border-l-4 transition-all ${rule.enabled ? 'border-l-emerald-500 bg-slate-800/50' : 'border-l-slate-600 bg-slate-900/50 opacity-70'}`}>
                            <div className="flex justify-between items-start">
                                <div>
                                    <div className="flex items-center gap-3">
                                        <h3 className="font-bold text-lg text-white">{rule.name}</h3>
                                        <Badge variant={rule.severity === 'critical' ? 'error' : rule.severity === 'high' ? 'warning' : 'default'}>
                                            {rule.severity.toUpperCase()}
                                        </Badge>
                                        {rule.threshold && <Badge variant="outline" className="text-xs">Threshold: {rule.threshold.count}/{rule.threshold.windowSeconds}s</Badge>}
                                    </div>
                                    <p className="text-slate-400 text-sm mt-1">{rule.description}</p>
                                    <div className="mt-2 font-mono text-xs bg-slate-950 p-2 rounded text-emerald-400 inline-block">
                                        {rule.logic.field} {rule.logic.operator} '{rule.logic.value}'
                                    </div>
                                </div>
                                <div className="flex gap-2">
                                    <Button variant="ghost" size="sm" onClick={() => toggleRule(rule.id)}>
                                        {rule.enabled ? <Pause className="h-4 w-4 text-emerald-400" /> : <Play className="h-4 w-4 text-slate-400" />}
                                    </Button>
                                    <Button variant="ghost" size="sm" onClick={() => deleteRule(rule.id)}>
                                        <Trash2 className="h-4 w-4 text-red-400" />
                                    </Button>
                                </div>
                            </div>
                        </Card>
                    ))}
                </div>

                {/* Live Signals Feed */}
                <div className="space-y-4">
                    <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                        <ShieldAlert className="h-5 w-5 text-red-500 animate-pulse" />
                        Live Detection Signals
                    </h2>
                    <div className="bg-slate-950 rounded-lg border border-slate-800 h-[600px] overflow-y-auto p-2 space-y-2">
                        {signals.length === 0 && (
                            <div className="text-center text-slate-500 py-10 italic">
                                No detections yet. <br />Start simulation to generate traffic.
                            </div>
                        )}
                        {signals.map(signal => (
                            <div key={signal.id} className="p-3 rounded border border-slate-800 bg-slate-900/80 hover:bg-slate-800 transition-colors animate-in slide-in-from-right duration-300">
                                <div className="flex justify-between items-start mb-1">
                                    <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${signal.severity === 'critical' ? 'bg-red-500/20 text-red-400' :
                                            signal.severity === 'high' ? 'bg-orange-500/20 text-orange-400' :
                                                'bg-blue-500/20 text-blue-400'
                                        }`}>
                                        {signal.severity}
                                    </span>
                                    <span className="text-xs text-slate-500">{new Date(signal.timestamp).toLocaleTimeString()}</span>
                                </div>
                                <p className="text-sm font-medium text-slate-200">{signal.message}</p>
                                <p className="text-xs text-slate-500 mt-1">Rule: {signal.ruleName}</p>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};
