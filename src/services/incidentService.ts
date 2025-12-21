import { v4 as uuidv4 } from 'uuid';
import type { DetectionSignal } from './correlationService';

export type IncidentStatus = 'New' | 'Assigned' | 'In Progress' | 'Resolved' | 'Closed';
export type IncidentSeverity = 'low' | 'medium' | 'high' | 'critical';

export interface IncidentJournalEntry {
    id: string;
    timestamp: string;
    author: string;
    message: string;
    type: 'note' | 'system';
}

export interface Incident {
    id: string;
    title: string;
    description: string;
    severity: IncidentSeverity;
    status: IncidentStatus;
    assignee?: string;
    sourceSignals: string[]; // IDs of detection signals
    journal: IncidentJournalEntry[];
    createdAt: string;
    updatedAt: string;
    riskScore: number;
}

class IncidentService {
    private incidents: Incident[] = [];
    private listeners: ((incidents: Incident[]) => void)[] = [];

    constructor() {
        this.loadMockIncidents();
    }

    getIncidents(): Incident[] {
        return this.incidents;
    }

    getIncidentById(id: string): Incident | undefined {
        return this.incidents.find(i => i.id === id);
    }

    createIncidentFromSignal(signal: DetectionSignal): Incident {
        const incident: Incident = {
            id: `INC-${uuidv4()}`,
            title: `Detection: ${signal.ruleName}`,
            description: signal.message,
            severity: signal.severity as IncidentSeverity,
            status: 'New',
            sourceSignals: [signal.id],
            journal: [
                {
                    id: uuidv4(),
                    timestamp: new Date().toISOString(),
                    author: 'System',
                    message: `Incident created automatically from detection signal ${signal.id}`,
                    type: 'system'
                }
            ],
            createdAt: signal.timestamp,
            updatedAt: new Date().toISOString(),
            riskScore: signal.riskScore
        };

        this.incidents.unshift(incident);
        this.notify();
        return incident;
    }

    updateIncidentStatus(id: string, status: IncidentStatus, author: string = 'Current User') {
        const incident = this.getIncidentById(id);
        if (incident) {
            const oldStatus = incident.status;
            incident.status = status;
            incident.updatedAt = new Date().toISOString();
            incident.journal.push({
                id: uuidv4(),
                timestamp: incident.updatedAt,
                author,
                message: `Status changed from ${oldStatus} to ${status}`,
                type: 'system'
            });
            this.notify();
        }
    }

    assignIncident(id: string, assignee: string, author: string = 'Current User') {
        const incident = this.getIncidentById(id);
        if (incident) {
            incident.assignee = assignee;
            incident.updatedAt = new Date().toISOString();
            incident.journal.push({
                id: uuidv4(),
                timestamp: incident.updatedAt,
                author,
                message: `Incident assigned to ${assignee}`,
                type: 'system'
            });
            this.notify();
        }
    }

    addNote(id: string, message: string, author: string = 'Current User') {
        const incident = this.getIncidentById(id);
        if (incident) {
            incident.updatedAt = new Date().toISOString();
            incident.journal.push({
                id: uuidv4(),
                timestamp: incident.updatedAt,
                author,
                message,
                type: 'note'
            });
            this.notify();
        }
    }

    private notify() {
        this.listeners.forEach(l => l([...this.incidents]));
    }

    subscribe(listener: (incidents: Incident[]) => void) {
        this.listeners.push(listener);
        return () => {
            this.listeners = this.listeners.filter(l => l !== listener);
        };
    }

    private loadMockIncidents() {
        // Initial state can be empty or have some history
        this.incidents = [
            {
                id: 'INC-1024',
                title: 'Suspicious Administrative Access',
                description: 'Detected use of sudo by unauthorized service account',
                severity: 'high',
                status: 'In Progress',
                assignee: 'Alice Analyst',
                sourceSignals: ['sig-001'],
                journal: [
                    { id: 'j-1', timestamp: new Date(Date.now() - 3600000).toISOString(), author: 'System', message: 'Incident created from signal sig-001', type: 'system' },
                    { id: 'j-2', timestamp: new Date(Date.now() - 3000000).toISOString(), author: 'Bob Admin', message: 'Investigating source IP logs.', type: 'note' }
                ],
                createdAt: new Date(Date.now() - 3600000).toISOString(),
                updatedAt: new Date(Date.now() - 3000000).toISOString(),
                riskScore: 85
            }
        ];
    }
}

export const incidentService = new IncidentService();
