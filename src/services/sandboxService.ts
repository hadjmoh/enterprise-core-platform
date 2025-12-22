import type { ResourceLimits, ResourceUsage } from './sandboxTypes';

const API_BASE = '/api/v1';

export interface SandboxStatus {
    app_name: string;
    active: boolean;
    created_at: string;
    limits: ResourceLimits;
    capabilities: string[];
}

export const sandboxService = {
    /**
     * Get sandbox status for an app
     */
    async getSandboxStatus(appName: string): Promise<SandboxStatus> {
        const response = await fetch(`${API_BASE}/apps/${appName}/sandbox`);
        if (!response.ok) {
            throw new Error('Failed to fetch sandbox status');
        }
        return response.json();
    },

    /**
     * Get current resource usage for an app
     */
    async getResourceUsage(appName: string): Promise<ResourceUsage> {
        const response = await fetch(`${API_BASE}/apps/${appName}/resources`);
        if (!response.ok) {
            throw new Error('Failed to fetch resource usage');
        }
        return response.json();
    },

    /**
     * Update resource limits for an app
     */
    async updateLimits(appName: string, limits: ResourceLimits): Promise<void> {
        const response = await fetch(`${API_BASE}/apps/${appName}/limits`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(limits),
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || 'Failed to update limits');
        }
    },
};
