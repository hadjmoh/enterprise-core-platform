import { v4 as uuidv4 } from 'uuid';
import { correlationService } from './correlationService';
import { secureRandom, secureRandomInt } from '../utils/crypto';

export interface LogEvent {
    id: string;
    timestamp: string;
    source: string;
    message: string;
    [key: string]: any;
}

class IngestionService {
    private isRunning: boolean = false;
    private intervalId: any;

    startMockIngestion() {
        if (this.isRunning) return;
        this.isRunning = true;
        console.log('Starting Mock Ingestion...');

        this.intervalId = setInterval(() => {
            const event = this.generateMockEvent();
            this.processEvent(event);
        }, 2000); // 1 event every 2 seconds
    }

    stopMockIngestion() {
        this.isRunning = false;
        clearInterval(this.intervalId);
        console.log('Stopped Mock Ingestion.');
    }

    processEvent(event: LogEvent) {
        // 1. Store event (Mock: skip storage)

        // 2. Correlation Hook
        correlationService.processEvent(event);
    }

    private generateMockEvent(): LogEvent {
        const sources = ['firewall-01', 'auth-server', 'web-gateway', 'endpoint-pc-04'];
        const actions = ['login_success', 'login_failed', 'file_access', 'connection_allow', 'connection_deny'];

        const source = sources[secureRandomInt(0, sources.length - 1)];
        const action = actions[secureRandomInt(0, actions.length - 1)];

        // Occasional "Attack" patterns
        let message = `Action ${action} on ${source}`;
        let extra: any = {};

        if (secureRandom() < 0.1) {
            // Simulate Brute Force (burst handled by interval? No, just random events for now)
            // To properly text brute force, we might need a burst generator.
            // But for random background noise:
            if (secureRandom() < 0.3) {
                message = 'sudo: session opened for root';
            }
        }

        const ip = '10.0.0.' + secureRandomInt(0, 255);
        const user = secureRandom() > 0.5 ? 'aanalyst' : 'unknown';

        return {
            id: uuidv4(),
            timestamp: new Date().toISOString(),
            source,
            message,
            action,
            src_ip: ip,
            user: user,
            ...extra
        };
    }

    // Helper to force specific events for testing
    injectEvent(event: Partial<LogEvent>) {
        const fullEvent: LogEvent = {
            id: uuidv4(),
            timestamp: new Date().toISOString(),
            source: 'manual-inject',
            message: 'Manual injection',
            ...event
        };
        this.processEvent(fullEvent);
    }
}

export const ingestionService = new IngestionService();
