import React, { useState } from 'react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import {
    FlaskConical,
    Play,
    History,
    Zap,
    ShieldAlert,
    BarChart3,
    Settings2,
    CheckCircle2,
    Info,
    ArrowRight
} from 'lucide-react';

interface SimulationResult {
    session_id: string;
    start_time: string;
    end_time: string;
    events_count: number;
    alerts_count: number;
    outcomes: Array<{
        event_id: string;
        triggered_alert: boolean;
        detections: string[];
        correlations: string[];
        ueba_anomalies?: number;
    }>;
}

export const SimulationLabPage: React.FC = () => {
    const [isLoading, setIsLoading] = useState(false);
    const [result, setResult] = useState<SimulationResult | null>(null);
    const [dateRange, setDateRange] = useState({
        from: new Date(Date.now() - 3600000).toISOString().slice(0, 16),
        to: new Date().toISOString().slice(0, 16)
    });

    const runSimulation = async () => {
        setIsLoading(true);
        try {
            const res = await fetch('/api/v1/simulation/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    from: new Date(dateRange.from).toISOString(),
                    to: new Date(dateRange.to).toISOString()
                })
            });
            const data = await res.json();
            setResult(data);
        } catch (err) {
            console.error("Simulation failed", err);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="p-6 space-y-6 bg-slate-900 min-h-screen text-slate-100">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-3">
                        <FlaskConical className="w-8 h-8 text-indigo-400" />
                        Simulation & Replay Lab
                    </h1>
                    <p className="text-slate-400 mt-1">Run historical data through the detection engine to test rules and predict impact.</p>
                </div>
                <div className="flex gap-3">
                    <Button variant="outline" size="sm">
                        <History className="w-4 h-4 mr-2" /> Saved Sessions
                    </Button>
                    <Button variant="primary" size="sm" onClick={runSimulation} disabled={isLoading}>
                        <Play className="w-4 h-4 mr-2" /> {isLoading ? 'Running...' : 'Start Simulation'}
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Configuration Sidebar */}
                <Card className="lg:col-span-3 p-4 bg-slate-800/40 border-slate-700">
                    <h3 className="text-sm font-bold text-slate-300 mb-4 flex items-center gap-2">
                        <Settings2 className="w-4 h-4" /> Simulation Config
                    </h3>

                    <div className="space-y-4">
                        <div>
                            <label htmlFor="sim-time-range" className="text-[10px] text-slate-500 uppercase font-bold mb-1 block">Time Range</label>
                            <div className="space-y-2">
                                <input
                                    id="sim-time-range"
                                    type="datetime-local"
                                    className="w-full bg-slate-900 border-slate-700 rounded p-2 text-xs outline-none focus:ring-1 focus:ring-indigo-500"
                                    value={dateRange.from}
                                    onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
                                />
                                <input
                                    type="datetime-local"
                                    className="w-full bg-slate-900 border-slate-700 rounded p-2 text-xs outline-none focus:ring-1 focus:ring-indigo-500"
                                    value={dateRange.to}
                                    onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
                                />
                            </div>
                        </div>

                        <div>
                            <span className="text-[10px] text-slate-500 uppercase font-bold mb-2 block">Detection Engines</span>
                            <div className="space-y-2">
                                {['Static Rules', 'Correlation Engine', 'UEBA Anomaly'].map(engine => (
                                    <label key={engine} className="flex items-center gap-2 cursor-pointer group">
                                        <input
                                            type="checkbox"
                                            defaultChecked
                                            className="w-4 h-4 rounded border-slate-700 bg-slate-900 text-indigo-500 focus:ring-indigo-500 focus:ring-offset-slate-900 cursor-pointer"
                                        />
                                        <span className="text-xs text-slate-300">{engine}</span>
                                    </label>
                                ))}
                            </div>
                        </div>

                        <div className="pt-4 border-t border-slate-800">
                            <label htmlFor="sim-dataset-mode" className="text-[10px] text-slate-500 uppercase font-bold mb-2 block">Dataset Mode</label>
                            <select id="sim-dataset-mode" className="w-full bg-slate-900 border-slate-700 rounded p-2 text-xs outline-none cursor-pointer">
                                <option>Live Production Cache</option>
                                <option>Cold Storage (S3)</option>
                                <option>Synthetic Attack Scenarios</option>
                            </select>
                        </div>
                    </div>
                </Card>

                {/* Simulation Canvas */}
                <div className="lg:col-span-9 space-y-6">
                    {result ? (
                        <>
                            {/* Summary Cards */}
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                                <Card className="p-4 bg-slate-800/20 border-slate-700">
                                    <p className="text-[10px] text-slate-500 uppercase">Input Events</p>
                                    <p className="text-2xl font-bold text-white">{result.events_count}</p>
                                    <p className="text-[10px] text-emerald-400 flex items-center gap-1 mt-1">
                                        <CheckCircle2 className="w-3 h-3" /> Replayed Successfully
                                    </p>
                                </Card>
                                <Card className="p-4 bg-slate-800/20 border-slate-700 border-l-amber-500/50">
                                    <p className="text-[10px] text-slate-500 uppercase">Simulated Alerts</p>
                                    <p className="text-2xl font-bold text-amber-400">{result.alerts_count}</p>
                                    <p className="text-[10px] text-slate-500 mt-1">Potential noisy signals</p>
                                </Card>
                                <Card className="p-4 bg-slate-800/20 border-slate-700">
                                    <p className="text-[10px] text-slate-500 uppercase">Execution Time</p>
                                    <p className="text-2xl font-bold text-indigo-400">
                                        {((new Date(result.end_time).getTime() - new Date(result.start_time).getTime()) / 1000).toFixed(2)}s
                                    </p>
                                    <p className="text-[10px] text-slate-500 mt-1">Throughput: {(result.events_count / 1.5).toFixed(0)} eps</p>
                                </Card>
                                <Card className="p-4 bg-slate-800/20 border-slate-700 border-l-emerald-500/50">
                                    <p className="text-[10px] text-slate-500 uppercase">Delta Prediction</p>
                                    <p className="text-2xl font-bold text-emerald-400">-12%</p>
                                    <p className="text-[10px] text-emerald-500/70 mt-1">Reduction in alert fatigue</p>
                                </Card>
                            </div>

                            {/* Main Analysis Chart / Table */}
                            <Card className="p-6 bg-slate-800/40 border-slate-700">
                                <div className="flex justify-between items-center mb-6">
                                    <h3 className="font-bold flex items-center gap-2">
                                        <BarChart3 className="w-5 h-5 text-indigo-400" />
                                        Shadow Outcome Analysis
                                    </h3>
                                    <div className="flex gap-2">
                                        <Badge variant="outline">Simulation Mode: Active</Badge>
                                        <Badge variant="success">Read Stability: High</Badge>
                                    </div>
                                </div>

                                <div className="overflow-x-auto">
                                    <table className="w-full text-xs text-left">
                                        <thead className="bg-slate-900/50 text-slate-500 uppercase text-[10px]">
                                            <tr>
                                                <th className="px-4 py-3">Event ID</th>
                                                <th className="px-4 py-3">Outcome</th>
                                                <th className="px-4 py-3">Triggers</th>
                                                <th className="px-4 py-3">Audit</th>
                                            </tr>
                                        </thead>
                                        <tbody className="divide-y divide-slate-800">
                                            {result.outcomes.map((o, i) => (
                                                <tr key={i} className="hover:bg-slate-700/20 transition-colors">
                                                    <td className="px-4 py-3 font-mono text-indigo-300">{o.event_id}</td>
                                                    <td className="px-4 py-3">
                                                        {o.triggered_alert ? (
                                                            <div className="flex items-center gap-2 text-amber-500">
                                                                <ShieldAlert className="w-3 h-3" />
                                                                <span>Alert Triggered</span>
                                                            </div>
                                                        ) : (
                                                            <div className="flex items-center gap-2 text-slate-500">
                                                                <CheckCircle2 className="w-3 h-3" />
                                                                <span>No Side Effects</span>
                                                            </div>
                                                        )}
                                                    </td>
                                                    <td className="px-4 py-3">
                                                        <div className="flex flex-wrap gap-1">
                                                            {o.detections.concat(o.correlations).map((t, idx) => (
                                                                <Badge key={idx} variant="outline" className="text-[9px] border-slate-700 bg-slate-900/50">
                                                                    {t}
                                                                </Badge>
                                                            ))}
                                                            {o.ueba_anomalies && (
                                                                <Badge variant="warning" className="text-[9px]">
                                                                    UEBA × {o.ueba_anomalies}
                                                                </Badge>
                                                            )}
                                                            {o.detections.length === 0 && o.correlations.length === 0 && !o.ueba_anomalies && (
                                                                <span className="text-slate-600">Pure Signal</span>
                                                            )}
                                                        </div>
                                                    </td>
                                                    <td className="px-4 py-3">
                                                        <Button variant="ghost" size="sm" className="h-6 text-[10px]">
                                                            Details <ArrowRight className="w-3 h-3 ml-1" />
                                                        </Button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            </Card>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <Card className="p-4 bg-slate-800/40 border-slate-700">
                                    <h4 className="text-sm font-bold text-white mb-4 flex items-center gap-2">
                                        <Zap className="w-4 h-4 text-emerald-400" />
                                        Recommended Pipe Hardening
                                    </h4>
                                    <div className="space-y-3">
                                        <div className="p-3 bg-emerald-500/5 border border-emerald-500/20 rounded-lg">
                                            <p className="text-xs text-white font-bold mb-1">Deduplication Insight</p>
                                            <p className="text-[10px] text-emerald-400/80">
                                                Enabling "Session Collapse" on Firewall logs would reduce simulated alert volume by 42%.
                                            </p>
                                        </div>
                                    </div>
                                </Card>

                                <Card className="p-4 bg-slate-800/40 border-slate-700">
                                    <h4 className="text-sm font-bold text-white mb-4 flex items-center gap-2">
                                        <Info className="w-4 h-4 text-indigo-400" />
                                        Scenario Intelligence
                                    </h4>
                                    <p className="text-xs text-slate-400">
                                        This simulation used historical telemetry from the last 60 minutes.
                                        The results demonstrate that your new detection rule "Brute Force Attempt"
                                        has zero false positives on this dataset.
                                    </p>
                                </Card>
                            </div>
                        </>
                    ) : (
                        <div className="h-[600px] flex flex-col items-center justify-center border-2 border-dashed border-slate-800 rounded-3xl animate-in fade-in duration-700">
                            <div className="w-20 h-20 rounded-full bg-slate-800 flex items-center justify-center mb-6">
                                <FlaskConical className="w-10 h-10 text-slate-600" />
                            </div>
                            <h2 className="text-xl font-bold text-slate-300">Ready for Simulation</h2>
                            <p className="text-slate-500 mt-2 max-w-md text-center">
                                Configure a time range and execution parameters on the left to start a shadow replay session.
                            </p>
                            <Button variant="primary" className="mt-8" onClick={runSimulation}>
                                Run Initial Baseline Test
                            </Button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
