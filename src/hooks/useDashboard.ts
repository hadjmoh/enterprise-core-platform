import { useState, useCallback } from 'react';
import type { Layout, LayoutItem } from 'react-grid-layout';

export interface Widget {
    id: string;
    type: string;
    spl: string;
    title: string;
    layout: LayoutItem;
}

export interface DashboardConfig {
    id: string;
    name: string;
    widgets: Widget[];
    isLocked: boolean;
}

const STORAGE_KEY = 'enterprise_core_dashboard_config';

const DEFAULT_CONFIG: DashboardConfig = {
    id: 'default',
    name: 'Main Dashboard',
    isLocked: false,
    widgets: [
        {
            id: 'w1',
            type: 'timechart',
            spl: 'search index=main EARLIEST=rt | timechart span=10s count by status',
            title: 'System Throughput',
            layout: { i: 'w1', x: 0, y: 0, w: 12, h: 4 }
        },
        {
            id: 'w2',
            type: 'bar',
            spl: 'search index=main | stats count by host',
            title: 'Events by Host',
            layout: { i: 'w2', x: 0, y: 4, w: 6, h: 4 }
        },
        {
            id: 'w3',
            type: 'pie',
            spl: 'search index=main | stats count by severity',
            title: 'Severity Distribution',
            layout: { i: 'w3', x: 6, y: 4, w: 6, h: 4 }
        }
    ]
};

export function useDashboard() {
    const [config, setConfig] = useState<DashboardConfig>(() => {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) {
            try {
                return JSON.parse(saved);
            } catch (e) {
                console.error('Failed to parse dashboard config', e);
            }
        }
        return DEFAULT_CONFIG;
    });

    const saveConfig = useCallback((newConfig: DashboardConfig) => {
        setConfig(newConfig);
        localStorage.setItem(STORAGE_KEY, JSON.stringify(newConfig));
    }, []);

    const updateLayouts = useCallback((layouts: Layout) => {
        setConfig(prev => {
            const nextWidgets = prev.widgets.map(w => {
                const newLayout = layouts.find(l => l.i === w.id);
                return newLayout ? { ...w, layout: newLayout } : w;
            });
            const nextConfig = { ...prev, widgets: nextWidgets };
            localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig));
            return nextConfig;
        });
    }, []);

    const addWidget = useCallback((widget: Omit<Widget, 'id' | 'layout'>) => {
        const id = `w-${Date.now()}`;
        const newWidget: Widget = {
            ...widget,
            id,
            layout: { i: id, x: 0, y: Infinity, w: 6, h: 4 }
        };
        saveConfig({ ...config, widgets: [...config.widgets, newWidget] });
    }, [config, saveConfig]);

    const removeWidget = useCallback((id: string) => {
        saveConfig({ ...config, widgets: config.widgets.filter(w => w.id !== id) });
    }, [config, saveConfig]);

    const resetLayout = useCallback(() => {
        saveConfig(DEFAULT_CONFIG);
    }, [saveConfig]);

    return {
        config,
        updateLayouts,
        addWidget,
        removeWidget,
        resetLayout
    };
}
