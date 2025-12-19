import { useState } from 'react';
import { Sidebar } from '../components/layout/Sidebar';
import { Header } from '../components/layout/Header';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { FileText, Download, Calendar, Search, Trash2, Clock, Info, X, ChevronRight, Settings } from 'lucide-react';
import { Badge } from '../components/ui/Badge';
import { getReportHistory, type ReportRecord, exportToCSV, exportToJSON, clearReportHistory, deleteReport } from '../services/exportService';
import { getSchedules, saveSchedule, deleteSchedule, type ScheduledReport, createMockSchedule } from '../services/schedulingService';

export function ReportingPage() {
    const [reports, setReports] = useState<ReportRecord[]>(() => getReportHistory());
    const [schedules, setSchedules] = useState<ScheduledReport[]>(() => {
        const current = getSchedules();
        if (current.length === 0) {
            createMockSchedule(3);
            return getSchedules();
        }
        return current;
    });
    const [searchQuery, setSearchQuery] = useState('');
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [selectedReport, setSelectedReport] = useState<ReportRecord | null>(null);
    const [isScheduling, setIsScheduling] = useState(false);
    const [activeTab, setActiveTab] = useState<'history' | 'schedules'>('history');
    const [newSchedule, setNewSchedule] = useState<Partial<ScheduledReport>>({
        freq: 'Daily',
        type: 'PDF',
        status: 'Active',
        recipients: ['admin@enterprise.local']
    });

    const handleCreateSchedule = () => {
        if (!newSchedule.name) {
            alert('Please enter a schedule name.');
            return;
        }
        const schedule: ScheduledReport = {
            id: Math.random().toString(36).substring(2, 9),
            name: newSchedule.name,
            freq: (newSchedule.freq as any) || 'Daily',
            type: (newSchedule.type as any) || 'PDF',
            query: newSchedule.query || 'index=main | stats count by host',
            recipients: newSchedule.recipients || ['admin@enterprise.local'],
            status: 'Active'
        };
        saveSchedule(schedule);
        setSchedules(getSchedules());
        setIsScheduling(false);
        setNewSchedule({ freq: 'Daily', type: 'PDF', status: 'Active', recipients: ['admin@enterprise.local'] });
    };

    const filteredReports = reports.filter(report =>
        report.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const refreshHistory = () => {
        setReports(getReportHistory());
    };

    const handleClearAll = () => {
        if (confirm('Are you sure you want to clear all report history?')) {
            clearReportHistory();
            refreshHistory();
        }
    };

    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
    const [isComparing, setIsComparing] = useState(false);

    const toggleSelect = (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        const next = new Set(selectedIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        setSelectedIds(next);
    };

    const handleBulkDelete = () => {
        if (confirm(`Are you sure you want to delete ${selectedIds.size} reports?`)) {
            selectedIds.forEach(id => deleteReport(id));
            setSelectedIds(new Set());
            refreshHistory();
        }
    };

    const handleDelete = (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        deleteReport(id);
        refreshHistory();
        if (selectedReport?.id === id) setSelectedReport(null);
    };

    const handleSampleExport = () => {
        const sampleData = [
            { timestamp: new Date().toISOString(), event: 'Login', user: 'admin', status: 'success' },
            { timestamp: new Date().toISOString(), event: 'Search', user: 'analyst_1', status: 'failure' },
            { timestamp: new Date().toISOString(), event: 'Export', user: 'admin', status: 'success' },
        ];
        exportToCSV(sampleData, 'Sample_System_Audit', { query: 'index=audit | stats count by user', user: 'admin', timeRange: 'Last 24h' });
        setTimeout(refreshHistory, 500);
    };

    const handleRedownload = (report: ReportRecord, e: React.MouseEvent) => {
        e.stopPropagation();
        const dummy = [{ message: "Re-downloading archived report: " + report.name }];
        const meta = { query: report.query, user: report.user, timeRange: report.timeRange };
        if (report.type === 'CSV') exportToCSV(dummy, report.name, meta);
        else if (report.type === 'JSON') exportToJSON(dummy, report.name, meta);
        else alert('PDF Re-download would fetch from Enterprise Blob Storage.');
    };

    return (
        <div className="flex h-screen bg-slate-950 text-slate-200 font-sans selection:bg-brand-500/30">
            <Sidebar isOpen={isSidebarOpen} setIsOpen={setIsSidebarOpen} isMobile={false} />
            <div className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
                <Header openAlerts={() => { }} toggleSidebar={() => setIsSidebarOpen(!isSidebarOpen)} />

                <main className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-hide">
                    {/* Tab Navigation */}
                    <div className="flex items-center gap-1 border-b border-slate-800">
                        <button
                            className={`px-4 py-2 text-xs font-bold uppercase tracking-widest transition-all ${activeTab === 'history' ? 'text-brand-400 border-b-2 border-brand-500 bg-brand-500/5' : 'text-slate-500 hover:text-slate-300'}`}
                            onClick={() => setActiveTab('history')}
                        >
                            Export History
                        </button>
                        <button
                            className={`px-4 py-2 text-xs font-bold uppercase tracking-widest transition-all ${activeTab === 'schedules' ? 'text-brand-400 border-b-2 border-brand-500 bg-brand-500/5' : 'text-slate-500 hover:text-slate-300'}`}
                            onClick={() => setActiveTab('schedules')}
                        >
                            Automation Schedules
                        </button>
                    </div>

                    {/* Controls Header */}
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <h1 className="text-xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
                                {activeTab === 'history' ? 'Reporting Engine' : 'Automation Center'}
                            </h1>
                            <p className="text-xs text-slate-500 mt-1">
                                {activeTab === 'history' ? 'Manage and track enterprise data exports' : 'Configure recurring automated reporting cycles'}
                            </p>
                        </div>
                        <div className="relative flex-1 max-w-md">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                            <input
                                type="text"
                                placeholder={`Search ${activeTab === 'history' ? 'reports' : 'schedules'}...`}
                                className="w-full bg-slate-900/50 border border-slate-800 rounded-lg pl-10 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500/50 transition-all font-mono"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                        <div className="flex items-center gap-2">
                            {activeTab === 'history' ? (
                                <>
                                    <Button variant="ghost" size="sm" onClick={handleClearAll} className="text-slate-500 hover:text-red-400 font-bold tracking-widest text-[10px]">
                                        <Trash2 className="h-4 w-4 mr-2" />
                                        PURGE ALL
                                    </Button>
                                    <Button variant="primary" size="sm" onClick={handleSampleExport} className="font-bold tracking-widest text-[10px]">
                                        <FileText className="h-4 w-4 mr-2" />
                                        CUSTOM EXPORT
                                    </Button>
                                </>
                            ) : (
                                <Button variant="primary" size="sm" onClick={() => setIsScheduling(true)} className="font-bold tracking-widest text-[10px]">
                                    <Clock className="h-4 w-4 mr-2" />
                                    NEW SCHEDULE
                                </Button>
                            )}
                        </div>
                    </div>

                    {/* Bulk Action Bar */}
                    {selectedIds.size > 0 && (
                        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-slate-900 border border-brand-500/50 shadow-2xl shadow-brand-500/20 px-6 py-3 rounded-2xl flex items-center gap-6 z-[60] animate-in fade-in slide-in-from-bottom-4 duration-300">
                            <span className="text-xs font-bold text-brand-400 uppercase tracking-widest">{selectedIds.size} Reports Selected</span>
                            <div className="h-4 w-[1px] bg-slate-800" />
                            <div className="flex items-center gap-2">
                                {selectedIds.size === 2 && (
                                    <Button variant="outline" size="sm" className="h-8 border-brand-500/30 text-brand-400 hover:bg-brand-500/10 font-bold text-[10px] tracking-widest" onClick={() => setIsComparing(true)}>
                                        COMPARE DELTA
                                    </Button>
                                )}
                                <Button variant="ghost" size="sm" className="h-8 text-slate-400 hover:text-red-400 font-bold text-[10px] tracking-widest" onClick={handleBulkDelete}>
                                    <Trash2 className="h-3.5 w-3.5 mr-2" />
                                    PURGE SELECTED
                                </Button>
                                <Button variant="ghost" size="sm" className="h-8 text-slate-400 hover:text-slate-200" onClick={() => setSelectedIds(new Set())}>
                                    <X className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    )}

                    {/* Content Grid */}
                    {activeTab === 'history' ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {filteredReports.map((report) => (
                                <Card
                                    key={report.id}
                                    onClick={() => setSelectedReport(report)}
                                    className={`p-4 bg-slate-900/40 border-slate-800/50 hover:border-brand-500/30 transition-all group relative overflow-hidden cursor-pointer ${selectedReport?.id === report.id ? 'ring-1 ring-brand-500/50 bg-brand-500/5' : ''} ${selectedIds.has(report.id) ? 'border-brand-500/50 bg-brand-500/5' : ''}`}
                                >
                                    <div className="absolute top-2 left-2 z-20">
                                        <input
                                            type="checkbox"
                                            checked={selectedIds.has(report.id)}
                                            onChange={() => { }}
                                            onClick={(e) => toggleSelect(report.id, e)}
                                            className="h-4 w-4 rounded border-slate-700 bg-slate-800 text-brand-500 focus:ring-brand-500/50"
                                        />
                                    </div>
                                    <div className="absolute top-0 right-0 p-2 opacity-10 rotate-12 group-hover:rotate-0 transition-transform">
                                        <FileText className="h-20 w-20 text-slate-700" />
                                    </div>
                                    <div className="flex items-start justify-between relative z-10">
                                        <div className="p-2 bg-slate-800/50 rounded-lg group-hover:bg-brand-500/10 transition-colors text-brand-400">
                                            <FileText className="h-6 w-6" />
                                        </div>
                                        <div className="flex gap-1">
                                            <Badge
                                                variant={report.status === 'Ready' ? 'success' : report.status === 'Expired' ? 'outline' : 'error'}
                                            >
                                                {report.status}
                                            </Badge>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-6 w-6 p-0 text-slate-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={(e) => handleDelete(report.id, e)}
                                            >
                                                <Trash2 className="h-3 w-3" />
                                            </Button>
                                        </div>
                                    </div>
                                    <div className="mt-4 relative z-10">
                                        <h3 className="text-sm font-semibold truncate" title={report.name}>{report.name}</h3>
                                        <div className="mt-2 flex flex-wrap items-center gap-3 text-[10px] text-slate-500 font-medium tracking-tight">
                                            <span className="flex items-center"><Calendar className="h-3 w-3 mr-1" /> {report.timestamp}</span>
                                            <span className="px-1.5 py-0.5 rounded-md bg-slate-800/50 text-slate-400 uppercase">{report.type}</span>
                                            <span>{report.size}</span>
                                        </div>
                                    </div>
                                    <div className="mt-4 pt-4 border-t border-slate-800/50 flex justify-between items-center relative z-10">
                                        <span className="text-[9px] text-slate-600 font-mono tracking-tighter truncate max-w-[120px]">
                                            USER: {report.user || 'admin'}
                                        </span>
                                        <div className="flex gap-1">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-8 text-slate-400 hover:text-brand-400 font-bold tracking-widest text-[10px]"
                                                onClick={(e) => handleRedownload(report, e)}
                                            >
                                                <Download className="h-4 w-4 mr-2" />
                                                ARCHIVE
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-8 text-slate-400 hover:text-emerald-400 font-bold tracking-widest text-[10px]"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    if (report.query) {
                                                        window.location.href = `/search?q=${encodeURIComponent(report.query)}`;
                                                    }
                                                }}
                                            >
                                                RE-RUN
                                            </Button>
                                        </div>
                                    </div>
                                </Card>
                            ))}
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {schedules.map((schedule) => (
                                <Card key={schedule.id} className="p-5 bg-slate-900/40 border-slate-800/50 hover:border-amber-500/30 transition-all relative">
                                    <div className="flex items-center justify-between">
                                        <div className="p-2 bg-amber-500/10 rounded-lg text-amber-500">
                                            <Clock className="h-5 w-5" />
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Badge variant={schedule.status === 'Active' ? 'success' : 'outline'}>{schedule.status}</Badge>
                                            <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-slate-500 hover:text-red-400" onClick={() => { deleteSchedule(schedule.id); setSchedules(getSchedules()); }}>
                                                <Trash2 className="h-3 w-3" />
                                            </Button>
                                        </div>
                                    </div>
                                    <div className="mt-4">
                                        <h3 className="text-sm font-bold">{schedule.name}</h3>
                                        <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-widest font-bold">{schedule.freq} Automated Trigger</p>
                                    </div>
                                    <div className="mt-4 p-3 bg-black/20 rounded-lg border border-slate-800/50 text-[10px] font-mono text-slate-400 truncate">
                                        {schedule.query}
                                    </div>
                                    <div className="mt-4 pt-4 border-t border-slate-800/50 flex justify-between items-center">
                                        <div className="text-[10px] text-slate-600">
                                            {schedule.recipients[0]}
                                        </div>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            className="h-7 text-[10px] font-bold tracking-widest text-amber-500 hover:text-amber-400 bg-amber-500/5"
                                            onClick={() => {
                                                const dummyData = [{ timestamp: new Date().toISOString(), schedule: schedule.name, trigger: 'manual_debug' }];
                                                if (schedule.type === 'CSV') exportToCSV(dummyData, `Scheduled_${schedule.name}`, { query: schedule.query, user: 'system_scheduler' });
                                                else if (schedule.type === 'JSON') exportToJSON(dummyData, `Scheduled_${schedule.name}`, { query: schedule.query, user: 'system_scheduler' });
                                                else alert('PDF Task triggered. Check Enterprise Job logs.');
                                                refreshHistory();
                                                alert(`Triggered ${schedule.name} automation successfully.`);
                                            }}
                                        >
                                            TRIGGER NOW
                                        </Button>
                                    </div>
                                </Card>
                            ))}
                        </div>
                    )}

                    {((activeTab === 'history' && filteredReports.length === 0) || (activeTab === 'schedules' && schedules.length === 0)) && (
                        <div className="flex flex-col items-center justify-center py-24 text-slate-500 bg-slate-900/20 rounded-xl border border-dashed border-slate-800/50">
                            <FileText className="h-12 w-12 mb-4 opacity-10" />
                            <p className="text-sm font-medium">No {activeTab} found.</p>
                            <p className="text-xs mt-1 text-slate-600">
                                {activeTab === 'history' ? 'Export data from Dashboard to see them here.' : 'Create an automation schedule to get started.'}
                            </p>
                        </div>
                    )}
                </main>

                {/* Report Details Sidebar overlay */}
                {selectedReport && (
                    <div className="absolute inset-y-0 right-0 w-96 bg-slate-900 border-l border-slate-800 shadow-2xl z-50 transform transition-transform duration-300 flex flex-col">
                        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
                            <h2 className="text-sm font-bold flex items-center gap-2">
                                <Info className="h-4 w-4 text-brand-400" />
                                Report Intelligence
                            </h2>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0" onClick={() => setSelectedReport(null)}>
                                <X className="h-4 w-4" />
                            </Button>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6 space-y-6">
                            <div className="space-y-4">
                                <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Name</label>
                                    <p className="text-sm font-semibold mt-1 text-white">{selectedReport.name}</p>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Format</label>
                                        <p className="text-sm font-semibold mt-1 text-white">{selectedReport.type}</p>
                                    </div>
                                    <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Size</label>
                                        <p className="text-sm font-semibold mt-1 text-white">{selectedReport.size}</p>
                                    </div>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">User</label>
                                        <p className="text-sm font-semibold mt-1 text-white">{selectedReport.user || 'N/A'}</p>
                                    </div>
                                    <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Time Range</label>
                                        <p className="text-sm font-semibold mt-1 text-white">{selectedReport.timeRange || 'All Time'}</p>
                                    </div>
                                </div>
                                <div className="p-4 bg-slate-950 rounded-xl border border-slate-800">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Generation Query (SPL)</label>
                                    <div className="mt-2 p-3 bg-black/50 rounded-lg text-[11px] font-mono text-emerald-400 border border-emerald-500/20 leading-relaxed overflow-x-auto">
                                        {selectedReport.query || 'manual_trigger | fields *'}
                                    </div>
                                </div>
                            </div>

                            <div className="pt-6 border-t border-slate-800">
                                <h3 className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-4">Actions</h3>
                                <div className="space-y-2">
                                    <Button variant="primary" className="w-full justify-between group" onClick={(e) => handleRedownload(selectedReport!, e)}>
                                        Download Now
                                        <Download className="h-4 w-4 group-hover:translate-y-0.5 transition-transform" />
                                    </Button>
                                    <Button variant="outline" className="w-full justify-between" onClick={(e) => handleDelete(selectedReport.id, e)}>
                                        Purge Archive
                                        <Trash2 className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        </div>
                        <div className="p-4 bg-slate-900 border-t border-slate-800">
                            <div className="flex items-center gap-2 text-[10px] text-slate-600">
                                <Badge variant="outline" className="h-4 px-1 text-[8px]">SECURITY</Badge>
                                <span>This report is classification: INTERNAL</span>
                            </div>
                        </div>
                    </div>
                )}

                {/* Scheduling Sidebar overlay */}
                {isScheduling && (
                    <div className="absolute inset-y-0 right-0 w-96 bg-slate-900 border-l border-slate-800 shadow-2xl z-50 transform transition-transform duration-300 flex flex-col">
                        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
                            <h2 className="text-sm font-bold flex items-center gap-2">
                                <Clock className="h-4 w-4 text-amber-400" />
                                Schedule Automation
                            </h2>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0" onClick={() => setIsScheduling(false)}>
                                <X className="h-4 w-4" />
                            </Button>
                        </div>
                        <div className="flex-1 p-6 space-y-6 overflow-y-auto">
                            <div className="p-4 bg-amber-500/5 border border-amber-500/10 rounded-xl space-y-2">
                                <p className="text-xs text-amber-200 font-medium">Enterprise Automation</p>
                                <p className="text-[10px] text-amber-500/70 leading-relaxed">Configure recurring reports to be generated and delivered to designated stakeholders automatically.</p>
                            </div>

                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Schedule Name</label>
                                    <input
                                        type="text"
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-amber-500/50"
                                        placeholder="e.g. Daily Security Audit"
                                        value={newSchedule.name || ''}
                                        onChange={(e) => setNewSchedule(prev => ({ ...prev, name: e.target.value }))}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Query (SPL)</label>
                                    <textarea
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-4 py-2 text-sm font-mono h-24 focus:outline-none focus:ring-1 focus:ring-amber-500/50"
                                        placeholder="index=main | stats count by host"
                                        value={newSchedule.query || ''}
                                        onChange={(e) => setNewSchedule(prev => ({ ...prev, query: e.target.value }))}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Frequency</label>
                                    <div className="grid grid-cols-3 gap-2">
                                        {(['Daily', 'Weekly', 'Monthly'] as const).map(f => (
                                            <Button
                                                key={f}
                                                variant="outline"
                                                size="sm"
                                                className={`text-[10px] font-bold ${newSchedule.freq === f ? 'bg-amber-500/10 border-amber-500/50 text-amber-500' : ''}`}
                                                onClick={() => setNewSchedule(prev => ({ ...prev, freq: f }))}
                                            >
                                                {f}
                                            </Button>
                                        ))}
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Report Type</label>
                                    <div className="grid grid-cols-3 gap-2 text-[10px]">
                                        {(['PDF', 'CSV', 'JSON'] as const).map(t => (
                                            <Button
                                                key={t}
                                                variant="outline"
                                                size="sm"
                                                className={`text-[10px] font-bold ${newSchedule.type === t ? 'bg-amber-500/10 border-amber-500/50 text-amber-500' : ''}`}
                                                onClick={() => setNewSchedule(prev => ({ ...prev, type: t }))}
                                            >
                                                {t}
                                            </Button>
                                        ))}
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Delivery Channel</label>
                                    <div className="space-y-2">
                                        <div className="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800 opacity-50 cursor-not-allowed">
                                            <span className="text-xs text-slate-300">Email SMTP</span>
                                            <Settings className="h-3 w-3 text-slate-600" />
                                        </div>
                                        <div className="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800 opacity-50 cursor-not-allowed">
                                            <span className="text-xs text-slate-300">Slack Webhook</span>
                                            <ChevronRight className="h-3 w-3 text-slate-600" />
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="p-6 border-t border-slate-800 bg-slate-900/50">
                            <Button variant="primary" className="w-full shadow-lg shadow-amber-500/20 bg-amber-600 hover:bg-amber-500 text-white" onClick={handleCreateSchedule}>
                                Create Schedule
                            </Button>
                        </div>
                    </div>
                )}
                {/* Forensic Comparison Modal */}
                {isComparing && selectedIds.size === 2 && (
                    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-300">
                        <Card className="w-full max-w-4xl bg-slate-900 border-slate-800 shadow-2xl relative flex flex-col max-h-[90vh]">
                            <div className="p-4 border-b border-slate-800 flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <div className="p-2 bg-brand-500/10 rounded text-brand-400">
                                        <Info className="h-5 w-5" />
                                    </div>
                                    <h2 className="text-lg font-bold">Forensic Delta Comparison</h2>
                                </div>
                                <Button variant="ghost" size="sm" onClick={() => setIsComparing(false)}>
                                    <X className="h-5 w-5" />
                                </Button>
                            </div>

                            <div className="flex-1 overflow-y-auto p-6">
                                <div className="grid grid-cols-2 gap-8">
                                    {[...selectedIds].map((id, idx) => {
                                        const r = reports.find(x => x.id === id);
                                        if (!r) return null;
                                        return (
                                            <div key={id} className="space-y-4">
                                                <div className="flex flex-col">
                                                    <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Report {idx + 1}</span>
                                                    <h3 className="text-sm font-bold text-white truncate" title={r.name}>{r.name}</h3>
                                                </div>

                                                <div className="p-4 bg-slate-950 rounded-xl border border-slate-800 space-y-3">
                                                    <div>
                                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Metadata Context</label>
                                                        <div className="mt-2 space-y-1">
                                                            <div className="flex justify-between text-xs">
                                                                <span className="text-slate-500">User</span>
                                                                <span className="text-slate-300 font-medium">{r.user || 'admin'}</span>
                                                            </div>
                                                            <div className="flex justify-between text-xs">
                                                                <span className="text-slate-500">Time Range</span>
                                                                <span className="text-slate-300 font-medium">{r.timeRange || 'All Time'}</span>
                                                            </div>
                                                            <div className="flex justify-between text-xs">
                                                                <span className="text-slate-500">Timestamp</span>
                                                                <span className="text-slate-300 font-medium">{r.timestamp}</span>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div>
                                                        <label className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">SPL Logic</label>
                                                        <div className="mt-2 p-3 bg-black/40 rounded border border-slate-800 font-mono text-[10px] text-emerald-400 leading-relaxed overflow-x-auto whitespace-pre-wrap">
                                                            {r.query || 'manual_trigger | fields *'}
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>

                                <div className="mt-8 p-4 bg-brand-500/5 border border-brand-500/20 rounded-xl">
                                    <h4 className="text-xs font-bold text-brand-400 uppercase tracking-widest flex items-center gap-2 mb-3">
                                        <Search className="h-3 w-3" />
                                        Delta Analysis (Inferred)
                                    </h4>
                                    <div className="space-y-2">
                                        <div className="flex items-start gap-3">
                                            <div className="h-1.5 w-1.5 rounded-full bg-emerald-500 mt-1.5" />
                                            <p className="text-xs text-slate-400"><span className="text-slate-200 font-semibold">Query Parity</span>: Logic remains structurally identical. Minor predicate changes detected in time boundaries.</p>
                                        </div>
                                        <div className="flex items-start gap-3">
                                            <div className="h-1.5 w-1.5 rounded-full bg-amber-500 mt-1.5" />
                                            <p className="text-xs text-slate-400"><span className="text-slate-200 font-semibold">Integrity Drift</span>: Reports were generated by different identities ({reports.find(x => x.id === [...selectedIds][0])?.user || 'admin'} vs {reports.find(x => x.id === [...selectedIds][1])?.user || 'admin'}).</p>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="p-4 border-t border-slate-800 bg-slate-900/50 flex justify-end gap-3">
                                <Button variant="ghost" onClick={() => setIsComparing(false)}>Dismiss</Button>
                                <Button variant="primary" onClick={() => setIsComparing(false)}>Log to Case File</Button>
                            </div>
                        </Card>
                    </div>
                )}
            </div>
        </div>
    );
}
