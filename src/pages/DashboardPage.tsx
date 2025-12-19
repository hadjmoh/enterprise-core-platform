import { useState, useEffect } from 'react';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import {
    Cpu,
    Server,
    Shield,
    Cloud,
    HardDrive,
    Network,
    Thermometer,
    RefreshCw,
    CheckCircle,
    Plus,
    RotateCcw,
    Lock,
    Unlock,
    Bell
} from 'lucide-react';
import { useAlerts } from '../contexts/AlertContextInstance';
import { ProgressBar } from '../components/ui/ProgressBar';
import { SearchHub } from '../components/dashboard/SearchHub';
import { SystemHealthChart } from '../components/dashboard/SystemHealthChart';
import { MetricRow } from '../components/monitor/MetricRow';
import { ResourceSection } from '../components/monitor/ResourceSection';
import { ProcessTable } from '../components/monitor/ProcessTable';
import { useDashboard } from '../hooks/useDashboard';
import { DashboardGrid } from '../components/dashboard/DashboardGrid';
import { useUI } from '../hooks/useUI';

const INITIAL_DATA = Array.from({ length: 20 }, (_, i) => ({
    time: `${10 + Math.floor(i / 2)}:${(i % 2) * 30}`.padStart(5, '0'),
    cpu: 40 + Math.random() * 20,
    memory: 60 + Math.random() * 15,
}));

