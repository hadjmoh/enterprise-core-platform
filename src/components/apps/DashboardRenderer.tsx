import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import type { DashboardLayout, DashboardPanel, DashboardRow, DashboardColumn } from '@/types/apps';
import { ResponsiveContainer, LineChart, Line, BarChart, Bar, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip } from 'recharts';

interface DashboardRendererProps {
    layout: DashboardLayout;
}

export const DashboardRenderer: React.FC<DashboardRendererProps> = ({ layout }) => {
    return (
        <div className="space-y-6">
            <div className="flex flex-col gap-1">
                <h1 className="text-2xl font-bold text-slate-100">{layout.title}</h1>
                <p className="text-sm text-slate-400">{layout.description}</p>
            </div>

            <div className="flex flex-col gap-6">
                {layout.rows.map((row: DashboardRow, rIdx: number) => (
                    <div
                        key={rIdx}
                        className="grid gap-6"
                        style={{ gridTemplateColumns: `repeat(12, minmax(0, 1fr))` }}
                    >
                        {row.columns.map((col: DashboardColumn, cIdx: number) => (
                            <div
                                key={cIdx}
                                className="col-span-12 md:col-span-6 lg:col-span-4"
                                style={{ gridColumn: `span ${col.width} / span ${col.width}` }}
                            >
                                {
                                    col.panels.map((panel: DashboardPanel) => (
                                        <PanelRenderer key={panel.id} panel={panel} />
                                    ))
                                }
                            </div >
                        ))}
                    </div >
                ))}
            </div >
        </div >
    );
};

const PanelRenderer: React.FC<{ panel: DashboardPanel }> = ({ panel }) => {
    // In a real implementation, we would fetch data here using panel.query
    // For now, we'll use mock data to demonstrate rendering.
    const [data] = useState(generateMockData());
    const [loading] = useState(false);

    if (loading) {
        return (
            <Card className="h-full border-slate-800 bg-slate-900/50">
                <div className="flex items-center justify-center h-[300px] text-slate-500">
                    Loading...
                </div>
            </Card>
        );
    }

    return (
        <Card className="h-full border-slate-800 bg-slate-900/50">
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-slate-300 uppercase tracking-wider">
                    {panel.title}
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="h-[300px] w-full">
                    {renderChart(panel, data)}
                </div>
            </CardContent>
        </Card>
    );
};

type ChartData = { name: string; value: number }[];

const renderChart = (panel: DashboardPanel, data: ChartData) => {
    const commonProps = {
        data,
        margin: { top: 10, right: 30, left: 0, bottom: 0 }
    };

    switch (panel.type) {
        case 'line':
            return (
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart {...commonProps}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                        <XAxis dataKey="name" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155' }}
                            itemStyle={{ color: '#e2e8f0' }}
                        />
                        <Line type="monotone" dataKey="value" stroke="#3b82f6" strokeWidth={2} dot={false} activeDot={{ r: 4 }} />
                    </LineChart>
                </ResponsiveContainer>
            );
        case 'bar':
            return (
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart {...commonProps}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                        <XAxis dataKey="name" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155' }}
                            itemStyle={{ color: '#e2e8f0' }}
                        />
                        <Bar dataKey="value" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
                    </BarChart>
                </ResponsiveContainer>
            );
        case 'area':
            return (
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart {...commonProps}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                        <XAxis dataKey="name" stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} axisLine={false} />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155' }}
                            itemStyle={{ color: '#e2e8f0' }}
                        />
                        <Area type="monotone" dataKey="value" stroke="#10b981" fill="#10b981" fillOpacity={0.2} />
                    </AreaChart>
                </ResponsiveContainer>
            );
        case 'stat':
            return (
                <div className="flex flex-col items-center justify-center h-full">
                    <span className="text-4xl font-bold text-white">{data[data.length - 1].value}</span>
                    <span className="text-sm text-green-400 mt-2 flex items-center gap-1">
                        ▲ 12% vs last hour
                    </span>
                </div>
            );
        default:
            return <div className="text-slate-500">Unsupported panel type: {panel.type}</div>;
    }
};

const generateMockData = (): ChartData => {
    // Generate simple time series data
    return Array.from({ length: 24 }, (_, i) => ({
        name: `${i}:00`,
        value: Math.floor(Math.random() * 100) + 20
    }));
};
