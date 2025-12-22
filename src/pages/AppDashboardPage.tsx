import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import type { DashboardLayout } from '@/types/apps';
import { DashboardRenderer } from '@/components/apps/DashboardRenderer';
import { AlertCircle } from 'lucide-react';

const AppDashboardPage = () => {
    const { appName, dashboardId } = useParams<{ appName: string; dashboardId: string }>();
    const [layout, setLayout] = useState<DashboardLayout | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchDashboard = async () => {
            if (!appName || !dashboardId) return;

            try {
                // In production, fetch from API:
                // const res = await fetch(`/api/v1/apps/${appName}/dashboards/${dashboardId}`);
                // if (!res.ok) throw new Error(await res.text());
                // const data = await res.json();

                // Mocking response for now to ensure UI works without running backend with apps installed
                await new Promise(r => setTimeout(r, 500));

                // Simulate success
                const mockLayout: DashboardLayout = {
                    id: dashboardId,
                    title: `${appName} Overview`,
                    description: "Main dashboard for this application",
                    rows: [
                        {
                            height: "auto",
                            columns: [
                                {
                                    width: 3,
                                    panels: [{ id: "p1", title: "Total Requests", type: "stat", query: "stats count" }]
                                },
                                {
                                    width: 3,
                                    panels: [{ id: "p2", title: "Error Rate", type: "stat", query: "stats avg(error)" }]
                                },
                                {
                                    width: 6,
                                    panels: [{ id: "p3", title: "Traffic Volume", type: "area", query: "timechart count" }]
                                }
                            ]
                        },
                        {
                            height: "auto",
                            columns: [
                                {
                                    width: 6,
                                    panels: [{ id: "p4", title: "Latency Distribution", type: "bar", query: "stats count by latency" }]
                                },
                                {
                                    width: 6,
                                    panels: [{ id: "p5", title: "Top Endpoints", type: "line", query: "timechart count by endpoint" }]
                                }
                            ]
                        }
                    ]
                };

                setLayout(mockLayout);
            } catch (err: any) {
                setError(err.message || 'Failed to load dashboard');
            } finally {
                setLoading(false);
            }
        };

        fetchDashboard();
    }, [appName, dashboardId]);

    if (loading) {
        return (
            <div className="p-8 text-center text-slate-500 animate-pulse">
                Loading dashboard definition...
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-8 flex items-center justify-center text-red-400 gap-2">
                <AlertCircle className="w-5 h-5" />
                <span>Error: {error}</span>
            </div>
        );
    }

    if (!layout) return null;

    return (
        <div className="p-6">
            <DashboardRenderer layout={layout} />
        </div>
    );
};

export default AppDashboardPage;
