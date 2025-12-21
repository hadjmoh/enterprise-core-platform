import React, { useState, useEffect } from 'react';
import {
    ShieldAlert,
    Search,
    Filter,
    CheckCircle,
    Clock,
    User,
    MessageSquare,
    History,
    MoreVertical,
    ChevronRight,
    AlertTriangle,
    Shield,
    Server
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { Input } from '../components/ui/Input';
import { incidentService } from '../services/incidentService';
import { correlationService } from '../services/correlationService';
import { identityService } from '../services/identityService';
import type { Incident, IncidentStatus } from '../services/incidentService';
import { cn } from '../lib/utils';

const IncidentMetadataField = ({ label, children }: { label: string; children: React.ReactNode }) => (
    <div className="space-y-1">
        <span className="text-[10px] uppercase font-bold text-slate-500 tracking-wider">{label}</span>
        <div className="flex items-center gap-2 text-sm text-slate-300">
            {children}
        </div>
    </div>
);

export const IncidentReviewPage: React.FC = () => {
    const [incidents, setIncidents] = useState<Incident[]>([]);
    const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'All'>('All');
    const [newNote, setNewNote] = useState('');

    const signals = correlationService.getSignals().filter(s => selectedIncident?.sourceSignals.includes(s.id));
    const resolvedEntities = Array.from(new Set(signals.flatMap(s => s.entities || [])))
        .map(id => identityService.getEntityById(id))
        .filter(Boolean);

    useEffect(() => {
        setIncidents(incidentService.getIncidents());
        const unsubscribe = incidentService.subscribe((updatedIncidents) => {
            setIncidents(updatedIncidents);
            if (selectedIncident) {
                const updated = updatedIncidents.find(i => i.id === selectedIncident.id);
                if (updated) setSelectedIncident(updated);
            }
        });
        return unsubscribe;
    }, [selectedIncident]);

    const filteredIncidents = incidents.filter(i => {
        const matchesSearch = i.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
            i.id.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'All' || i.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const handleUpdateStatus = (status: IncidentStatus) => {
        if (selectedIncident) {
            incidentService.updateIncidentStatus(selectedIncident.id, status);
        }
    };

    const handleAssign = () => {
        if (selectedIncident) {
            incidentService.assignIncident(selectedIncident.id, 'Current Analyst');
        }
    };

    const handleAddNote = (e: React.FormEvent) => {
        e.preventDefault();
        if (selectedIncident && newNote.trim()) {
            incidentService.addNote(selectedIncident.id, newNote);
            setNewNote('');
        }
    };

    return (
        <div className="flex flex-col h-[calc(100vh-120px)] bg-slate-900 text-slate-100 overflow-hidden rounded-xl border border-slate-800">
            {/* Toolbar */}
            <div className="p-4 border-b border-slate-800 bg-slate-800/50 flex flex-wrap items-center justify-between gap-4">
                <div className="flex items-center gap-4 flex-1">
                    <div className="relative w-64">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                        <Input
                            placeholder="Search incidents..."
                            className="pl-10 bg-slate-900 border-slate-700 h-9"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                    <div className="flex items-center gap-2">
                        <Filter className="h-4 w-4 text-slate-500" />
                        <select
                            className="bg-slate-900 border border-slate-700 text-sm rounded px-2 h-9 focus:outline-none focus:ring-1 focus:ring-brand-500"
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value as any)}
                        >
                            <option value="All">All Statuses</option>
                            <option value="New">New</option>
                            <option value="Assigned">Assigned</option>
                            <option value="In Progress">In Progress</option>
                            <option value="Resolved">Resolved</option>
                            <option value="Closed">Closed</option>
                        </select>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-slate-400">
                        {filteredIncidents.length} incidents found
                    </Badge>
                </div>
            </div>

            <div className="flex flex-1 overflow-hidden">
                {/* Incident Queue */}
                <div className={cn(
                    "w-full lg:w-1/3 border-r border-slate-800 overflow-y-auto flex flex-col",
                    selectedIncident && "hidden lg:flex"
                )}>
                    {filteredIncidents.length === 0 ? (
                        <div className="flex flex-col items-center justify-center p-10 text-slate-500 italic">
                            <ShieldAlert className="h-12 w-12 mb-4 opacity-20" />
                            No incidents found.
                        </div>
                    ) : (
                        filteredIncidents.map(incident => (
                            <button
                                key={incident.id}
                                onClick={() => setSelectedIncident(incident)}
                                aria-label={`Select incident ${incident.id}: ${incident.title}`}
                                className={cn(
                                    "w-full text-left p-4 border-b border-slate-800 cursor-pointer transition-colors hover:bg-slate-800/50 relative group focus:outline-none focus:bg-slate-800 focus:ring-inset focus:ring-2 focus:ring-brand-500",
                                    selectedIncident?.id === incident.id ? "bg-brand-500/10 border-l-4 border-l-brand-500" : "border-l-4 border-l-transparent"
                                )}
                            >
                                <div className="flex justify-between items-start mb-1">
                                    <span className="text-xs font-mono text-slate-500">{incident.id}</span>
                                    <Badge variant={
                                        incident.severity === 'critical' ? 'error' :
                                            incident.severity === 'high' ? 'warning' : 'default'
                                    } className="text-[10px] px-1.5 py-0">
                                        {incident.severity.toUpperCase()}
                                    </Badge>
                                </div>
                                <h4 className="font-semibold text-sm text-slate-200 truncate pr-4">{incident.title}</h4>
                                <div className="flex items-center justify-between mt-3">
                                    <div className="flex items-center gap-2">
                                        <Badge variant="outline" className="text-[10px] opacity-70">
                                            {incident.status}
                                        </Badge>
                                        {incident.assignee && (
                                            <div className="flex items-center gap-1 text-[10px] text-slate-500">
                                                <User className="h-3 w-3" />
                                                <span>{incident.assignee}</span>
                                            </div>
                                        )}
                                    </div>
                                    <span className="text-[10px] text-slate-500">{new Date(incident.createdAt).toLocaleDateString()}</span>
                                </div>
                                <ChevronRight className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-600 opacity-0 group-hover:opacity-100 transition-opacity" />
                            </button>
                        ))
                    )}
                </div>

                {/* Detail View */}
                <div className={cn(
                    "flex-1 flex flex-col bg-slate-950/50",
                    !selectedIncident && "hidden lg:flex items-center justify-center text-slate-600"
                )}>
                    {selectedIncident ? (
                        <>
                            {/* Detail Header */}
                            <div className="p-6 border-b border-slate-800 bg-slate-900/50">
                                <div className="flex justify-between items-start mb-4">
                                    <Button variant="ghost" size="sm" className="lg:hidden -ml-2 mb-2 p-1" onClick={() => setSelectedIncident(null)}>
                                        <ChevronRight className="h-5 w-5 rotate-180" /> Back to Queue
                                    </Button>
                                    <div className="flex gap-2">
                                        <Button variant="outline" size="sm" onClick={() => handleUpdateStatus('Resolved')}>
                                            <CheckCircle className="h-4 w-4 mr-2 text-emerald-500" /> Mark Resolved
                                        </Button>
                                        {!selectedIncident.assignee && (
                                            <Button variant="primary" size="sm" onClick={handleAssign}>
                                                <User className="h-4 w-4 mr-2" /> Assign to Me
                                            </Button>
                                        )}
                                        <div className="relative group">
                                            <Button variant="ghost" size="sm" className="h-9 w-9 p-0">
                                                <MoreVertical className="h-4 w-4" />
                                            </Button>
                                            {/* Status Dropdown Simulation */}
                                            <div className="absolute right-0 top-full mt-1 w-40 bg-slate-800 border border-slate-700 rounded-lg shadow-xl hidden group-hover:block z-50">
                                                <div className="p-1">
                                                    {(['New', 'Assigned', 'In Progress', 'Resolved', 'Closed'] as IncidentStatus[]).map(s => (
                                                        <button
                                                            key={s}
                                                            onClick={() => handleUpdateStatus(s)}
                                                            className="w-full text-left px-3 py-2 text-xs hover:bg-slate-700 rounded transition-colors"
                                                        >
                                                            Move to {s}
                                                        </button>
                                                    ))}
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div className="flex items-center gap-3 mb-2">
                                    <span className="text-sm font-mono text-brand-400">{selectedIncident.id}</span>
                                    <Badge variant={
                                        selectedIncident.severity === 'critical' ? 'error' :
                                            selectedIncident.severity === 'high' ? 'warning' : 'default'
                                    }>
                                        {selectedIncident.severity.toUpperCase()}
                                    </Badge>
                                    <Badge variant="outline">{selectedIncident.status}</Badge>
                                </div>
                                <h2 className="text-2xl font-bold text-white">{selectedIncident.title}</h2>
                                <p className="text-slate-400 mt-2">{selectedIncident.description}</p>
                            </div>

                            {/* Detail Content */}
                            <div className="flex-1 overflow-y-auto p-6 space-y-8">
                                {/* Metadata Grid */}
                                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                                    <IncidentMetadataField label="Assignee">
                                        <div className="h-6 w-6 rounded-full bg-brand-500 flex items-center justify-center text-[10px]">EA</div>
                                        <span>{selectedIncident.assignee || 'Unassigned'}</span>
                                    </IncidentMetadataField>
                                    <IncidentMetadataField label="First Detected">
                                        <Clock className="h-4 w-4 text-slate-500" />
                                        <span>{new Date(selectedIncident.createdAt).toLocaleString()}</span>
                                    </IncidentMetadataField>
                                    <IncidentMetadataField label="Risk Score">
                                        <AlertTriangle className={cn(
                                            "h-4 w-4",
                                            selectedIncident.riskScore > 80 ? "text-red-500" : "text-amber-500"
                                        )} />
                                        <span className="text-lg font-bold">{selectedIncident.riskScore}</span>
                                    </IncidentMetadataField>
                                </div>

                                {/* Identity Section */}
                                {resolvedEntities.length > 0 && (
                                    <div className="space-y-3 p-4 bg-slate-900/50 rounded-xl border border-slate-800">
                                        <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
                                            <Shield className="h-3.5 w-3.5" /> Resolved Identities
                                        </h3>
                                        <div className="flex flex-wrap gap-4">
                                            {resolvedEntities.map(entity => entity && (
                                                <div key={entity.id} className="flex items-center gap-3 bg-slate-800 p-2 rounded-lg border border-slate-700 hover:border-brand-500 transition-colors cursor-pointer group">
                                                    <div className={cn(
                                                        "p-1.5 rounded-lg",
                                                        entity.type === 'user' ? "bg-brand-500/20 text-brand-400" : "bg-indigo-500/20 text-indigo-400"
                                                    )}>
                                                        {entity.type === 'user' ? <User className="h-4 w-4" /> : <Server className="h-4 w-4" />}
                                                    </div>
                                                    <div>
                                                        <div className="text-sm font-bold text-white group-hover:text-brand-400 transition-colors">
                                                            <Link to={`/entity/${entity.id}`}>{entity.name}</Link>
                                                        </div>
                                                        <div className="text-[10px] text-slate-500 font-mono">{entity.id}</div>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}

                                {/* Timeline / Journal */}
                                <div className="space-y-4">
                                    <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                                        <History className="h-5 w-5 text-slate-400" />
                                        Case Journal & Timeline
                                    </h3>

                                    <div className="relative space-y-6 before:absolute before:inset-0 before:ml-5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-slate-800 before:via-slate-800 before:to-transparent">
                                        {selectedIncident.journal.map((entry, idx) => (
                                            <div key={entry.id} className="relative flex items-start gap-4 animate-in slide-in-from-top-4 duration-500" style={{ animationDelay: `${idx * 50}ms` }}>
                                                <div className={cn(
                                                    "absolute left-0 mt-1 h-10 w-10 rounded-full border-4 border-slate-900 flex items-center justify-center translate-x-[-15%] z-10",
                                                    entry.type === 'system' ? "bg-slate-800" : "bg-brand-600"
                                                )}>
                                                    {entry.type === 'system' ? <Settings className="h-4 w-4 text-slate-400" /> : <MessageSquare className="h-4 w-4 text-white" />}
                                                </div>
                                                <div className="ml-12 flex-1 pt-1">
                                                    <div className="flex items-center justify-between mb-1">
                                                        <span className="text-xs font-bold text-slate-400">{entry.author}</span>
                                                        <span className="text-[10px] text-slate-500">{new Date(entry.timestamp).toLocaleTimeString()}</span>
                                                    </div>
                                                    <div className={cn(
                                                        "p-3 rounded-lg text-sm",
                                                        entry.type === 'system' ? "bg-slate-900/50 text-slate-500 italic" : "bg-slate-800 text-slate-200"
                                                    )}>
                                                        {entry.message}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>

                            {/* Footer / Input */}
                            <form onSubmit={handleAddNote} className="p-4 border-t border-slate-800 bg-slate-900/80">
                                <div className="flex gap-2">
                                    <Input
                                        placeholder="Add analysis note or attachment..."
                                        className="bg-slate-950 border-slate-800 h-10"
                                        value={newNote}
                                        onChange={(e) => setNewNote(e.target.value)}
                                    />
                                    <Button type="submit" variant="primary" disabled={!newNote.trim()}>
                                        Post Note
                                    </Button>
                                </div>
                            </form>
                        </>
                    ) : (
                        <div className="flex flex-col items-center justify-center space-y-4">
                            <ShieldAlert className="h-16 w-16 text-slate-800" />
                            <div className="text-center">
                                <h3 className="text-xl font-bold text-slate-400">No Incident Selected</h3>
                                <p className="text-slate-600">Select an incident from the queue to start investigation.</p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

// Helper components that were missing in icons but used
const Settings = ({ className }: { className?: string }) => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" /></svg>
);
