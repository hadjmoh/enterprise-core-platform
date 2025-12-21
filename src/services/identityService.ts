

export type EntityType = 'user' | 'host' | 'ip';

export interface Entity {
    id: string;
    name: string;
    type: EntityType;
    labels: string[];
    riskScore: number;
    metadata: Record<string, any>;
    activities: {
        timestamp: string;
        action: string;
        source: string;
    }[];
}

export interface IdentityMapping {
    identifier: string; // e.g., "192.168.1.50" or "jdoe"
    identifierType: 'ip' | 'account';
    entityId: string;
    validFrom: string;
    validTo?: string;
}

class IdentityService {
    private entities: Entity[] = [];
    private mappings: IdentityMapping[] = [];

    constructor() {
        this.loadMockData();
    }

    getEntities(): Entity[] {
        return this.entities;
    }

    getEntityById(id: string): Entity | undefined {
        return this.entities.find(e => e.id === id);
    }

    resolveIdentity(identifier: string, timestamp: string = new Date().toISOString()): Entity | undefined {
        const mapping = this.mappings.find(m => {
            if (m.identifier !== identifier) return false;
            const ts = new Date(timestamp).getTime();
            const from = new Date(m.validFrom).getTime();
            const to = m.validTo ? new Date(m.validTo).getTime() : Infinity;
            return ts >= from && ts <= to;
        });

        if (mapping) {
            return this.getEntityById(mapping.entityId);
        }
        return undefined;
    }

    searchEntities(query: string): Entity[] {
        const q = query.toLowerCase();
        return this.entities.filter(e =>
            e.name.toLowerCase().includes(q) ||
            e.id.toLowerCase().includes(q) ||
            e.labels.some(l => l.toLowerCase().includes(q))
        );
    }

    updateEntityRisk(id: string, delta: number) {
        const entity = this.getEntityById(id);
        if (entity) {
            entity.riskScore = Math.min(100, Math.max(0, entity.riskScore + delta));
        }
    }

    private loadMockData() {
        this.entities = [
            {
                id: 'ENT-USR-01',
                name: 'Alice Analyst',
                type: 'user',
                labels: ['SOC-L2', 'Paris-Office', 'VIP'],
                riskScore: 45,
                metadata: {
                    department: 'Security Operations',
                    title: 'Senior Security Analyst',
                    email: 'alice@enterprise-core.io'
                },
                activities: [
                    { timestamp: new Date(Date.now() - 3600000).toISOString(), action: 'login', source: 'VPN-Paris' },
                    { timestamp: new Date(Date.now() - 7200000).toISOString(), action: 'access_file', source: 'FS-HR-01' }
                ]
            },
            {
                id: 'ENT-HST-42',
                name: 'prod-db-cluster-0a',
                type: 'host',
                labels: ['Production', 'Database', 'Critical-Infra'],
                riskScore: 12,
                metadata: {
                    os: 'Ubuntu 22.04 LTS',
                    kernel: '5.15.0-generic',
                    cpu: '64 Cores',
                    region: 'eu-central-1'
                },
                activities: [
                    { timestamp: new Date(Date.now() - 100000).toISOString(), action: 'reboot', source: 'System' }
                ]
            }
        ];

        this.mappings = [
            {
                identifier: '10.0.0.50',
                identifierType: 'ip',
                entityId: 'ENT-HST-42',
                validFrom: '2023-01-01T00:00:00Z'
            },
            {
                identifier: 'aanalyst',
                identifierType: 'account',
                entityId: 'ENT-USR-01',
                validFrom: '2023-01-01T00:00:00Z'
            }
        ];
    }
}

export const identityService = new IdentityService();
