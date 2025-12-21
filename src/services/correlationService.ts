import { v4 as uuidv4 } from 'uuid';

export interface CorrelationRule {
    id: string;
    name: string;
    description: string;
    severity: 'low' | 'medium' | 'high' | 'critical';
    riskScore: number;
    enabled: boolean;
    logic: {
        field: string;
        operator: 'equals' | 'contains' | 'gt' | 'lt';
        value: string | number;
    };
    threshold?: {
        count: number;
        windowSeconds: number;
    };
}

export interface DetectionSignal {
    id: string;
    ruleId: string;
    ruleName: string;
    severity: 'low' | 'medium' | 'high' | 'critical';
    timestamp: string;
    sourceIds: string[];
    riskScore: number;
    message: string;
}

interface LogEvent {
    id: string;
    timestamp: string;
    source: string;
    [key: string]: any;
}

class CorrelationEngine {
    private rules: CorrelationRule[] = [];
    private signalHistory: DetectionSignal[] = [];
    private eventBuffer: Map<string, LogEvent[]> = new Map(); // RuleID -> Event[] for windowing
    private listeners: ((signal: DetectionSignal) => void)[] = [];

    constructor() {
        this.loadRules();
    }

    // --- Configuration ---

    getRules() {
        return this.rules;
    }

    addRule(rule: Omit<CorrelationRule, 'id'>) {
        const newRule: CorrelationRule = { ...rule, id: uuidv4(), enabled: true };
        this.rules.push(newRule);
        return newRule;
    }

    toggleRule(id: string) {
        const rule = this.rules.find(r => r.id === id);
        if (rule) rule.enabled = !rule.enabled;
    }

    deleteRule(id: string) {
        this.rules = this.rules.filter(r => r.id !== id);
    }

    // --- Core Engine ---

    processEvent(event: LogEvent) {
        this.rules.filter(r => r.enabled).forEach(rule => {
            if (this.matchesLogic(event, rule)) {
                if (rule.threshold) {
                    this.handleThresholdRule(event, rule);
                } else {
                    this.triggerDetection(event, rule, [event.id]);
                }
            }
        });
    }

    private matchesLogic(event: any, rule: CorrelationRule): boolean {
        const val = event[rule.logic.field];
        if (val === undefined) return false;

        switch (rule.logic.operator) {
            case 'equals': return val == rule.logic.value;
            case 'contains': return String(val).includes(String(rule.logic.value));
            case 'gt': return Number(val) > Number(rule.logic.value);
            case 'lt': return Number(val) < Number(rule.logic.value);
            default: return false;
        }
    }

    private handleThresholdRule(event: LogEvent, rule: CorrelationRule) {
        if (!rule.threshold) return;

        // Initialize buffer
        if (!this.eventBuffer.has(rule.id)) {
            this.eventBuffer.set(rule.id, []);
        }

        const buffer = this.eventBuffer.get(rule.id)!;
        buffer.push(event);

        // Prune old events
        const now = new Date(event.timestamp).getTime();
        const cutoff = now - (rule.threshold.windowSeconds * 1000);

        // Optimize: remove events strictly older than window
        const validEvents = buffer.filter(e => new Date(e.timestamp).getTime() > cutoff);
        this.eventBuffer.set(rule.id, validEvents);

        // Check Trigger
        if (validEvents.length >= rule.threshold.count) {
            this.triggerDetection(event, rule, validEvents.map(e => e.id));
            // Optional: Clear buffer to prevent identifying the same cluster multiple times? 
            // For now, let's just clear specific IDs or implement a cooldown. 
            // Simple approach: Clear buffer to reset count.
            this.eventBuffer.set(rule.id, []);
        }
    }

    private triggerDetection(triggerEvent: LogEvent, rule: CorrelationRule, sourceIds: string[]) {
        const signal: DetectionSignal = {
            id: uuidv4(),
            ruleId: rule.id,
            ruleName: rule.name,
            severity: rule.severity,
            timestamp: new Date().toISOString(),
            sourceIds: sourceIds,
            riskScore: rule.riskScore,
            message: `Rule '${rule.name}' triggered by ${triggerEvent.source}`
        };

        this.signalHistory.unshift(signal);
        if (this.signalHistory.length > 1000) this.signalHistory.pop();

        this.notifyListeners(signal);
    }

    // --- Signals & Pub/Sub ---

    getSignals() {
        return this.signalHistory;
    }

    onSignal(callback: (signal: DetectionSignal) => void) {
        this.listeners.push(callback);
    }

    private notifyListeners(signal: DetectionSignal) {
        this.listeners.forEach(cb => cb(signal));
    }

    // --- Persistence (Mock) ---
    private loadRules() {
        this.rules = [
            {
                id: 'rule-brute-force',
                name: 'Potential Brute Force',
                description: 'Detects 5+ failed logins in 1 minute',
                severity: 'high',
                riskScore: 80,
                enabled: true,
                logic: { field: 'action', operator: 'equals', value: 'login_failed' },
                threshold: { count: 5, windowSeconds: 60 }
            },
            {
                id: 'rule-root-access',
                name: 'Root Access Granted',
                description: 'Detects any successful sudo escalation',
                severity: 'critical',
                riskScore: 100,
                enabled: true,
                logic: { field: 'message', operator: 'contains', value: 'sudo: session opened' }
            },
            {
                id: 'rule-lateral',
                name: 'Suspicious Lateral Movement',
                description: 'Detects use of PsExec or admin shares',
                severity: 'medium',
                riskScore: 50,
                enabled: true,
                logic: { field: 'process', operator: 'contains', value: 'psexec' }
            }
        ];
    }
}

export const correlationService = new CorrelationEngine();
