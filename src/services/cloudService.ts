// Cloud Service - Multi-cloud asset and event management
// NOTE: This is a prototype implementation with mock data
// TODO: Connect to backend API in /api/cloud/* endpoints

export interface CloudAsset {
    id: string;
    provider: 'aws' | 'azure' | 'gcp';
    type: string;
    name: string;
    region: string;
    account: string;
    tags: Record<string, string>;
    state: string;
    created_at: string;
    last_seen: string;
    private_ip?: string;
    public_ip?: string;
    vpc?: string;
    security_groups?: string[];
    iam_role?: string;
    permissions?: string[];
    risk_score: number;
    risk_factors?: string[];
    raw_metadata: Record<string, any>;
}

export interface CloudEvent {
    id: string;
    provider: 'aws' | 'azure' | 'gcp';
    event_type: string;
    event_name: string;
    timestamp: string;
    actor_id: string;
    actor_type: string;
    actor_name: string;
    resource_id: string;
    resource_type: string;
    source_ip?: string;
    user_agent?: string;
    region: string;
    account: string;
    success: boolean;
    error_code?: string;
    error_message?: string;
    raw_event: Record<string, any>;
}

export interface AssetFilter {
    provider?: 'aws' | 'azure' | 'gcp' | 'all';
    type?: string;
    region?: string;
    account?: string;
    min_risk?: number;
    max_risk?: number;
}

export interface AssetGraphResponse {
    center: CloudAsset;
    connected: CloudAsset[];
    depth: number;
}

export interface BlastRadiusResponse {
    asset_id: string;
    blast_radius: number;
    severity: 'low' | 'medium' | 'high' | 'critical';
}

export interface CloudStats {
    total_assets: number;
    total_relationships: number;
    by_provider: Record<string, number>;
    by_type: Record<string, number>;
}

class CloudService {
    private mockAssets: CloudAsset[] = [];

    constructor() {
        this.loadMockData();
    }

    private loadMockData() {
        // Mock AWS EC2 instances
        this.mockAssets = [
            {
                id: 'aws:ec2:i-1234567890abcdef0',
                provider: 'aws',
                type: 'ec2',
                name: 'prod-web-server-01',
                region: 'us-east-1',
                account: '123456789012',
                tags: { Environment: 'Production', Team: 'Platform' },
                state: 'running',
                created_at: '2024-01-15T10:00:00Z',
                last_seen: new Date().toISOString(),
                private_ip: '10.0.1.50',
                public_ip: '54.123.45.67',
                vpc: 'vpc-abc123',
                security_groups: ['sg-web', 'sg-ssh'],
                iam_role: 'arn:aws:iam::123456789012:role/WebServerRole',
                risk_score: 45,
                risk_factors: ['Public IP', 'SSH Access'],
                raw_metadata: {}
            },
            {
                id: 'aws:ec2:i-0987654321fedcba0',
                provider: 'aws',
                type: 'ec2',
                name: 'prod-db-server-01',
                region: 'us-east-1',
                account: '123456789012',
                tags: { Environment: 'Production', Team: 'Data' },
                state: 'running',
                created_at: '2024-01-10T08:00:00Z',
                last_seen: new Date().toISOString(),
                private_ip: '10.0.2.100',
                vpc: 'vpc-abc123',
                security_groups: ['sg-db'],
                iam_role: 'arn:aws:iam::123456789012:role/DatabaseRole',
                risk_score: 75,
                risk_factors: ['Database', 'High Privilege IAM Role'],
                raw_metadata: {}
            },
            {
                id: 'azure:vm:vm-prod-app-01',
                provider: 'azure',
                type: 'vm',
                name: 'prod-app-01',
                region: 'eastus',
                account: 'subscription-abc-123',
                tags: { Environment: 'Production', Application: 'API' },
                state: 'running',
                created_at: '2024-02-01T12:00:00Z',
                last_seen: new Date().toISOString(),
                private_ip: '10.1.1.50',
                public_ip: '20.123.45.67',
                risk_score: 30,
                risk_factors: ['Public Endpoint'],
                raw_metadata: {}
            },
            {
                id: 'gcp:compute:instance-prod-analytics-01',
                provider: 'gcp',
                type: 'compute-instance',
                name: 'prod-analytics-01',
                region: 'us-central1-a',
                account: 'project-analytics-prod',
                tags: { environment: 'production', workload: 'analytics' },
                state: 'running',
                created_at: '2024-01-20T14:00:00Z',
                last_seen: new Date().toISOString(),
                private_ip: '10.2.1.100',
                risk_score: 20,
                risk_factors: [],
                raw_metadata: {}
            }
        ];
    }

