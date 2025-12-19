import { InputsList } from '../components/inputs/InputsList';
import { ForwarderStatus } from '../components/inputs/ForwarderStatus';
import { Button } from '../components/ui/Button';
import { Download } from 'lucide-react';

export function DataInputsPage() {
    return (
        <div className="flex flex-col gap-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-white tracking-tight">Data Inputs</h1>
                    <p className="text-slate-400">Configure how data enters the Enterprise Platform.</p>
                </div>
                <Button variant="outline">
                    <Download className="h-4 w-4 mr-2" />
                    Download Universal Forwarder
                </Button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <InputsList />
                </div>
                <div>
                    <ForwarderStatus />
                </div>
            </div>
        </div>
    );
}
