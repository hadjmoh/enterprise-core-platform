import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import {
    Shield,
    Link as LinkIcon,
    Share2,
    Search,
    History,
    AlertCircle,
    CheckCircle2,
    Database,
    Fingerprint
} from 'lucide-react';
import { LineageFlow } from '../components/visualizations/LineageFlow';

interface Node extends d3.SimulationNodeDatum {
    id: string;
    type: string;
    trust_level: string;
    last_seen: string;
}

interface Edge extends d3.SimulationLinkDatum<Node> {
    source: string | Node;
    target: string | Node;
    relation: string;
    trust_level: string;
}

export const TrustGraphPage: React.FC = () => {
    const svgRef = useRef<SVGSVGElement>(null);
    const [summary, setSummary] = useState<any>(null);
    const [selectedNode, setSelectedNode] = useState<Node | null>(null);
    const [lineageQuery, setLineageQuery] = useState('');
    const [lineageResults, setLineageResults] = useState<any | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 10000);
        return () => clearInterval(interval);
    }, []);

    const fetchData = async () => {
        try {
            const res = await fetch('/api/v1/trust/graph');
            const data = await res.json();

            // Render Graph
            if (svgRef.current) {
                renderGraph(data.nodes, data.edges);
            }

            // Calculate simple summary
            const distribution = data.nodes.reduce((acc: any, n: any) => {
                acc[n.trust_level] = (acc[n.trust_level] || 0) + 1;
                return acc;
            }, {});

            setSummary({
                total_nodes: data.nodes.length,
                total_edges: data.edges.length,
                distribution
            });
        } catch (err) {
            console.error("Failed to fetch trust graph", err);
        }
    };

    const renderGraph = (nodes: Node[], edges: Edge[]) => {
        const svg = d3.select(svgRef.current);
        const width = svgRef.current?.clientWidth || 800;
        const height = 400;

        svg.selectAll("*").remove();

        const simulation = d3.forceSimulation<Node>(nodes)
            .force("link", d3.forceLink<Node, Edge>(edges).id(d => d.id).distance(100))
            .force("charge", d3.forceManyBody().strength(-200))
            .force("center", d3.forceCenter(width / 2, height / 2));

        const link = svg.append("g")
            .attr("stroke", "#334155")
            .attr("stroke-opacity", 0.6)
            .selectAll("line")
            .data(edges)
            .join("line")
            .attr("stroke-width", 1.5)
            .attr("stroke-dasharray", d => d.trust_level === 'untrusted' ? '4 4' : '0');

        const node = svg.append("g")
            .selectAll("g")
            .data(nodes)
            .join("g")
            .call(d3.drag<SVGGElement, Node>()
                .on("start", dragstarted)
                .on("drag", dragged)
                .on("end", dragended) as any);

        node.append("circle")
            .attr("r", 12)
            .attr("fill", d => {
                switch (d.trust_level) {
                    case 'verified': return '#10b981';
                    case 'reputable': return '#6366f1';
                    case 'suspicious': return '#f59e0b';
                    case 'untrusted': return '#ef4444';
                    default: return '#64748b';
                }
            })
            .attr("stroke", "#1e293b")
            .attr("stroke-width", 2)
            .on("click", (_, d) => setSelectedNode(d));

        node.append("text")
            .text(d => d.id)
            .attr("x", 16)
            .attr("y", 4)
            .attr("fill", "#94a3b8")
            .attr("font-size", "10px")
            .attr("font-family", "monospace");

        simulation.on("tick", () => {
            link
                .attr("x1", d => (d.source as Node).x!)
                .attr("y1", d => (d.source as Node).y!)
                .attr("x2", d => (d.target as Node).x!)
                .attr("y2", d => (d.target as Node).y!);

            node
                .attr("transform", d => `translate(${d.x},${d.y})`);
        });

        function dragstarted(event: any) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        function dragged(event: any) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        function dragended(event: any) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }
    };

    const handleSearchLineage = async () => {
        if (!lineageQuery) return;
        setIsLoading(true);
        try {
            const res = await fetch(`/api/v1/lineage/event?id=${lineageQuery}`);
            const data = await res.json();
            setLineageResults(data);
        } catch (err) {
            console.error("Lineage lookup failed", err);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="p-6 space-y-6 bg-slate-900 min-h-screen">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                        <Fingerprint className="w-8 h-8 text-indigo-400" />
                        Trust Graph & Data Lineage
                    </h1>
                    <p className="text-slate-400 mt-1">Verifiable chain-of-custody and entity relationship mapping.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm">
                        <History className="w-4 h-4 mr-2" /> Audit Logs
                    </Button>
                    <Button variant="primary" size="sm">
                        <Shield className="w-4 h-4 mr-2" /> Rotate Keys
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Stats Summary */}
                <Card className="p-4 bg-slate-800/40 border-slate-700">
                    <div className="flex items-center gap-2 mb-4">
                        <Database className="w-4 h-4 text-indigo-400" />
                        <h3 className="font-bold text-white">Trust Distribution</h3>
                    </div>
                    {summary ? (
                        <div className="space-y-3">
                            {Object.entries(summary.distribution).map(([level, count]: [any, any]) => (
                                <div key={level} className="flex justify-between items-center">
                                    <Badge variant={level === 'verified' ? 'success' : level === 'untrusted' ? 'error' : 'outline'}>
                                        {level}
                                    </Badge>
                                    <span className="text-white font-mono">{count}</span>
                                </div>
                            ))}
                        </div>
                    ) : <div className="animate-pulse h-20 bg-slate-700 rounded" />}
                </Card>

                <Card className="p-4 bg-slate-800/40 border-slate-700">
                    <div className="flex items-center gap-2 mb-4">
                        <Share2 className="w-4 h-4 text-emerald-400" />
                        <h3 className="font-bold text-white">Graph Connectivity</h3>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <p className="text-xs text-slate-500 uppercase">Total Nodes</p>
                            <p className="text-2xl font-bold text-white">{summary?.total_nodes || 0}</p>
                        </div>
                        <div>
                            <p className="text-xs text-slate-500 uppercase">Active Flows</p>
                            <p className="text-2xl font-bold text-white">{summary?.total_edges || 0}</p>
                        </div>
                    </div>
                </Card>

                <Card className="p-4 bg-slate-800/40 border-slate-700">
                    <div className="flex items-center gap-2 mb-4">
                        <CheckCircle2 className="w-4 h-4 text-blue-400" />
                        <h3 className="font-bold text-white">Integrity Status</h3>
                    </div>
                    <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-emerald-500/20 flex items-center justify-center">
                            <Shield className="w-6 h-6 text-emerald-400" />
                        </div>
                        <div>
                            <p className="text-sm font-bold text-white">Continuous Verification</p>
                            <p className="text-xs text-emerald-400/70">100% of pipeline stages signed.</p>
                        </div>
                    </div>
                </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Main Graph View */}
                <div className="lg:col-span-8 flex flex-col gap-6">
                    <Card className="p-0 bg-slate-900 border-slate-700 overflow-hidden relative">
                        <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                            <h3 className="text-sm font-bold text-slate-300 flex items-center gap-2">
                                <LinkIcon className="w-4 h-4" /> Global Trust Topography
                            </h3>
                            <div className="flex gap-2">
                                <div className="flex items-center gap-1">
                                    <div className="w-2 h-2 rounded-full bg-emerald-500" />
                                    <span className="text-[10px] text-slate-400">Verified</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <div className="w-2 h-2 rounded-full bg-indigo-500" />
                                    <span className="text-[10px] text-slate-400">Reputable</span>
                                </div>
                            </div>
                        </div>
                        <svg ref={svgRef} className="w-full h-[400px] cursor-crosshair" />

                        {selectedNode && (
                            <div className="absolute top-16 right-4 w-64 bg-slate-800 border border-slate-700 rounded-lg shadow-2xl p-4 animate-in fade-in slide-in-from-right">
                                <div className="flex justify-between items-start mb-4">
                                    <h4 className="font-bold text-white">{selectedNode.id}</h4>
                                    <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setSelectedNode(null)}>
                                        ×
                                    </Button>
                                </div>
                                <div className="space-y-2 text-xs">
                                    <div className="flex justify-between">
                                        <span className="text-slate-500 uppercase">Type</span>
                                        <span className="text-slate-200 capitalize">{selectedNode.type}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-slate-500 uppercase">Trust Level</span>
                                        <Badge variant={selectedNode.trust_level === 'verified' ? 'success' : 'outline'}>
                                            {selectedNode.trust_level}
                                        </Badge>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-slate-500 uppercase">Last Activity</span>
                                        <span className="text-slate-200">{new Date(selectedNode.last_seen).toLocaleTimeString()}</span>
                                    </div>
                                </div>
                                <Button variant="primary" size="sm" className="w-full mt-4" onClick={() => {
                                    setLineageQuery(selectedNode.id);
                                    handleSearchLineage();
                                }}>
                                    Explore Lineage
                                </Button>
                            </div>
                        )}
                    </Card>

                    <Card className="p-6 bg-slate-800/40 border-slate-700">
                        <div className="flex items-center gap-2 mb-6">
                            <Search className="w-4 h-4 text-indigo-400" />
                            <h3 className="font-bold text-white">Lineage Explorer</h3>
                        </div>
                        <div className="flex gap-2 mb-8">
                            <div className="relative flex-1">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                                <input
                                    type="text"
                                    placeholder="Enter Event ID or Hash to trace..."
                                    className="w-full bg-slate-900 border-slate-700 rounded-lg py-2 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-indigo-500 outline-none"
                                    value={lineageQuery}
                                    onChange={(e) => setLineageQuery(e.target.value)}
                                    onKeyDown={(e) => e.key === 'Enter' && handleSearchLineage()}
                                />
                            </div>
                            <Button variant="primary" onClick={handleSearchLineage} disabled={isLoading}>
                                {isLoading ? 'Tracing...' : 'Trace Path'}
                            </Button>
                        </div>

                        {lineageResults ? (
                            <div className="animate-in fade-in duration-500">
                                <div className="mb-4 flex items-center justify-between">
                                    <h4 className="text-sm font-medium text-slate-300">Detailed Chain-of-Custody</h4>
                                    <Badge variant="success" className="bg-emerald-500/10 text-emerald-400 border-emerald-500/20">
                                        Verified Path
                                    </Badge>
                                </div>
                                <LineageFlow steps={lineageResults.steps} isVerified={lineageResults.verified} />
                            </div>
                        ) : (
                            <div className="h-40 flex flex-col items-center justify-center border-2 border-dashed border-slate-700 rounded-xl text-slate-500">
                                <History className="w-8 h-8 mb-2 opacity-20" />
                                <p className="text-sm">Search for an event to visualize its journey through the infrastructure.</p>
                            </div>
                        )}
                    </Card>
                </div>

                {/* Sidebar Alerts / Recent Activity */}
                <div className="lg:col-span-4 space-y-6">
                    <Card className="p-4 bg-slate-800/40 border-slate-700">
                        <h3 className="font-bold text-white mb-4 flex items-center gap-2">
                            <AlertCircle className="w-4 h-4 text-amber-400" />
                            Suspicious Data Flows
                        </h3>
                        <div className="space-y-3">
                            {[
                                { id: 'evt-8821', reason: 'Missing signature at enrichment', severity: 'high' },
                                { id: 'usr-analyst-02', reason: 'Sudden high volume push', severity: 'medium' }
                            ].map((alert, i) => (
                                <div key={i} className="p-3 bg-slate-900/50 border border-slate-800 rounded-lg hover:border-amber-500/30 transition-colors">
                                    <div className="flex justify-between items-start">
                                        <span className="text-[10px] font-mono text-indigo-400">{alert.id}</span>
                                        <Badge variant={alert.severity === 'high' ? 'error' : 'warning'} className="text-[8px] h-4">
                                            {alert.severity}
                                        </Badge>
                                    </div>
                                    <p className="text-xs text-slate-300 mt-1">{alert.reason}</p>
                                </div>
                            ))}
                        </div>
                    </Card>

                    <Card className="p-4 bg-slate-800/40 border-slate-700">
                        <h3 className="font-bold text-white mb-4">Verification Audit</h3>
                        <div className="space-y-4">
                            {[
                                { stage: 'Auth Service', status: 'Healthy', time: '1m ago' },
                                { stage: 'KMS Key Store', status: 'Healthy', time: '5m ago' },
                                { stage: 'Merkle Root', status: 'Updated', time: '12s ago' }
                            ].map((s, i) => (
                                <div key={i} className="flex justify-between items-center pb-3 border-b border-slate-800 last:border-0">
                                    <div className="flex items-center gap-3">
                                        <div className="w-2 h-2 rounded-full bg-emerald-500" />
                                        <span className="text-xs text-slate-300">{s.stage}</span>
                                    </div>
                                    <div className="text-right">
                                        <p className="text-[10px] text-emerald-400">{s.status}</p>
                                        <p className="text-[8px] text-slate-500">{s.time}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </Card>
                </div>
            </div>
        </div>
    );
};
