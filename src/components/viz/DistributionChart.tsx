import React, { useMemo } from 'react';
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    BarElement,
    ArcElement,
    Title,
    Tooltip,
    Legend,
    type ChartOptions,
    type ChartData as ChartJSData
} from 'chart.js';
import { Bar, Pie } from 'react-chartjs-2';

ChartJS.register(
    CategoryScale,
    LinearScale,
    BarElement,
    ArcElement,
    Title,
    Tooltip,
    Legend
);

import type { Threshold } from '../../types/viz';

interface DistributionChartProps {
    data: Record<string, unknown>[];
    type: 'bar' | 'pie';
    title?: string;
    onDrillDown?: (filters: Record<string, unknown>, x: number, y: number) => void;
    thresholds?: Threshold[];
}

const DistributionChartImpl: React.FC<DistributionChartProps> = ({ data, type, title, onDrillDown, thresholds }) => {
    const palette = useMemo(() => [
        'rgba(99, 102, 241, 0.8)', // indigo
        'rgba(16, 185, 129, 0.8)', // emerald
        'rgba(245, 158, 11, 0.8)', // amber
        'rgba(239, 68, 68, 0.8)',   // red
        'rgba(139, 92, 246, 0.8)',  // violet
        'rgba(20, 184, 166, 0.8)',  // teal
        'rgba(236, 72, 153, 0.8)',  // pink
        'rgba(14, 165, 233, 0.8)',  // sky
    ], []);

    const schema = useMemo(() => {
        if (!data || data.length === 0) return { labelKey: '', valueKeys: [] };
        const firstRow = data[0];
        const allKeys = Object.keys(firstRow).filter(k => !k.startsWith('_'));

        const labelKey = allKeys.find(k => typeof firstRow[k] === 'string') || allKeys[0];
        const valueKeys = allKeys.filter(k => k !== labelKey && typeof firstRow[k] === 'number');

        return { labelKey, valueKeys };
    }, [data]);

    const chartData = useMemo(() => {
        const { labelKey, valueKeys } = schema;
        if (!labelKey || valueKeys.length === 0) return { labels: [], datasets: [] };

        const labels = data.map(d => String(d[labelKey]));

        const datasets = valueKeys.map((vKey: string, vIdx: number) => {
            const baseColor = palette[(vIdx + 2) % palette.length];

            return {
                label: vKey,
                data: data.map((d: Record<string, unknown>) => d[vKey]),
                backgroundColor: data.map((d: Record<string, unknown>) => {
                    const val = d[vKey] as number;
                    const activeThreshold = thresholds
                        ?.sort((a, b) => b.value - a.value)
                        .find(t => val >= t.value);

                    if (activeThreshold) return activeThreshold.color;

                    if (type === 'pie') {
                        const labelValue = String(d[labelKey]);
                        let hash = 0;
                        for (let j = 0; j < labelValue.length; j++) {
                            hash = labelValue.charCodeAt(j) + ((hash << 5) - hash);
                        }
                        const colorIdx = Math.abs(hash) % palette.length;
                        return palette[colorIdx];
                    }
                    return baseColor;
                }),
                borderColor: type === 'pie' ? 'rgba(15, 23, 42, 1)' : baseColor.replace('0.8', '1'),
                borderWidth: 1,
                borderRadius: type === 'bar' ? 4 : 0,
                hoverOffset: 4,
            };
        });

        return { labels, datasets };
    }, [data, type, palette, schema, thresholds]);

    const options: ChartOptions<'bar' | 'pie'> = useMemo(() => ({
        responsive: true,
        maintainAspectRatio: false,
        onClick: (_event, elements, chart) => {
            if (!onDrillDown || elements.length === 0) return;
            const element = elements[0];
            const index = element.index;
            const label = chart.data.labels?.[index];
            const { labelKey } = schema;

            if (labelKey && label) {
                onDrillDown({ [labelKey]: label }, _event.native ? (_event.native as MouseEvent).clientX : 0, _event.native ? (_event.native as MouseEvent).clientY : 0);
            }
        },
        plugins: {
            title: {
                display: !!title,
                text: title,
                color: '#f8fafc',
                font: { size: 14, weight: 'bold' }
            },
            legend: {
                display: true,
                position: type === 'pie' ? 'right' as const : 'top' as const,
                labels: {
                    color: '#94a3b8',
                    font: { size: 10, family: 'Inter' },
                    usePointStyle: true,
                    padding: 20
                }
            },
            tooltip: {
                backgroundColor: '#0f172a',
                titleColor: '#f8fafc',
                bodyColor: '#94a3b8',
                borderColor: '#1e293b',
                borderWidth: 1,
                padding: 12,
                cornerRadius: 8,
                callbacks: {
                    label: function (context) {
                        const chart = context.chart;
                        const dataIndex = context.dataIndex;
                        const datasets = chart.data.datasets;

                        if (datasets.length === 1) {
                            return `${context.dataset.label || ''}: ${context.formattedValue}`;
                        }

                        const values = datasets.map((ds, i) => ({
                            index: i,
                            val: ds.data[dataIndex] as number
                        })).sort((a, b) => b.val - a.val);

                        const rank = values.findIndex(v => v.index === context.datasetIndex);
                        if (rank < 10) {
                            return `${context.dataset.label}: ${context.formattedValue}`;
                        }

                        if (rank === 10) {
                            const remainingSum = values.slice(10).reduce((acc, v) => acc + v.val, 0);
                            return `... and ${values.length - 10} others: ${remainingSum.toFixed(2)}`;
                        }
                        return '';
                    }
                }
            }
        },
        scales: type === 'bar' ? {
            x: {
                grid: { display: false },
                ticks: { color: '#64748b', font: { size: 10 } }
            },
            y: {
                beginAtZero: true,
                grid: {
                    color: 'rgba(30, 41, 59, 0.5)'
                },
                ticks: { color: '#64748b', font: { size: 10 } }
            }
        } : undefined,
    }), [title, type, onDrillDown, schema]);

    return (
        <div
            className="w-full h-full relative focus:outline-none focus:ring-2 focus:ring-brand-500/50 rounded-lg cursor-pointer"
            tabIndex={0}
            role="button"
            aria-label={`Interactive ${type} chart: ${title || 'Data Distribution'}. Press Enter for tactical drill-down.`}
            onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    // For accessibility, we trigger a generic drill-down if specific points aren't selectable via keyboard
                    if (onDrillDown && data.length > 0) {
                        const firstRow = data[0];
                        const { labelKey } = schema;
                        if (labelKey) {
                            onDrillDown({ [labelKey]: firstRow[labelKey] }, 0, 0);
                        }
                    }
                }
            }}
        >
            {type === 'pie' ? (
                <Pie data={chartData as ChartJSData<'pie'>} options={options as ChartOptions<'pie'>} />
            ) : (
                <Bar data={chartData as ChartJSData<'bar'>} options={options as ChartOptions<'bar'>} />
            )}
        </div>
    );
};

export const DistributionChart = React.memo(DistributionChartImpl);
