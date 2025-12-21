import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { Settings, Play, Pause, FileText, Globe, Terminal } from 'lucide-react';

interface InputItemProps {
    type: 'TCP' | 'UDP' | 'HTTP' | 'Script' | 'File';
    name: string;
    source: string;
    status: 'Running' | 'Stopped' | 'Error';
    port?: number;
}

function InputItem({ type, name, source, status, port }: InputItemProps) {
    const iconMap = {
        'TCP': Globe,
        'UDP': Globe,
        'HTTP': Globe,
        'Script': Terminal,
        'File': FileText
    };
    const Icon = iconMap[type];

    return (
        <div className="flex items-center justify-between p-4 bg-slate-900/50 border border-slate-800 rounded-lg group hover:border-slate-700 transition-colors" role="listitem">
            <div className="flex items-center gap-4">
                <div className="h-10 w-10 rounded-lg bg-slate-800 flex items-center justify-center">
                    <Icon className="h-5 w-5 text-slate-400 group-hover:text-brand-400 transition-colors" />
                </div>
                <div>
                    <h4 className="font-medium text-white">{name}</h4>
                    <p className="text-sm text-slate-500">
                        {type} {port && `• Port ${port}`} • {source}
                    </p>
                </div>
            </div>

            <div className="flex items-center gap-4">
                <Badge variant={status === 'Running' ? 'success' : status === 'Stopped' ? 'warning' : 'error'}>
                    {status}
                </Badge>
                <div className="flex gap-2">
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        {status === 'Running' ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                    </Button>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <Settings className="h-4 w-4" />
                    </Button>
                </div>
            </div>
        </div>
    );
}

export function InputsList() {
    return (
        <Card>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h3 className="text-lg font-semibold text-white">Local Inputs</h3>
                    <p className="text-sm text-slate-400">Manage data ingestion points for this node.</p>
                </div>
                <Button variant="primary" size="sm">Add New Input</Button>
            </div>

            <div className="space-y-3">
                <InputItem type="TCP" name="Syslog Stream" port={514} source="0.0.0.0" status="Running" />
                <InputItem type="HTTP" name="HEC Event Collector" port={8088} source="/services/collector" status="Running" />
                <InputItem type="File" name="Security Logs" source="/var/log/secure" status="Running" />
                <InputItem type="Script" name="AWS Inventory" source="./scripts/aws_inv.py" status="Stopped" />
                <InputItem type="UDP" name="Network Traps" port={162} source="0.0.0.0" status="Error" />
            </div>
        </Card>
    );
}