export function DashboardPage() {
    const { isLocked, setIsLocked } = useUI();
    const [chartData, setChartData] = useState(INITIAL_DATA);
    const { config, updateLayouts, removeWidget, resetLayout, addWidget } = useDashboard();
    const { addAlert } = useAlerts();
    const [globalSearchSpl, setGlobalSearchSpl] = useState<string>('');

    useEffect(() => {
        const interval = setInterval(() => {
            setChartData(prev => {
                const lastTime = prev[prev.length - 1].time;
                const [hours, minutes] = lastTime.split(':').map(Number);
                let newMinutes = minutes + 1;
                let newHours = hours;
                if (newMinutes >= 60) {
                    newMinutes = 0;
                    newHours = (hours + 1) % 24;
                }
                const newPoint = {
                    time: `${newHours}:${newMinutes.toString().padStart(2, '0')}`,
                    cpu: 30 + Math.random() * 40,
                    memory: 50 + Math.random() * 30,
                };
                return [...prev.slice(1), newPoint];
            });
        }, 2000);
        return () => clearInterval(interval);
    }, []);

    const [metrics, setMetrics] = useState({
        cpuTemp: 50,
        fanSpeed: 1250,
        networkIn: "28.5",
        networkOut: "14.2"
    });

    useEffect(() => {
        const interval = setInterval(() => {
            setMetrics({
                cpuTemp: 45 + Math.random() * 10,
                fanSpeed: 1200 + Math.random() * 100,
                networkIn: (25 + Math.random() * 15).toFixed(1),
                networkOut: (12 + Math.random() * 5).toFixed(1)
            });
        }, 5000);
        return () => clearInterval(interval);
    }, []);

    const { cpuTemp, fanSpeed, networkIn, networkOut } = metrics;

    const handleAddQuickWidget = () => {
        addWidget({
            title: 'New Web Traffic',
            type: 'timechart',
            spl: 'search index=web | timechart span=1m count'
        });
    };

    const handleGlobalSearch = (spl: string) => {
        setGlobalSearchSpl(spl);
        addWidget({
            title: `Investigation: ${spl.substring(0, 20)}...`,
            type: 'auto',
            spl: spl
        });
    };

    return (
        <div className="flex flex-col gap-6 pb-20">
            {/* Header */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div className="space-y-1">
                    <h1 className="text-3xl font-bold text-white tracking-tight">Enterprise Ops</h1>
                    <div className="flex items-center gap-2 text-slate-400">
                        <Badge variant="success" className="h-5">Cluster Ready</Badge>
                        <span>•</span>
                        <span>{config.name}</span>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="text-amber-500 hover:bg-amber-500/10"
                        onClick={() => addAlert({
                            title: "Security Threat",
                            message: "Multiple failed login attempts detected from IP 192.168.1.100",
                            severity: "critical",
                            action: {
                                label: "View Logs",
                                url: "/data-inputs"
                            }
                        })}
                    >
                        <Bell className="h-4 w-4 mr-2" />
                        Sim Alert
                    </Button>
                    <Button
                        variant={isLocked ? "outline" : "primary"}
                        size="sm"
                        onClick={() => setIsLocked(!isLocked)}
                    >
                        {isLocked ? (
                            <><Lock className="h-4 w-4 mr-2" /> Unlock Layout</>
                        ) : (
                            <><Unlock className="h-4 w-4 mr-2" /> Lock Layout</>
                        )}
                    </Button>
                    <Button variant="outline" size="sm" onClick={resetLayout}>
                        <RotateCcw className="h-4 w-4 mr-2" />
                        Reset
                    </Button>
                    {!isLocked && (
                        <Button variant="primary" size="sm" onClick={handleAddQuickWidget}>
                            <Plus className="h-4 w-4 mr-2" />
                            Add Widget
                        </Button>
                    )}
                    <Button variant="outline" size="sm" onClick={() => window.location.reload()}>
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Reload
                    </Button>
                </div>
            </div>

            {/* Global Search Bar */}
            {!isLocked && (
                <div className="space-y-4">
                    <SearchHub
                        initialSpl={globalSearchSpl}
                        onSearch={(spl) => {
                            addWidget({
                                title: `Query: ${spl.substring(0, 20)}...`,
                                type: 'auto',
                                spl: spl
                            });
                        }}
                        isLoading={false}
                    />
                </div>
            )}

            {/* Dynamic Dashboard Grid */}
            <DashboardGrid
                widgets={config.widgets}
                isLocked={isLocked}
                onLayoutChange={updateLayouts}
                onRemoveWidget={removeWidget}
                onGlobalSearch={handleGlobalSearch}
                onAddWidget={addWidget}
            />

            {/* Legacy System Health Section */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 opacity-50 grayscale hover:opacity-100 hover:grayscale-0 transition-all duration-500">
                <div className="md:col-span-3">
                    <SystemHealthChart data={chartData} />
                </div>
                <div className="flex flex-col gap-4">
                    <ResourceSection title="Security Status" icon={Shield} variant="glass">
                        <div className="flex flex-col gap-3">
                            <div className="flex items-center justify-between p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
                                <span className="text-emerald-400 font-medium flex items-center gap-2">
                                    <CheckCircle className="h-4 w-4" /> IDS Active
                                </span>
                            </div>
                            <MetricRow label="Firewall Rules" value="842 Active" />
                            <MetricRow label="Risk Score" value="Low (12/100)" />
                        </div>
                    </ResourceSection>
                    <ResourceSection title="Cloud Sync" icon={Cloud}>
                        <MetricRow label="Region" value="AWS-West-2" />
                        <MetricRow label="Instance" value="c6g.4xlarge" />
                    </ResourceSection>
                </div>
            </div>

            {/* Hardware Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <ResourceSection
                    title="Processor"
                    icon={Cpu}
                    mainMetric={{ label: "Usage", value: "42%", percent: 42, color: "bg-blue-500" }}
                >
                    <MetricRow label="Load" value="1.45, 1.20, 0.95" />
                </ResourceSection>

                <ResourceSection
                    title="Memory"
                    icon={Server}
                    mainMetric={{ label: "Used", value: "12.4 GB", percent: 64, color: "bg-purple-500" }}
                >
                    <MetricRow label="Available" value="11.4 GB" />
                </ResourceSection>

                <ResourceSection title="Storage" icon={HardDrive}>
                    <div className="space-y-4">
                        <div>
                            <ProgressBar value={84} color="bg-red-500" />
                            <p className="text-xs text-slate-500 mt-1">420GB used</p>
                        </div>
                    </div>
                </ResourceSection>

                <ResourceSection title="Network" icon={Network}>
                    <div className="grid grid-cols-2 gap-2">
                        <div className="p-2 bg-slate-900 rounded text-center">
                            <p className="text-[10px] text-slate-500">IN</p>
                            <p className="text-sm font-bold text-emerald-400">{networkIn} MB</p>
                        </div>
                        <div className="p-2 bg-slate-900 rounded text-center">
                            <p className="text-[10px] text-slate-500">OUT</p>
                            <p className="text-sm font-bold text-blue-400">{networkOut} MB</p>
                        </div>
                    </div>
                </ResourceSection>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <ProcessTable />
                </div>
                <ResourceSection title="Environmental" icon={Thermometer} variant="neon">
                    <div className="flex justify-between items-center px-2">
                        <span className="text-2xl font-bold">{Math.round(cpuTemp)}°C</span>
                        <span className="text-sm text-slate-400">{Math.round(fanSpeed)} RPM</span>
                    </div>
                </ResourceSection>
            </div>
        </div>
    );
}
