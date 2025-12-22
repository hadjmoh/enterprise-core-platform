export interface AppManifest {
    apiVersion: string;
    kind: string;
    metadata: {
        name: string;
        displayName: string;
        version: string;
        description: string;
        icon?: string;
    };
}

export interface DashboardLayout {
    id: string;
    title: string;
    description: string;
    rows: DashboardRow[];
}

export interface DashboardRow {
    height: string;
    columns: DashboardColumn[];
}

export interface DashboardColumn {
    width: number; // 1-12
    panels: DashboardPanel[];
}

export interface DashboardPanel {
    id: string;
    title: string;
    type: 'line' | 'bar' | 'area' | 'stat' | 'table';
    query: string;
    options?: Record<string, any>;
}

// App Store Types
export interface AppMetadata {
    name: string;
    version: string;
    title: string;
    author: string;
    category: string;
    description: string;
    longDescription?: string;
    icon?: string;
    screenshots?: string[];
    permissions?: string[];
    tags?: string[];
    downloads?: number;
    rating?: number;
    downloadUrl?: string;
}
