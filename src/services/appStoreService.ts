import type { AppMetadata } from '@/types/apps';

const API_BASE = '/api/v1';

export interface AppWithStatus extends AppMetadata {
    installed: boolean;
}

export const appStoreService = {
    /**
     * Fetch all available apps from the registry
     */
    async fetchAvailableApps(): Promise<AppWithStatus[]> {
        const response = await fetch(`${API_BASE}/apps/available`);
        if (!response.ok) {
            throw new Error('Failed to fetch available apps');
        }
        return response.json();
    },

    /**
     * Search for apps matching a query
     */
    async searchApps(query: string): Promise<AppWithStatus[]> {
        const response = await fetch(`${API_BASE}/apps/search?q=${encodeURIComponent(query)}`);
        if (!response.ok) {
            throw new Error('Failed to search apps');
        }
        return response.json();
    },

    /**
     * Get detailed metadata for a specific app
     */
    async getAppDetails(name: string): Promise<AppWithStatus> {
        const response = await fetch(`${API_BASE}/apps/available/${name}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch app details for ${name}`);
        }
        return response.json();
    },

    /**
     * Install an app from a .tar.gz file
     */
    async installApp(file: File, onProgress?: (progress: number) => void): Promise<void> {
        const formData = new FormData();
        formData.append('bundle', file);

        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();

            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable && onProgress) {
                    const progress = (e.loaded / e.total) * 100;
                    onProgress(progress);
                }
            });

            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    resolve();
                } else {
                    reject(new Error(xhr.responseText || 'Installation failed'));
                }
            });

            xhr.addEventListener('error', () => {
                reject(new Error('Network error during installation'));
            });

            xhr.open('POST', `${API_BASE}/apps/install`);
            xhr.send(formData);
        });
    },

    /**
     * Uninstall an app
     */
    async uninstallApp(name: string): Promise<void> {
        const response = await fetch(`${API_BASE}/apps/${name}`, {
            method: 'DELETE',
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || 'Failed to uninstall app');
        }
    },
};
