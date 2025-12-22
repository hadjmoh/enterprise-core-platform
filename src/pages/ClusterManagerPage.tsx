import React, { useState, useEffect } from 'react';
import {
    Server,
    Cpu,
    Activity,
    ShieldCheck,
    TrendingUp,
    Settings,
    DollarSign,
    AlertTriangle,
    Play,
    Pause,
    Plus,
    RefreshCw,
    Clock
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';

interface ClusterNode {
    id: string;
    address: string;
    status: 'online' | 'offline' | 'degraded';
    uptime_seconds: number;
    metrics: {
        cpu_usage: number;
        mem_usage: number;
        disk_usage: number;
        ingest_rate: number;
    };
}

interface ScalingEvent {
    timestamp: string;
    action: string;
    nodes_delta: number;
    reason: string;
    cost_estimate: number;
}

interface ScalingPolicy {
    min_nodes: number;
    max_nodes: number;
    target_cpu: number;
    target_memory: number;
    scaling_cooldown: number;
    blast_radius_limit: number;
    cost_limit_per_month: number;
    autoscaling_enabled: boolean;
}

export const ClusterManagerPage: React.FC = () => {
    const [nodes, setNodes] = useState<ClusterNode[]>([]);
    const [history, setHistory] = useState<ScalingEvent[]>([]);
    const [policy, setPolicy] = useState<ScalingPolicy | null>(null);
    const [killSwitch, setKillSwitch] = useState(false);
    const [loading, setLoading] = useState(true);

    const fetchClusterStatus = async () => {
        try {
            const response = await fetch('/api/v1/cluster/status');
            const data = await response.json();
            setNodes(data.nodes);
            setHistory(data.history || []);
            setPolicy(data.policy);
            setKillSwitch(data.kill_switch);
            setLoading(false);
        } catch (error) {
            console.error('Failed to fetch cluster status:', error);
        }
    };

    useEffect(() => {
        fetchClusterStatus();
        const interval = setInterval(fetchClusterStatus, 5000);
        return () => clearInterval(interval);
    }, []);

    const handleManualScale = async (action: 'scale_up' | 'scale_down') => {
        try {
            await fetch('/api/v1/cluster/scale', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    action,
                    delta: 1,
                    reason: 'Manual override from UI'
                })
            });
            fetchClusterStatus();
        } catch (error) {
            console.error('Scaling action failed:', error);
        }
    };

    const toggleKillSwitch = async () => {
        try {
            await fetch('/api/v1/cluster/killswitch', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ enabled: !killSwitch })
            });
            setKillSwitch(!killSwitch);
        } catch (error) {
            console.error('Kill-switch toggle failed:', error);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-screen bg-slate-950">
                <RefreshCw className="w-8 h-8 text-indigo-500 animate-spin" />
            </div>
        );
    }

    return (
        <div className="space-y-6 p-6 bg-slate-950 min-h-screen text-slate-100 font-sans">
            {/* Header & Kill Switch */}
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-2">
                        <Server className="text-indigo-500" />
                        Cluster Intelligence
                    </h1>
                    <p className="text-slate-400 mt-1">Cost-aware elastic scaling & node management</p>
                </div>
                <div className="flex gap-4">
                    <Button
                        variant={killSwitch ? "accent" : "outline"}
                        className="flex items-center gap-2"
                        onClick={toggleKillSwitch}
                    >
                        {killSwitch ? <Play className="w-4 h-4" /> : <Pause className="w-4 h-4" />}
                        {killSwitch ? "REACTIVE AUTO-SCALING" : "KILL-SWITCH (HALT SCALING)"}
                    </Button>
                    <Button variant="primary" className="flex items-center gap-2" onClick={() => handleManualScale('scale_up')}>
                        <Plus className="w-4 h-4" />
                        SCALE UP
                    </Button>
                </div>
            </div>

            {/* Cluster Stats Overview */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card className="bg-slate-900/50 border-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex justify-between items-start">
                            <div>
                                <p className="text-sm font-medium text-slate-400">Total Nodes</p>
                                <h3 className="text-2xl font-bold text-white">{nodes.length}</h3>
                            </div>
                            <div className="p-2 bg-indigo-500/10 rounded-lg">
                                <Server className="w-5 h-5 text-indigo-400" />
                            </div>
                        </div>
                        <div className="mt-4 flex items-center gap-2 text-xs">
                            <Badge variant="success" className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20">HEALTHY</Badge>
                            <span className="text-slate-500">All nodes operational</span>
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-slate-900/50 border-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex justify-between items-start">
                            <div>
                                <p className="text-sm font-medium text-slate-400">Avg Cluster CPU</p>
                                <h3 className="text-2xl font-bold text-white">
                                    {(nodes.reduce((acc, n) => acc + n.metrics.cpu_usage, 0) / nodes.length).toFixed(1)}%
                                </h3>
                            </div>
                            <div className="p-2 bg-blue-500/10 rounded-lg">
                                <Cpu className="w-5 h-5 text-blue-400" />
                            </div>
                        </div>
                        <div className="mt-4 w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-blue-500"
                                style={{ width: `${nodes.reduce((acc, n) => acc + n.metrics.cpu_usage, 0) / nodes.length}%` }}
                            />
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-slate-900/50 border-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex justify-between items-start">
                            <div>
                                <p className="text-sm font-medium text-slate-400">Network Ingest</p>
                                <h3 className="text-2xl font-bold text-white">
                                    {(nodes.reduce((acc, n) => acc + n.metrics.ingest_rate, 0) / 1000).toFixed(1)}k <span className="text-sm text-slate-500">EPS</span>
                                </h3>
                            </div>
                            <div className="p-2 bg-pink-500/10 rounded-lg">
                                <TrendingUp className="w-5 h-5 text-pink-400" />
                            </div>
                        </div>
                        <p className="text-xs text-slate-500 mt-4 flex items-center gap-1">
                            <Activity className="w-3 h-3 text-emerald-500" />
                            Normal flow detected
                        </p>
                    </CardContent>
                </Card>

                <Card className="bg-slate-900/50 border-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex justify-between items-start">
                            <div>
                                <p className="text-sm font-medium text-slate-400">Est. Monthly Cost</p>
                                <h3 className="text-2xl font-bold text-white">${nodes.length * 100}</h3>
                            </div>
                            <div className="p-2 bg-emerald-500/10 rounded-lg">
                                <DollarSign className="w-5 h-5 text-emerald-400" />
                            </div>
                        </div>
                        <div className="mt-4 flex items-center justify-between text-xs text-slate-500">
                            <span>Budget used:</span>
                            <span>{((nodes.length * 100 / (policy?.cost_limit_per_month || 5000)) * 100).toFixed(1)}%</span>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Node Grid */}
                <div className="lg:col-span-2 space-y-4">
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        Active Nodes
                        <Badge className="bg-indigo-500/20 text-indigo-400 border-none">{nodes.length}</Badge>
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {nodes.map(node => (
                            <Card key={node.id} className="bg-slate-900 border-slate-800 hover:border-indigo-500/30 transition-all cursor-pointer group">
                                <CardHeader className="pb-2">
                                    <div className="flex justify-between items-center">
                                        <CardTitle className="text-sm font-bold flex items-center gap-2">
                                            <div className={`w-2 h-2 rounded-full ${node.status === 'online' ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                                            {node.id.toUpperCase()}
                                        </CardTitle>
                                        <Badge variant="outline" className="text-[10px] text-slate-500 border-slate-800">{node.address}</Badge>
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    <div className="grid grid-cols-2 gap-y-4 gap-x-8">
                                        <div>
                                            <p className="text-[10px] text-slate-500 uppercase tracking-wider mb-1">CPU Load</p>
                                            <div className="flex items-center gap-2">
                                                <div className="flex-1 bg-slate-800 h-1.5 rounded-full overflow-hidden">
                                                    <div className={`h-full ${node.metrics.cpu_usage > 80 ? 'bg-red-500' : 'bg-indigo-500'}`} style={{ width: `${node.metrics.cpu_usage}%` }} />
                                                </div>
                                                <span className="text-xs font-medium">{node.metrics.cpu_usage.toFixed(0)}%</span>
                                            </div>
                                        </div>
                                        <div>
                                            <p className="text-[10px] text-slate-500 uppercase tracking-wider mb-1">Memory</p>
                                            <div className="flex items-center gap-2">
                                                <div className="flex-1 bg-slate-800 h-1.5 rounded-full overflow-hidden">
                                                    <div className={`h-full ${node.metrics.mem_usage > 80 ? 'bg-amber-500' : 'bg-blue-500'}`} style={{ width: `${node.metrics.mem_usage}%` }} />
                                                </div>
                                                <span className="text-xs font-medium">{node.metrics.mem_usage.toFixed(0)}%</span>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="mt-4 pt-4 border-t border-slate-800 flex justify-between items-center">
                                        <div className="flex items-center gap-1 text-[10px] text-slate-500">
                                            <Clock className="w-3 h-3" />
                                            UPTIME: {Math.floor(node.uptime_seconds / 3600)}h
                                        </div>
                                        <Button variant="ghost" size="sm" className="h-7 text-[10px] hover:bg-slate-800">
                                            DIAGNOSTICS
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                </div>

                {/* Sidebar: Scaling History & Policy */}
                <div className="space-y-6">
                    <Card className="bg-slate-900 border-slate-800">
                        <CardHeader>
                            <CardTitle className="text-md flex items-center gap-2">
                                <ShieldCheck className="w-5 h-5 text-indigo-400" />
                                Scaling Policy
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-400">Target CPU</span>
                                <span className="text-white font-medium">{policy?.target_cpu}%</span>
                            </div>
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-400">Node Range</span>
                                <span className="text-white font-medium">{policy?.min_nodes} - {policy?.max_nodes}</span>
                            </div>
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-400">Blast Radius</span>
                                <span className="text-white font-medium">MAX +{policy?.blast_radius_limit}</span>
                            </div>
                            <Button variant="outline" className="w-full mt-2 text-xs border-slate-700 hover:bg-slate-800 gap-2">
                                <Settings className="w-3 h-3" />
                                EDIT POLICY
                            </Button>
                        </CardContent>
                    </Card>

                    <Card className="bg-slate-900 border-slate-800">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-md flex items-center gap-2">
                                <Activity className="w-5 h-5 text-pink-400" />
                                Scaling History
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="p-0">
                            <div className="divide-y divide-slate-800">
                                {history.length > 0 ? (
                                    history.slice(0, 5).map((ev, i) => (
                                        <div key={i} className="p-4 hover:bg-slate-800/30 transition-colors">
                                            <div className="flex justify-between items-start mb-1">
                                                <Badge className={`${ev.action === 'scale_up' ? 'bg-indigo-500/20 text-indigo-400' : 'bg-amber-500/20 text-amber-400'} border-none text-[10px]`}>
                                                    {ev.action.toUpperCase()} (+{ev.nodes_delta})
                                                </Badge>
                                                <span className="text-[10px] text-slate-500">{new Date(ev.timestamp).toLocaleTimeString()}</span>
                                            </div>
                                            <p className="text-xs text-slate-300 line-clamp-1">{ev.reason}</p>
                                            <div className="flex items-center gap-1 text-[10px] text-indigo-400 mt-2 font-medium">
                                                <DollarSign className="w-3 h-3" />
                                                Cost Impact: +${ev.cost_estimate}/mo
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="p-8 text-center text-slate-500 italic text-sm">
                                        No recent scaling events.
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    {killSwitch && (
                        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 flex items-start gap-4">
                            <AlertTriangle className="w-6 h-6 text-red-500 flex-shrink-0" />
                            <div>
                                <h4 className="text-sm font-bold text-red-400">Scaling Suspended</h4>
                                <p className="text-[10px] text-red-300/80 mt-1">
                                    Global kill-switch is ACTIVE. Auto-scaling is disabled and resource bottlenecks will not be automated corrected.
                                </p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