    async getAssets(filter: AssetFilter = {}): Promise<CloudAsset[]> {
        // Simulate API delay
        await new Promise(resolve => setTimeout(resolve, 300));

        let filtered = [...this.mockAssets];

        if (filter.provider && filter.provider !== 'all') {
            filtered = filtered.filter(a => a.provider === filter.provider);
        }
        if (filter.type) {
            filtered = filtered.filter(a => a.type === filter.type);
        }
        if (filter.region) {
            filtered = filtered.filter(a => a.region === filter.region);
        }
        if (filter.account) {
            filtered = filtered.filter(a => a.account === filter.account);
        }
        if (filter.min_risk !== undefined) {
            filtered = filtered.filter(a => a.risk_score >= filter.min_risk!);
        }
        if (filter.max_risk !== undefined) {
            filtered = filtered.filter(a => a.risk_score <= filter.max_risk!);
        }

        return filtered;
    }

    async getAsset(id: string): Promise<CloudAsset> {
        await new Promise(resolve => setTimeout(resolve, 100));

        const asset = this.mockAssets.find(a => a.id === id);
        if (!asset) {
            throw new Error('Asset not found');
        }
        return asset;
    }

    async getAssetGraph(id: string, depth: number = 2): Promise<AssetGraphResponse> {
        await new Promise(resolve => setTimeout(resolve, 200));

        const center = await this.getAsset(id);

        // Mock: return assets in same VPC or region
        const connected = this.mockAssets.filter(a =>
            a.id !== id && (
                (a.vpc && a.vpc === center.vpc) ||
                (a.region === center.region && a.provider === center.provider)
            )
        );

        return {
            center,
            connected: connected.slice(0, 5), // Limit for demo
            depth
        };
    }

    async getBlastRadius(id: string): Promise<BlastRadiusResponse> {
        await new Promise(resolve => setTimeout(resolve, 150));

        const graph = await this.getAssetGraph(id, 3);
        const radius = graph.connected.length;

        let severity: 'low' | 'medium' | 'high' | 'critical';
        if (radius > 10) severity = 'critical';
        else if (radius > 5) severity = 'high';
        else if (radius > 2) severity = 'medium';
        else severity = 'low';

        return {
            asset_id: id,
            blast_radius: radius,
            severity
        };
    }

    async getStats(): Promise<CloudStats> {
        await new Promise(resolve => setTimeout(resolve, 100));

        const byProvider: Record<string, number> = {};
        const byType: Record<string, number> = {};

        this.mockAssets.forEach(asset => {
            byProvider[asset.provider] = (byProvider[asset.provider] || 0) + 1;
            byType[asset.type] = (byType[asset.type] || 0) + 1;
        });

        return {
            total_assets: this.mockAssets.length,
            total_relationships: 8, // Mock
            by_provider: byProvider,
            by_type: byType
        };
    }

    async ingestAsset(provider: 'aws' | 'azure' | 'gcp', rawData: Record<string, any>): Promise<CloudAsset> {
        await new Promise(resolve => setTimeout(resolve, 200));

        // Mock: create a simple asset
        const asset: CloudAsset = {
            id: `${provider}:${rawData.type || 'unknown'}:${Date.now()}`,
            provider,
            type: rawData.type || 'unknown',
            name: rawData.name || 'Unnamed Asset',
            region: rawData.region || 'unknown',
            account: rawData.account || 'unknown',
            tags: rawData.tags || {},
            state: rawData.state || 'unknown',
            created_at: new Date().toISOString(),
            last_seen: new Date().toISOString(),
            risk_score: 0,
            raw_metadata: rawData
        };

        this.mockAssets.push(asset);
        return asset;
    }
}

export const cloudService = new CloudService();
