import { Badge } from '../ui/Badge';
import { Card } from '../ui/Card';

interface Process {
    pid: number;
    name: string;
    user: string;
    cpu: number;
    mem: number;
    status: 'Running' | 'Sleeping' | 'Zombie';
}

export function ProcessTable() {
    // Mock Data
    const processes: Process[] = [
        { pid: 1192, name: 'enterprise-core', user: 'root', cpu: 12.4, mem: 4.2, status: 'Running' },
        { pid: 832, name: 'indexer-service', user: 'splunk', cpu: 8.1, mem: 12.5, status: 'Running' },
        { pid: 221, name: 'postgres-db', user: 'postgres', cpu: 1.2, mem: 6.8, status: 'Sleeping' },
        { pid: 4421, name: 'node-exporter', user: 'root', cpu: 0.5, mem: 0.2, status: 'Running' },
        { pid: 911, name: 'defunct_proc', user: 'root', cpu: 0.0, mem: 0.0, status: 'Zombie' },
    ];

    return (
        <Card>
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">Top Processes</h3>
                <Badge variant="outline">Live</Badge>
            </div>
            <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                    <thead>
                        <tr className="border-b border-slate-800 text-slate-400">
                            <th className="pb-2 font-medium">PID</th>
                            <th className="pb-2 font-medium">Name</th>
                            <th className="pb-2 font-medium">User</th>
                            <th className="pb-2 font-medium">CPU%</th>
                            <th className="pb-2 font-medium">MEM%</th>
                            <th className="pb-2 font-medium">Status</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {processes.map((p) => (
                            <tr key={p.pid} className="group hover:bg-slate-800/30 transition-colors">
                                <td className="py-2 text-slate-500 font-mono">{p.pid}</td>
                                <td className="py-2 text-slate-200 font-medium">{p.name}</td>
                                <td className="py-2 text-slate-400">{p.user}</td>
                                <td className="py-2 text-slate-300">{p.cpu}%</td>
                                <td className="py-2 text-slate-300">{p.mem}%</td>
                                <td className="py-2">
                                    <Badge
                                        variant={
                                            p.status === 'Running' ? 'success' :
                                                p.status === 'Zombie' ? 'error' : 'default'
                                        }
                                        className="scale-90 origin-left"
                                    >
                                        {p.status}
                                    </Badge>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </Card>
    );
}
