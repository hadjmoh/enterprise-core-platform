import React, { useMemo } from 'react';
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    TimeScale,
    TimeSeriesScale,
    Decimation,
    type ChartOptions,
    type ChartData as ChartJSData,
    Filler
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import zoomPlugin from 'chartjs-plugin-zoom';
import { useUI } from '../../hooks/useUI';
import { AlertCircle } from 'lucide-react';

// Advanced Chart.js registration
ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    TimeScale,
    TimeSeriesScale,
    Decimation,
    Title,
    Tooltip,
    Legend,
    Filler,
    zoomPlugin
);

import type { Threshold } from '../../types/viz';

interface TimeSeriesChartProps {
    data: Record<string, unknown>[];
    span?: number;
    onDrillDown?: (filters: Record<string, unknown>, x: number, y: number) => void;
    thresholds?: Threshold[];
}

interface WorkerDataset {
    label: string;
    data: { x: number | string, y: number }[];
    [key: string]: unknown;
}

export const TimeSeriesChartImpl: React.FC<TimeSeriesChartProps> = ({ data, span, onDrillDown, thresholds }) => {
    const { timezone } = useUI();
    const isHighDensity = data.length > 500;
    const [workerDatasets, setWorkerDatasets] = React.useState<WorkerDataset[]>([]);
    const workerRef = React.useRef<Worker | null>(null);

    React.useEffect(() => {
        // Initialize worker
        workerRef.current = new Worker(
            new URL('../../workers/dataMapper.worker.ts', import.meta.url),
            { type: 'module' }
        );

        workerRef.current.onmessage = (e) => {
            setWorkerDatasets(e.data.datasets);
        };

        return () => {
            workerRef.current?.terminate();
        };
    }, []);

    React.useEffect(() => {
        if (!data || data.length === 0) {
            setWorkerDatasets([]);
            return;
        }

        const keys = Object.keys(data[0]).filter(k => k !== '_time' && !k.startsWith('_'));
        workerRef.current?.postMessage({ data, keys });
    }, [data, span, onDrillDown, timezone]);

    const chartData = useMemo(() => {
        const baseDatasets = workerDatasets.map((ds, i) => ({
            ...ds,
            borderColor: i === 0 ? '#6366f1' : '#10b981', // indigo / emerald
            backgroundColor: i === 0 ? 'rgba(99, 102, 241, 0.1)' : 'rgba(16, 185, 129, 0.1)',
            tension: isHighDensity ? 0 : 0.2,
            pointRadius: isHighDensity ? 0 : 2,
            borderWidth: isHighDensity ? 1 : 2,
            fill: true,
            spanGaps: true,
            data: ds.data as unknown as (number | null)[]
        })) as unknown as unknown[];

        const datasets = [...baseDatasets];

        // Add threshold baseline datasets
        if (thresholds && thresholds.length > 0 && workerDatasets.length > 0) {
            const timeRange = workerDatasets[0].data.map(d => d.x);
            thresholds.forEach(t => {
                datasets.push({
                    label: t.label || 'Threshold',
                    data: timeRange.map(x => ({ x, y: t.value })),
                    borderColor: t.color,
                    borderWidth: 1,
                    borderDash: [5, 5],
                    pointRadius: 0,
                    fill: false,
                    tension: 0
                });
            });
        }

        return { datasets: datasets as ChartJSData<'line'>['datasets'] };
    }, [workerDatasets, isHighDensity, thresholds]);

    const alertZonePlugin = useMemo(() => ({
        id: 'alertZone',
        beforeDraw: (chart: ChartJS<'line'>) => {
            if (!thresholds) return;
            const { ctx, chartArea: { top, bottom, left, right }, scales: { y } } = chart;
            if (!y) return;

            thresholds.forEach(t => {
                if (t.severity === 'critical' || t.severity === 'warning') {
                    const yPos = y.getPixelForValue(t.value);
                    if (yPos < bottom && yPos > top) {
                        ctx.save();
                        ctx.fillStyle = t.severity === 'critical' ? 'rgba(239, 68, 68, 0.05)' : 'rgba(251, 191, 36, 0.05)';
                        ctx.fillRect(left, top, right - left, yPos - top);
                        ctx.restore();
                    }
                }
            });
        }
    }), [thresholds]);

    const options: ChartOptions<'line'> = useMemo(() => ({
        responsive: true,
        maintainAspectRatio: false,
        animation: isHighDensity ? (false as const) : undefined,
        parsing: false,
        normalized: true,
        plugins: {
            decimation: {
                enabled: true,
                algorithm: 'min-max',
                threshold: 500,
            },
            legend: {
                position: 'top' as const,
                align: 'end' as const,
                labels: {
                    color: '#94a3b8',
                    font: { size: 10, family: 'Inter' },
                    usePointStyle: true,
                    boxWidth: 6,
                    padding: 15,
                },
                maxHeight: 40, // Limit legend height
            },
            tooltip: {
                enabled: true,
                mode: 'index',
                intersect: false,
                backgroundColor: '#0f172a',
                titleColor: '#f8fafc',
                bodyColor: '#94a3b8',
                borderColor: '#1e293b',
                borderWidth: 1,
                padding: 12,
                cornerRadius: 8,
                itemSort: (a, b) => (b.raw as number) - (a.raw as number),
                callbacks: {
                    label: function (context) {
                        const chart = context.chart;
                        const dataIndex = context.dataIndex;
                        const datasets = chart.data.datasets;

                        // If it's one of the top 10, show it
                        const values = datasets.map((ds, i) => {
                            const raw = ds.data[dataIndex];
                            let val = 0;
                            if (raw !== undefined && raw !== null) {
                                if (typeof raw === 'object' && 'y' in raw) {
                                    val = (raw as { y: number }).y;
                                } else {
                                    val = raw as number;
                                }
                            }
                            return { index: i, val };
                        }).sort((a, b) => b.val - a.val);

                        const rank = values.findIndex(v => v.index === context.datasetIndex);
                        if (rank < 10) {
                            return `${context.dataset.label}: ${context.formattedValue}`;
                        }

                        // Summary for the rest
                        if (rank === 10) {
                            const remainingSum = values.slice(10).reduce((acc, v) => acc + v.val, 0);
                            return `... and ${values.length - 10} others: ${remainingSum.toFixed(2)}`;
                        }
                        return '';
                    }
                }
            },
            zoom: {
                pan: {
                    enabled: true,
                    mode: 'x',
                },
                zoom: {
                    wheel: {
                        enabled: true,
                    },
                    pinch: {
                        enabled: true
                    },
                    mode: 'x',
                    drag: {
                        enabled: true,
                        backgroundColor: 'rgba(99, 102, 241, 0.1)',
                        borderColor: 'rgb(99, 102, 241)',
                        borderWidth: 1,
                    },
                }
            }
        },
        scales: {
            x: {
                type: 'time',
                time: {
                    unit: span && span < 60 ? 'second' : (span && span < 3600 ? 'minute' : 'hour'),
                    displayFormats: {
                        second: 'HH:mm:ss',
                        minute: 'HH:mm',
                        hour: 'MMM d, HH:mm'
                    }
                },
                adapters: {
                    date: {
                        useUTC: timezone === 'UTC'
                    }
                },
                grid: { color: '#1e293b' },
                ticks: { color: '#64748b', font: { size: 10 }, maxRotation: 0 }
            },
            y: {
                beginAtZero: true,
                grid: { color: '#1e293b' },
                ticks: { color: '#64748b', font: { size: 10 } }
            }
        },
        interaction: {
            mode: 'nearest',
            axis: 'x',
            intersect: false
        },
        onClick: (_event, elements, chart) => {
            if (!onDrillDown || elements.length === 0) return;
            const element = elements[0];
            const datasetIndex = element.datasetIndex;
            const index = element.index;

            const dataset = chart.data.datasets[datasetIndex];
            const seriesName = dataset.label;
            const dataPoint = dataset.data[index];
            const value = typeof dataPoint === 'object' && dataPoint !== null && 'y' in dataPoint ? dataPoint.y : dataPoint;

            if (seriesName && seriesName !== 'Value' && seriesName !== 'count') {
                onDrillDown({ [seriesName]: value }, _event.native ? (_event.native as MouseEvent).clientX : 0, _event.native ? (_event.native as MouseEvent).clientY : 0);
            }
        }
    }), [isHighDensity, span, onDrillDown, timezone]);

    if (!data || data.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-slate-600">
                <AlertCircle className="w-8 h-8 mb-2 opacity-20" />
                <p className="text-[10px] font-bold uppercase tracking-[0.2em]">No Time Series Data</p>
            </div>
        );
    }

    return (
        <div
            className="w-full h-full relative focus:outline-none focus:ring-2 focus:ring-brand-500/50 rounded-lg cursor-pointer"
            tabIndex={0}
            role="button"
            aria-label="Interactive Time Series Chart. Use mouse to zoom/pan. Press Enter for tactical drill-down."
            onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    // For accessibility, we trigger a generic drill-down if specific points aren't selectable via keyboard
                    if (onDrillDown && workerDatasets.length > 0) {
                        const firstDS = workerDatasets[0];
                        if (firstDS.data.length > 0) {
                            onDrillDown({ [firstDS.label]: firstDS.data[0].y }, 0, 0);
                        }
                    }
                }
            }}
        >
            <Line data={chartData} options={options} plugins={[alertZonePlugin]} />
        </div>
    );
};

export const TimeSeriesChart = React.memo(TimeSeriesChartImpl);
