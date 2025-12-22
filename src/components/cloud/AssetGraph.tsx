import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import type { CloudAsset } from '../../services/cloudService';

interface AssetGraphProps {
    assets: CloudAsset[];
    onNodeClick?: (asset: CloudAsset) => void;
    selectedAssetId?: string;
}

interface GraphNode extends d3.SimulationNodeDatum {
    id: string;
    asset: CloudAsset;
    x?: number;
    y?: number;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
    source: GraphNode | string;
    target: GraphNode | string;
    type: string;
}

export function AssetGraph({ assets, onNodeClick, selectedAssetId }: AssetGraphProps) {
    const svgRef = useRef<SVGSVGElement>(null);
    const dimensions = { width: 1200, height: 800 };

    useEffect(() => {
        if (!svgRef.current || assets.length === 0) return;

        const svg = d3.select(svgRef.current);
        svg.selectAll('*').remove(); // Clear previous render

        const { width, height } = dimensions;

        // Create nodes from assets
        const nodes: GraphNode[] = assets.map(asset => ({
            id: asset.id,
            asset
        }));

        // Create links based on relationships (VPC, region, provider)
        const links: GraphLink[] = [];

        // Link assets in same VPC
        for (let i = 0; i < assets.length; i++) {
            for (let j = i + 1; j < assets.length; j++) {
                const a1 = assets[i];
                const a2 = assets[j];

                if (a1.vpc && a1.vpc === a2.vpc) {
                    links.push({
                        source: a1.id,
                        target: a2.id,
                        type: 'vpc'
                    });
                } else if (a1.provider === a2.provider && a1.region === a2.region) {
                    links.push({
                        source: a1.id,
                        target: a2.id,
                        type: 'region'
                    });
                }
            }
        }

        // Create force simulation
        const simulation = d3.forceSimulation<GraphNode>(nodes)
            .force('link', d3.forceLink<GraphNode, GraphLink>(links)
                .id(d => d.id)
                .distance(d => d.type === 'vpc' ? 100 : 200))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collision', d3.forceCollide().radius(40));

        // Create container group
        const g = svg.append('g');

        // Add zoom behavior
        const zoom = d3.zoom<SVGSVGElement, unknown>()
            .scaleExtent([0.5, 3])
            .on('zoom', (event) => {
                g.attr('transform', event.transform);
            });

        svg.call(zoom);

        // Draw links
        const link = g.append('g')
            .selectAll('line')
            .data(links)
            .join('line')
            .attr('stroke', d => d.type === 'vpc' ? '#3b82f6' : '#64748b')
            .attr('stroke-opacity', d => d.type === 'vpc' ? 0.6 : 0.3)
            .attr('stroke-width', d => d.type === 'vpc' ? 2 : 1)
            .attr('stroke-dasharray', d => d.type === 'vpc' ? '0' : '4,4');

        // Draw nodes
        const node = g.append('g')
            .selectAll('g')
            .data(nodes)
            .join('g')
            .attr('cursor', 'pointer')
            .call(d3.drag<SVGGElement, GraphNode>()
                .on('start', dragstarted)
                .on('drag', dragged)
                .on('end', dragended) as any);

        // Add circles for nodes
        node.append('circle')
            .attr('r', d => {
                if (d.id === selectedAssetId) return 25;
                if (d.asset.risk_score >= 70) return 20;
                return 15;
            })
            .attr('fill', d => getNodeColor(d.asset))
            .attr('stroke', d => d.id === selectedAssetId ? '#fbbf24' : '#1e293b')
            .attr('stroke-width', d => d.id === selectedAssetId ? 3 : 2)
            .on('click', (event, d) => {
                event.stopPropagation();
                onNodeClick?.(d.asset);
            });

        // Add provider icons as text
        node.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', '0.35em')
            .attr('font-size', '16px')
            .text(d => getProviderIcon(d.asset.provider))
            .attr('pointer-events', 'none');

        // Add labels
        node.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', '2em')
            .attr('font-size', '10px')
            .attr('fill', '#cbd5e1')
            .text(d => d.asset.name.length > 15 ? d.asset.name.substring(0, 15) + '...' : d.asset.name)
            .attr('pointer-events', 'none');

        // Add risk indicator
        node.filter(d => d.asset.risk_score >= 70)
            .append('circle')
            .attr('r', 5)
            .attr('cx', 15)
            .attr('cy', -15)
            .attr('fill', '#ef4444')
            .attr('stroke', '#1e293b')
            .attr('stroke-width', 1);

        // Update positions on tick
        simulation.on('tick', () => {
            link
                .attr('x1', d => (d.source as GraphNode).x!)
                .attr('y1', d => (d.source as GraphNode).y!)
                .attr('x2', d => (d.target as GraphNode).x!)
                .attr('y2', d => (d.target as GraphNode).y!);

            node.attr('transform', d => `translate(${d.x},${d.y})`);
        });

        // Drag functions
        function dragstarted(event: d3.D3DragEvent<SVGGElement, GraphNode, GraphNode>) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        function dragged(event: d3.D3DragEvent<SVGGElement, GraphNode, GraphNode>) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        function dragended(event: d3.D3DragEvent<SVGGElement, GraphNode, GraphNode>) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }

        // Cleanup
        return () => {
            simulation.stop();
        };
    }, [assets, onNodeClick, selectedAssetId, dimensions]);

    // Helper functions
    const getNodeColor = (asset: CloudAsset): string => {
        // Color by provider
        if (asset.provider === 'aws') return '#ff9900';
        if (asset.provider === 'azure') return '#0078d4';
        if (asset.provider === 'gcp') return '#4285f4';
        return '#64748b';
    };

    const getProviderIcon = (provider: string): string => {
        if (provider === 'aws') return '🟠';
        if (provider === 'azure') return '🔵';
        if (provider === 'gcp') return '🟢';
        return '⚪';
    };

    return (
        <div className="relative w-full h-full bg-slate-950 rounded-xl border border-slate-800 overflow-hidden">
            {/* Legend */}
            <div className="absolute top-4 left-4 bg-slate-900/90 backdrop-blur-sm border border-slate-800 rounded-lg p-3 z-10">
                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-2">Legend</h3>
                <div className="space-y-2 text-xs">
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-0.5 bg-blue-500"></div>
                        <span className="text-slate-400">Same VPC</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-0.5 bg-slate-600 border-dashed border-t"></div>
                        <span className="text-slate-400">Same Region</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full bg-red-500"></div>
                        <span className="text-slate-400">High Risk</span>
                    </div>
                </div>
            </div>

            {/* Controls */}
            <div className="absolute top-4 right-4 bg-slate-900/90 backdrop-blur-sm border border-slate-800 rounded-lg p-2 z-10">
                <div className="text-[10px] text-slate-500 font-medium mb-1">Controls</div>
                <div className="text-[9px] text-slate-600 space-y-0.5">
                    <div>• Drag nodes to reposition</div>
                    <div>• Scroll to zoom</div>
                    <div>• Click node for details</div>
                </div>
            </div>

            {/* Stats */}
            <div className="absolute bottom-4 left-4 bg-slate-900/90 backdrop-blur-sm border border-slate-800 rounded-lg p-3 z-10">
                <div className="text-xs space-y-1">
                    <div className="flex items-center justify-between gap-4">
                        <span className="text-slate-500">Nodes:</span>
                        <span className="text-white font-bold">{assets.length}</span>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                        <span className="text-slate-500">Providers:</span>
                        <div className="flex gap-1">
                            {assets.some(a => a.provider === 'aws') && <span>🟠</span>}
                            {assets.some(a => a.provider === 'azure') && <span>🔵</span>}
                            {assets.some(a => a.provider === 'gcp') && <span>🟢</span>}
                        </div>
                    </div>
                </div>
            </div>

            {/* SVG Canvas */}
            <svg
                ref={svgRef}
                width={dimensions.width}
                height={dimensions.height}
                className="w-full h-full"
            />

            {/* Empty State */}
            {assets.length === 0 && (
                <div className="absolute inset-0 flex items-center justify-center">
                    <div className="text-center">
                        <div className="text-6xl mb-4">🌐</div>
                        <p className="text-slate-500 font-medium">No assets to visualize</p>
                        <p className="text-xs text-slate-600 mt-1">Add filters to see asset relationships</p>
                    </div>
                </div>
            )}
        </div>
    );
}
