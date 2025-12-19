import { useState } from 'react';
import { Sidebar } from '../components/layout/Sidebar';
import { Header } from '../components/layout/Header';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { FileText, Download, Calendar, Filter, Search } from 'lucide-react';
import { Badge } from '../components/ui/Badge';

const MOCK_REPORTS = [
    { id: '1', name: 'Executive Summary - Q4 2023', type: 'PDF', date: '2023-12-01', size: '2.4 MB', status: 'Ready' },
    { id: '2', name: 'System Performance Audit', type: 'CSV', date: '2023-11-28', size: '15.8 MB', status: 'Processing' },
    { id: '3', name: 'User Activity Log - Weekly', type: 'JSON', date: '2023-12-05', size: '842 KB', status: 'Ready' },
    { id: '4', name: 'Security Incident Report', type: 'PDF', date: '2023-12-04', size: '1.1 MB', status: 'Ready' },
    { id: '5', name: 'Resource Utilization Forecast', type: 'CSV', date: '2023-11-30', size: '4.2 MB', status: 'Error' },
];

export function ReportingPage() {
    const [reports] = useState(MOCK_REPORTS);
    const [searchQuery, setSearchQuery] = useState('');
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);

    const filteredReports = reports.filter(report =>
        report.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex h-screen bg-slate-950 text-slate-200 font-sans selection:bg-brand-500/30">
            <Sidebar isOpen={isSidebarOpen} setIsOpen={setIsSidebarOpen} isMobile={false} />
            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Header openAlerts={() => { }} />

                <main className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-hide">
                    {/* Controls Header */}
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <h1 className="text-xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">Reporting Engine</h1>
                        <div className="relative flex-1 max-w-md">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                            <input
                                type="text"
                                placeholder="Search reports..."
                                className="w-full bg-slate-900/50 border border-slate-800 rounded-lg pl-10 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500/50 transition-all"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                        <div className="flex items-center gap-2">
                            <Button variant="outline" size="sm">
                                <Calendar className="h-4 w-4 mr-2" />
                                Date Range
                            </Button>
                            <Button variant="outline" size="sm">
                                <Filter className="h-4 w-4 mr-2" />
                                Filter
                            </Button>
                            <Button variant="primary" size="sm">
                                <FileText className="h-4 w-4 mr-2" />
                                Generate Report
                            </Button>
                        </div>
                    </div>

                    {/* Report Grid */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {filteredReports.map((report) => (
                            <Card key={report.id} className="p-4 bg-slate-900/40 border-slate-800/50 hover:border-brand-500/30 transition-all group">
                                <div className="flex items-start justify-between">
                                    <div className="p-2 bg-slate-800/50 rounded-lg group-hover:bg-brand-500/10 transition-colors">
                                        <FileText className="h-6 w-6 text-brand-400" />
                                    </div>
                                    <Badge
                                        variant={report.status === 'Ready' ? 'success' : report.status === 'Processing' ? 'outline' : 'error'}
                                    >
                                        {report.status}
                                    </Badge>
                                </div>
                                <div className="mt-4">
                                    <h3 className="text-sm font-semibold truncate" title={report.name}>{report.name}</h3>
                                    <div className="mt-2 flex items-center gap-3 text-[10px] text-slate-500 font-medium tracking-tight">
                                        <span className="flex items-center"><Calendar className="h-3 w-3 mr-1" /> {report.date}</span>
                                        <span className="flex items-center">Format: {report.type}</span>
                                        <span>{report.size}</span>
                                    </div>
                                </div>
                                <div className="mt-4 pt-4 border-t border-slate-800/50 flex justify-end">
                                    <Button variant="ghost" size="sm" className="h-8 text-slate-400 hover:text-brand-400">
                                        <Download className="h-4 w-4 mr-2" />
                                        Download
                                    </Button>
                                </div>
                            </Card>
                        ))}
                    </div>

                    {filteredReports.length === 0 && (
                        <div className="flex flex-col items-center justify-center py-20 text-slate-500">
                            <FileText className="h-12 w-12 mb-4 opacity-20" />
                            <p className="text-sm">No reports found matching your search.</p>
                        </div>
                    )}
                </main>
            </div>
        </div>
    );
}
