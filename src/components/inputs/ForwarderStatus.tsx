import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Server, Wifi, Activity } from 'lucide-react';

export function ForwarderStatus() {
    return (
        <Card variant="glass" className="h-full">
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                    <Server className="h-5 w-5 text-brand-400" />
                    <h3 className="text-lg font-semibold text-white">Forwarder Fleet</h3>
                </div>
                <Badge variant="success" className="animate-pulse">Live</Badge>
            </div>

            <div className="grid grid-cols-3 gap-4 mb-6">
                <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-center">
                    <div className="text-2xl font-bold text-emerald-400">142</div>
                    <div className="text-xs text-emerald-300/70 uppercase font-medium mt-1">Active</div>
                </div>
                <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20 text-center">
                    <div className="text-2xl font-bold text-amber-400">3</div>
                    <div className="text-xs text-amber-300/70 uppercase font-medium mt-1">Warning</div>
                </div>
                <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-center">
                    <div className="text-2xl font-bold text-red-400">1</div>
                    <div className="text-xs text-red-300/70 uppercase font-medium mt-1">Missing</div>
                </div>
            </div>

            <div className="space-y-4">
                <h4 className="text-sm font-medium text-slate-400 uppercase tracking-wider">Recent Activity</h4>
                <div className="space-y-3">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="flex items-center gap-3 text-sm">
                            <Activity className="h-4 w-4 text-slate-500" />
                            <span className="text-slate-300">host-us-east-{i} sent 45 MB of logs</span>
                            <span className="ml-auto text-slate-600">Now</span>
                        </div>
                    ))}
                    <div className="flex items-center gap-3 text-sm">
                        <Wifi className="h-4 w-4 text-amber-400" />
                        <span className="text-amber-200/80">host-eu-west-9 connection unstable</span>
                        <span className="ml-auto text-slate-600">2m ago</span>
                    </div>
                </div>
            </div>
        </Card>
    );
}
