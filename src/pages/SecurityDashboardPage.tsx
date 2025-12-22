import React, { useState, useEffect } from 'react';
import {
    ShieldCheck,
    Activity,
    BarChart2,
    Layout,
    Target,
    Clock,
    Briefcase
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { MitreHeatmap } from '../components/visualizations/MitreHeatmap';
import { correlationService } from '../services/correlationService';
import { ExplanationModal } from '../components/analytics/ExplanationModal';

export const SecurityDashboardPage: React.FC = () => {
    const [activeTab, setActiveTab] = useState<'strategic' | 'operational'>('strategic');
    const [riskScore, setRiskScore] = useState(0);
    const [activeIncidents, setActiveIncidents] = useState(0);
    const [mitreCoverage, setMitreCoverage] = useState<Record<string, number>>({});
    const [explanation, setExplanation] = useState<any | null>(null);

    useEffect(() => {
        refreshMetrics();
        const interval = setInterval(refreshMetrics, 5000);
        return () => clearInterval(interval);
    }, []);

    const refreshMetrics = () => {
        // Calculate Pseudo-Metrics based on active rules/signals
        const rules = correlationService.getRules().filter(r => r.enabled);
        const signals = correlationService.getSignals();

        // 1. Risk Score: 0-100 based on recent critical alerts
        const criticalCount = signals.filter(s => s.severity === 'critical').length;
        const newRisk = Math.min(100, criticalCount * 10 + (signals.length * 0.5));
        setRiskScore(Math.floor(newRisk));

        // 2. Active Incidents (mock count for now)
        setActiveIncidents(Math.floor(signals.length / 5));

        // 3. MITRE Coverage Mapping (Mock logic based on rules)
        setMitreCoverage({
            'Brute Force': rules.filter(r => r.name.includes('Brute')).length + 1,
            'Sudo': 1,
            'Phishing': 1,
            'PowerShell': 2,
            'Process Injection': 1,
            'SMB/Windows Admin': 1
        });
    };

    const handleExplainRisk = async () => {
        // Fetch explanation from API
        try {
            const res = await fetch(`/api/explain/anomaly?entity=global&type=risk_score&value=${riskScore}`);
            if (res.ok) {
                const data = await res.json();
                setExplanation(data);
            } else {
                // Fallback for demo if API fails
                setExplanation({
                    entity_id: 'global_posture',
                    anomaly_type: 'risk_score_spike',
                    confidence: 0.89,
                    description: 'Global risk score elevated due to multiple critical signals.',
                    factors: [
                        { name: 'Critical Alerts', weight: 0.7, value: 5, description: 'High volume of critical severity alerts.' },
                        { name: 'New Threat Actor', weight: 0.3, value: 1, description: 'Detection of APT-29 related TTPs.' }
                    ]
                });
            }
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-slate-100 font-sans">
            {/* Header */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-400 to-purple-400 bg-clip-text text-transparent flex items-center gap-2">
                        <Layout className="h-8 w-8 text-indigo-400" />
                        Security Posture
                    </h1>
                    <p className="text-slate-400 mt-1">Situational awareness and compliance monitoring.</p>
                </div>
                <div className="bg-slate-800 p-1 rounded-lg flex gap-1">
                    <Button
                        variant={activeTab === 'strategic' ? 'primary' : 'ghost'}
                        size="sm"
                        onClick={() => setActiveTab('strategic')}
                    >
                        <Briefcase className="h-4 w-4 mr-2" /> Strategic (CISO)
                    </Button>
                    <Button
                        variant={activeTab === 'operational' ? 'primary' : 'ghost'}
                        size="sm"
                        onClick={() => setActiveTab('operational')}
                    >
                        <Activity className="h-4 w-4 mr-2" /> Operational (SOC)
                    </Button>
                </div>
            </div>

            {/* Strategic View */}
            {activeTab === 'strategic' && (
                <div className="space-y-6 animate-in fade-in slide-in-from-left duration-300">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <Card className="p-6 border-slate-700 bg-slate-800/40 relative overflow-hidden">
                            <div className="absolute top-0 right-0 p-3 opacity-10">
                                <ShieldCheck className="h-24 w-24 text-indigo-400" />
                            </div>
                            <div className="flex justify-between items-start">
                                <h3 className="text-slate-400 font-medium mb-2">Global Risk Score</h3>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-6 px-2 text-xs text-indigo-400 hover:bg-indigo-500/10"
                                    onClick={handleExplainRisk}
                                >
                                    Why?
                                </Button>
                            </div>
                            <div className="flex items-end gap-3">
                                <span className={`text-5xl font-bold ${riskScore > 80 ? 'text-red-500' :
                                    riskScore > 50 ? 'text-orange-400' : 'text-emerald-400'
                                    }`}>
                                    {riskScore}
                                </span>
                                <span className="text-slate-500 text-sm mb-1">/ 100</span>
                            </div>
                            <div className="w-full bg-slate-700 h-2 mt-4 rounded-full overflow-hidden">
                                <div
                                    className={`h-full transition-all duration-1000 ${riskScore > 80 ? 'bg-red-500' :
                                        riskScore > 50 ? 'bg-orange-400' : 'bg-emerald-400'
                                        }`}
                                    style={{ width: `${riskScore}%` }}
                                ></div>
                            </div>
                        </Card>

                        <Card className="p-6 border-slate-700 bg-slate-800/40">
                            <h3 className="text-slate-400 font-medium mb-4 flex items-center justify-between">
                                <span>Mean Time to Detect (MTTD)</span>
                                <Badge variant="success" className="bg-emerald-500/10 text-emerald-400 border-none">-12%</Badge>
                            </h3>
                            <div className="flex items-center gap-4">
                                <Clock className="h-8 w-8 text-blue-400" />
                                <div>
                                    <span className="text-3xl font-bold text-white">4m 12s</span>
                                    <p className="text-xs text-slate-500">vs 4m 45s last week</p>
                                </div>
                            </div>
                        </Card>

                        <Card className="p-6 border-slate-700 bg-slate-800/40">
                            <h3 className="text-slate-400 font-medium mb-4">Active Incidents</h3>
                            <div className="flex items-center gap-4">
                                <Target className="h-8 w-8 text-red-400" />
                                <div>
                                    <span className="text-3xl font-bold text-white">{activeIncidents}</span>
                                    <p className="text-xs text-slate-500">Requiring attention</p>
                                </div>
                            </div>
                        </Card>
                    </div>

                    <MitreHeatmap coverage={mitreCoverage} />
                </div>
            )}

            {/* Operational View */}
            {activeTab === 'operational' && (
                <div className="space-y-6 animate-in fade-in slide-in-from-right duration-300">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <Card className="p-0 border-slate-700 bg-slate-800/40 h-80 flex flex-col">
                            <div className="p-4 border-b border-slate-700">
                                <h3 className="font-bold text-white flex items-center gap-2">
                                    <BarChart2 className="h-4 w-4" /> Alert Volume (24h)
                                </h3>
                            </div>
                            <div className="flex-1 flex items-center justify-center text-slate-500 italic">
                                <div className="flex items-end gap-2 h-40 w-3/4 justify-between">
                                    {[40, 65, 30, 80, 55, 90, 45, 60, 20, 75, 50, 65].map((h, i) => (
                                        <div key={i} className="w-4 bg-indigo-500/50 hover:bg-indigo-500 transition-colors rounded-t" style={{ height: `${h}%` }}></div>
                                    ))}
                                </div>
                            </div>
                        </Card>

                        <Card className="p-0 border-slate-700 bg-slate-800/40 h-80 flex flex-col">
                            <div className="p-4 border-b border-slate-700">
                                <h3 className="font-bold text-white flex items-center gap-2">
                                    <Target className="h-4 w-4" /> Top Attacked Assets
                                </h3>
                            </div>
                            <div className="p-4 space-y-3">
                                {[
                                    { name: 'auth-server-01', count: 142, risk: 'Critical' },
                                    { name: 'web-gateway-prod', count: 89, risk: 'High' },
                                    { name: 'db-cluster-04', count: 56, risk: 'Medium' },
                                    { name: 'workstation-dev-22', count: 12, risk: 'Low' }
                                ].map((asset, i) => (
                                    <div key={i} className="flex justify-between items-center p-2 bg-slate-900/50 rounded hover:bg-slate-700/50 transition-colors cursor-pointer">
                                        <span className="font-mono text-sm text-indigo-300">{asset.name}</span>
                                        <div className="flex items-center gap-3">
                                            <span className="text-white font-bold">{asset.count}</span>
                                            <Badge variant={asset.risk === 'Critical' ? 'error' : asset.risk === 'High' ? 'warning' : 'outline'}>
                                                {asset.risk}
                                            </Badge>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </Card>
                    </div>
                </div>
            )}
            {explanation && (
                <ExplanationModal
                    explanation={explanation}
                    onClose={() => setExplanation(null)}
                />
            )}
        </div>
    );
};
