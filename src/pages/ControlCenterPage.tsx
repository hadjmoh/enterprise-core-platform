import React, { useEffect, useState } from 'react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import {
    ShieldCheck,
    AlertTriangle,
    Power,
    Lock,
    Unlock,
    RefreshCcw,
    Activity,
    Settings,
    FileSearch,
    Fingerprint,
    ZapOff,
    SearchX
} from 'lucide-react';
import { cn } from '../lib/utils';

interface GovernanceStatus {
    lockdown: boolean;
    switches: Record<string, boolean>;
}

interface DriftReport {
    timestamp: string;
    drift_detected: boolean;
    discrepancies: string[];
}

export const ControlCenterPage: React.FC = () => {
    const [status, setStatus] = useState<GovernanceStatus | null>(null);
    const [drift, setDrift] = useState<DriftReport | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [confirmLockdown, setConfirmLockdown] = useState(false);

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 5000);
        return () => clearInterval(interval);
    }, []);

    const fetchData = async () => {
        try {
            const [sRes, dRes] = await Promise.all([
                fetch('/api/v1/governance/status'),
                fetch('/api/v1/governance/drift')
            ]);
            setStatus(await sRes.json());
            setDrift(await dRes.json());
        } catch (err) {
            console.error("Failed to fetch governance data", err);
        }
    };

    const toggleSwitch = async (sub: string, active: boolean) => {
        setIsLoading(true);
        try {
            await fetch('/api/v1/governance/emergency/killswitch', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ subsystem: sub, active: !active })
            });
            fetchData();
        } finally {
            setIsLoading(false);
        }
    };

    const triggerLockdown = async () => {
        if (!confirmLockdown) {
            setConfirmLockdown(true);
            return;
        }
        setIsLoading(true);
        try {
            await fetch('/api/v1/governance/emergency/lockdown', { method: 'POST' });
            setConfirmLockdown(false);
            fetchData();
        } finally {
            setIsLoading(false);
        }
    };

    const releaseLockdown = async () => {
        setIsLoading(true);
        try {
            await fetch('/api/v1/governance/emergency/release', { method: 'POST' });
            fetchData();
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="p-6 space-y-6 bg-slate-950 min-h-screen text-slate-100">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-3">
                        <ShieldCheck className="w-8 h-8 text-indigo-400" />
                        Security Control Center
                    </h1>
                    <p className="text-slate-400 mt-1">Foundational governance and emergency orchestration plane.</p>
                </div>
                <div className="flex gap-3">
                    <Button variant="outline" size="sm" onClick={fetchData}>
                        <RefreshCcw className="w-4 h-4 mr-2" /> Refresh
                    </Button>
                    {status?.lockdown ? (
                        <Button variant="primary" size="sm" onClick={releaseLockdown}>
                            <Unlock className="w-4 h-4 mr-2" /> Release Lockdown
                        </Button>
                    ) : (
                        <Button
                            className={cn(
                                "transition-all duration-300",
                                confirmLockdown ? "bg-red-600 hover:bg-red-700 animate-pulse" : "bg-red-900/50 hover:bg-red-800 border-red-500/30"
                            )}
                            size="sm"
                            onClick={triggerLockdown}
                        >
                            <Lock className="w-4 h-4 mr-2" /> {confirmLockdown ? "CONFIRM GLOBAL LOCKDOWN" : "Emergency Lockdown"}
                        </Button>
                    )}
                </div>
            </div>

            {/* System Status Banner */}
            {status?.lockdown && (
                <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-xl flex items-center justify-between animate-in fade-in slide-in-from-top">
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center animate-pulse">
                            <AlertTriangle className="w-6 h-6 text-red-500" />
                        </div>
                        <div>
                            <h2 className="text-lg font-bold text-red-500 uppercase tracking-widest">Global Lockdown Active</h2>
                            <p className="text-red-400/70 text-sm font-medium">All non-essential services halted. System is in read-only defensive posture.</p>
                        </div>
                    </div>
                    <Badge className="bg-red-500 text-white animate-bounce">CRITICAL</Badge>
                </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Subsystem Kill-switches */}
                <Card className="lg:col-span-2 p-6 bg-slate-900 border-slate-800">
                    <h3 className="font-bold text-lg mb-6 flex items-center gap-2">
                        <Activity className="w-5 h-5 text-indigo-400" />
                        Subsystem Orchestration
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {[
                            { id: 'ingest', name: 'Data Ingestion', icon: <Fingerprint />, color: 'blue' },
                            { id: 'search', name: 'Search Engine', icon: <SearchX />, color: 'amber' },
                            { id: 'alerting', name: 'Alert Dispatch', icon: <ZapOff />, color: 'purple' }
                        ].map(sub => {
                            const active = status?.switches[sub.id] || false;
                            return (
                                <div key={sub.id} className={cn(
                                    "p-4 rounded-xl border transition-all duration-300",
                                    active ? "bg-slate-800/20 border-red-500/30" : "bg-slate-900 border-slate-800 hover:border-slate-700"
                                )}>
                                    <div className="flex justify-between items-start mb-4">
                                        <div className={cn("p-2 rounded-lg", active ? "bg-red-500/20 text-red-400" : "bg-slate-800 text-slate-400")}>
                                            {React.cloneElement(sub.icon as React.ReactElement<any>, { className: 'w-5 h-5' })}
                                        </div>
                                        <Badge variant={active ? 'error' : 'success'}>
                                            {active ? 'Disabled' : 'Operational'}
                                        </Badge>
                                    </div>
                                    <h4 className="font-bold text-sm mb-1">{sub.name}</h4>
                                    <p className="text-[10px] text-slate-500 mb-4 uppercase font-bold">Subsystem-01/Primary</p>
                                    <Button
                                        variant={active ? 'primary' : 'outline'}
                                        className="w-full text-xs h-8"
                                        onClick={() => toggleSwitch(sub.id, active)}
                                        disabled={isLoading || status?.lockdown}
                                    >
                                        <Power className="w-3 h-3 mr-2" /> {active ? 'Enable' : 'Halt'}
                                    </Button>
                                </div>
                            );
                        })}
                    </div>
                </Card>

                {/* Drift Detection */}
                <Card className="p-6 bg-slate-900 border-slate-800">
                    <h3 className="font-bold text-lg mb-6 flex items-center gap-2">
                        <FileSearch className="w-5 h-5 text-emerald-400" />
                        Configuration Drift
                    </h3>

                    {drift?.drift_detected ? (
                        <div className="space-y-4">
                            <div className="p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg">
                                <p className="text-xs text-amber-500 font-bold flex items-center gap-2">
                                    <AlertTriangle className="w-4 h-4" /> Integrity Mismatch Detected
                                </p>
                                <p className="text-[10px] text-amber-500/70 mt-1">Runtime parameters deviate from Golden Image.</p>
                            </div>
                            <div className="space-y-2">
                                {drift.discrepancies.map((d, i) => (
                                    <div key={i} className="text-[10px] p-2 bg-slate-950 border border-slate-800 rounded font-mono text-slate-400">
                                        {d}
                                    </div>
                                ))}
                            </div>
                            <Button variant="outline" className="w-full text-xs">
                                Re-Apply Golden Config
                            </Button>
                        </div>
                    ) : (
                        <div className="flex flex-col items-center justify-center h-48 py-8">
                            <div className="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center mb-4">
                                <ShieldCheck className="w-8 h-8 text-emerald-500" />
                            </div>
                            <p className="text-sm font-bold text-white">Config in Compliance</p>
                            <p className="text-xs text-slate-500 text-center mt-1">Verified against immutable baseline.</p>
                        </div>
                    )}
                </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Governance Dashboard */}
                <Card className="lg:col-span-8 p-6 bg-slate-900 border-slate-800">
                    <div className="flex justify-between items-center mb-6">
                        <h3 className="font-bold text-lg flex items-center gap-2">
                            <Settings className="w-5 h-5 text-slate-400" />
                            Policy Enforcement Matrix
                        </h3>
                        <Button variant="ghost" size="sm" className="text-xs">Export Audit</Button>
                    </div>

                    <div className="space-y-4">
                        {[
                            { name: 'Data Residency Enforcement', description: 'Prevent cross-region ingestion', status: 'active', priority: 'high' },
                            { name: 'PII Tokenization Mandatory', description: 'Reject non-tokenized sensitive fields', status: 'active', priority: 'critical' },
                            { name: 'Signed Lineage Required', description: 'Drop events with invalid hop signatures', status: 'active', priority: 'high' },
                            { name: 'Auto-Scaling Blast Radius', description: 'Cap max cluster expansion per hour', status: 'learning', priority: 'medium' }
                        ].map((policy, i) => (
                            <div key={i} className="flex items-center justify-between p-4 bg-slate-950/50 border border-slate-800 rounded-xl hover:bg-slate-900 transition-colors">
                                <div className="flex items-center gap-4">
                                    <div className={cn(
                                        "w-2 h-2 rounded-full",
                                        policy.status === 'active' ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-blue-500"
                                    )} />
                                    <div>
                                        <p className="text-sm font-bold text-slate-200">{policy.name}</p>
                                        <p className="text-xs text-slate-500">{policy.description}</p>
                                    </div>
                                </div>
                                <div className="flex items-center gap-3">
                                    <Badge variant={policy.priority === 'critical' ? 'error' : 'outline'} className="text-[10px]">
                                        {policy.priority}
                                    </Badge>
                                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                                        <Settings className="w-4 h-4 text-slate-600" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </Card>

                {/* Audit Trail */}
                <Card className="lg:col-span-4 p-6 bg-slate-900 border-slate-800">
                    <h3 className="font-bold text-lg mb-6">Control Log</h3>
                    <div className="space-y-4">
                        {[
                            { user: 'sys-admin', action: 'Killswitch: Ingest OFF', time: '12m ago', severity: 'warning' },
                            { user: 'ciso-auto', action: 'Lockdown Release', time: '1h ago', severity: 'info' },
                            { user: 'drift-agent', action: 'Compliance Check', time: '5m ago', severity: 'info' }
                        ].map((log, i) => (
                            <div key={i} className="flex gap-3 pb-4 border-b border-slate-800 last:border-0">
                                <div className={cn(
                                    "w-1 h-8 rounded-full",
                                    log.severity === 'warning' ? "bg-amber-500" : "bg-indigo-500"
                                )} />
                                <div>
                                    <p className="text-xs font-bold text-slate-300">{log.action}</p>
                                    <div className="flex gap-2 text-[10px] text-slate-500 mt-1">
                                        <span>{log.user}</span>
                                        <span>•</span>
                                        <span>{log.time}</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    <Button variant="ghost" className="w-full mt-6 text-xs text-indigo-400 hover:text-indigo-300">
                        View Full System Audit
                    </Button>
                </Card>
            </div>
        </div>
    );
};
