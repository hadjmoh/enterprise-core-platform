import { useState, useEffect } from 'react';
import { Sidebar } from '../components/layout/Sidebar';
import { Header } from '../components/layout/Header';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { Cloud, Server, Shield, AlertTriangle, Filter, Network } from 'lucide-react';
import { cloudService, type CloudAsset, type AssetFilter } from '../services/cloudService';
import { AssetGraph } from '../components/cloud/AssetGraph';

const providerIcons: Record<string, string> = {
    aws: '🟠',
    azure: '🔵',
    gcp: '🟢'
};

const providerColors: Record<string, string> = {
    aws: 'text-orange-400',
    azure: 'text-blue-400',
    gcp: 'text-green-400'
};

export function CloudAssetsPage() {
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [assets, setAssets] = useState<CloudAsset[]>([]);
    const [selectedAsset, setSelectedAsset] = useState<CloudAsset | null>(null);
    const [viewMode, setViewMode] = useState<'grid' | 'list' | 'graph'>('grid');
    const [loading, setLoading] = useState(true);
    const [showFilters, setShowFilters] = useState(false);

    const [filters, setFilters] = useState<AssetFilter>({
        provider: 'all',
        type: '',
        region: '',
        min_risk: undefined,
        max_risk: undefined
    });

    useEffect(() => {
        loadAssets();
    }, [filters]);

    const loadAssets = async () => {
        setLoading(true);
        try {
            const data = await cloudService.getAssets(filters);
            setAssets(data);
        } catch (error) {
            console.error('Failed to load assets:', error);
        } finally {
            setLoading(false);
        }
    };

    const getRiskBadgeVariant = (score: number): 'success' | 'warning' | 'error' => {
        if (score >= 70) return 'error';
        if (score >= 40) return 'warning';
        return 'success';
    };

    const getRiskLabel = (score: number): string => {
        if (score >= 70) return 'High Risk';
        if (score >= 40) return 'Medium Risk';
        return 'Low Risk';
    };

    return (
        <div className="flex h-screen bg-slate-950 text-slate-200">
            <Sidebar isOpen={isSidebarOpen} setIsOpen={setIsSidebarOpen} isMobile={false} />

            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Header
                    openAlerts={() => { }}
                    toggleSidebar={() => setIsSidebarOpen(!isSidebarOpen)}
                />

                <main className="flex-1 overflow-y-auto p-6 space-y-6">
                    {/* Header */}
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                                <div className="p-2 bg-brand-500/10 rounded-lg">
                                    <Cloud className="h-6 w-6 text-brand-400" />
                                </div>
                                Cloud Asset Inventory
                            </h1>
                            <p className="text-sm text-slate-500 mt-1">
                                Multi-cloud visibility across AWS, Azure, and GCP
                            </p>
                        </div>

                        <div className="flex items-center gap-2">
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setShowFilters(!showFilters)}
                                className="text-slate-400 hover:text-white"
                            >
                                <Filter className="h-4 w-4 mr-2" />
                                Filters
                            </Button>

                            <div className="flex gap-1 bg-slate-900 rounded-lg p-1">
                                <button
                                    onClick={() => setViewMode('grid')}
                                    className={`px-3 py-1.5 rounded text-xs font-medium transition-all ${viewMode === 'grid'
                                        ? 'bg-brand-500 text-white'
                                        : 'text-slate-400 hover:text-white'
                                        }`}
                                >
                                    Grid
                                </button>
                                <button
                                    onClick={() => setViewMode('list')}
                                    className={`px-3 py-1.5 rounded text-xs font-medium transition-all ${viewMode === 'list'
                                        ? 'bg-brand-500 text-white'
                                        : 'text-slate-400 hover:text-white'
                                        }`}
                                >
                                    List
                                </button>
                                <button
                                    onClick={() => setViewMode('graph')}
                                    className={`px-3 py-1.5 rounded text-xs font-medium transition-all flex items-center gap-1 ${viewMode === 'graph'
                                        ? 'bg-brand-500 text-white'
                                        : 'text-slate-400 hover:text-white'
                                        }`}
                                >
                                    <Network className="h-3 w-3" />
                                    Graph
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Filters */}
                    {showFilters && (
                        <Card className="p-4 bg-slate-900/50 border-slate-800">
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                                <div>
                                    <label htmlFor="filter-provider" className="text-xs font-medium text-slate-400 mb-2 block">Provider</label>
                                    <select
                                        id="filter-provider"
                                        value={filters.provider || 'all'}
                                        onChange={(e) => setFilters({ ...filters, provider: e.target.value as any })}
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-brand-500/50"
                                    >
                                        <option value="all">All Providers</option>
                                        <option value="aws">🟠 AWS</option>
                                        <option value="azure">🔵 Azure</option>
                                        <option value="gcp">🟢 GCP</option>
                                    </select>
                                </div>

                                <div>
                                    <label htmlFor="filter-type" className="text-xs font-medium text-slate-400 mb-2 block">Type</label>
                                    <select
                                        id="filter-type"
                                        value={filters.type || ''}
                                        onChange={(e) => setFilters({ ...filters, type: e.target.value })}
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-brand-500/50"
                                    >
                                        <option value="">All Types</option>
                                        <option value="ec2">EC2 Instance</option>
                                        <option value="vm">Virtual Machine</option>
                                        <option value="compute-instance">Compute Instance</option>
                                    </select>
                                </div>

                                <div>
                                    <label htmlFor="filter-min-risk" className="text-xs font-medium text-slate-400 mb-2 block">Min Risk</label>
                                    <input
                                        id="filter-min-risk"
                                        type="number"
                                        min="0"
                                        max="100"
                                        value={filters.min_risk || ''}
                                        onChange={(e) => setFilters({ ...filters, min_risk: e.target.value ? Number.parseInt(e.target.value, 10) : undefined })}
                                        placeholder="0"
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-brand-500/50"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="filter-max-risk" className="text-xs font-medium text-slate-400 mb-2 block">Max Risk</label>
                                    <input
                                        id="filter-max-risk"
                                        type="number"
                                        min="0"
                                        max="100"
                                        value={filters.max_risk || ''}
                                        onChange={(e) => setFilters({ ...filters, max_risk: e.target.value ? Number.parseInt(e.target.value, 10) : undefined })}
                                        placeholder="100"
                                        className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-brand-500/50"
                                    />
                                </div>
                            </div>
                        </Card>
                    )}

                    {/* Stats Overview */}
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <Card className="p-4 bg-slate-900/50 border-slate-800">
                            <div className="flex items-center justify-between">
                                <div>
                                    <p className="text-xs text-slate-500 font-medium">Total Assets</p>
                                    <p className="text-2xl font-bold text-white mt-1">{assets.length}</p>
                                </div>
                                <Server className="h-8 w-8 text-brand-400 opacity-50" />
                            </div>
                        </Card>

                        <Card className="p-4 bg-slate-900/50 border-slate-800">
                            <div className="flex items-center justify-between">
                                <div>
                                    <p className="text-xs text-slate-500 font-medium">AWS Assets</p>
                                    <p className="text-2xl font-bold text-orange-400 mt-1">
                                        {assets.filter(a => a.provider === 'aws').length}
                                    </p>
                                </div>
                                <span className="text-3xl">🟠</span>
                            </div>
                        </Card>

                        <Card className="p-4 bg-slate-900/50 border-slate-800">
                            <div className="flex items-center justify-between">
                                <div>
                                    <p className="text-xs text-slate-500 font-medium">Azure Assets</p>
                                    <p className="text-2xl font-bold text-blue-400 mt-1">
                                        {assets.filter(a => a.provider === 'azure').length}
                                    </p>
                                </div>
                                <span className="text-3xl">🔵</span>
                            </div>
                        </Card>

                        <Card className="p-4 bg-slate-900/50 border-slate-800">
                            <div className="flex items-center justify-between">
                                <div>
                                    <p className="text-xs text-slate-500 font-medium">GCP Assets</p>
                                    <p className="text-2xl font-bold text-green-400 mt-1">
                                        {assets.filter(a => a.provider === 'gcp').length}
                                    </p>
                                </div>
                                <span className="text-3xl">🟢</span>
                            </div>
                        </Card>
                    </div>

                    {/* Assets Grid/List */}
                    {loading ? (
                        <div className="flex items-center justify-center py-24">
                            <div className="text-center">
                                <Cloud className="h-12 w-12 text-brand-400 animate-pulse mx-auto mb-4" />
                                <p className="text-slate-500">Loading cloud assets...</p>
                            </div>
                        </div>
                    ) : assets.length === 0 ? (
                        <Card className="p-12 bg-slate-900/20 border-slate-800 border-dashed">
                            <div className="text-center">
                                <Cloud className="h-16 w-16 text-slate-700 mx-auto mb-4" />
                                <h3 className="text-lg font-semibold text-slate-400 mb-2">No assets found</h3>
                                <p className="text-sm text-slate-600">
                                    Adjust your filters or connect cloud providers to see assets
                                </p>
                            </div>
                        </Card>
                    ) : viewMode === 'graph' ? (
                        <div className="h-[800px]">
                            <AssetGraph
                                assets={assets}
                                onNodeClick={setSelectedAsset}
                                selectedAssetId={selectedAsset?.id}
                            />
                        </div>
                    ) : (
                        <div className={viewMode === 'grid' ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4' : 'space-y-3'}>
                            {assets.map((asset) => (
                                <button
                                    key={asset.id}
                                    onClick={() => setSelectedAsset(asset)}
                                    className={`w-full text-left p-4 bg-slate-900/40 border border-slate-800/50 rounded-xl hover:border-brand-500/30 transition-all group relative overflow-hidden cursor-pointer focus:outline-none focus:ring-2 focus:ring-brand-500/50 ${selectedAsset?.id === asset.id ? 'ring-1 ring-brand-500/50 bg-brand-500/5' : ''
                                        }`}
                                >
                                    {/* Background decoration */}
                                    <div className="absolute top-0 right-0 p-2 opacity-5 group-hover:opacity-10 transition-opacity">
                                        <Server className="h-24 w-24 text-slate-700" />
                                    </div>

                                    <div className="relative z-10">
                                        {/* Header */}
                                        <div className="flex items-start justify-between mb-3">
                                            <div className="flex items-center gap-2">
                                                <span className="text-2xl">{providerIcons[asset.provider]}</span>
                                                <Server className="h-4 w-4 text-slate-400" />
                                            </div>
                                            <Badge variant={getRiskBadgeVariant(asset.risk_score)}>
                                                {getRiskLabel(asset.risk_score)} ({asset.risk_score})
                                            </Badge>
                                        </div>

                                        {/* Asset Info */}
                                        <h3 className="text-sm font-bold text-white mb-1 truncate" title={asset.name}>
                                            {asset.name}
                                        </h3>
                                        <p className="text-xs text-slate-500 font-mono mb-3 truncate" title={asset.id}>
                                            {asset.id}
                                        </p>

                                        {/* Metadata */}
                                        <div className="flex flex-wrap gap-2 text-xs mb-3">
                                            <span className="px-2 py-1 bg-slate-800/50 rounded text-slate-400">
                                                {asset.region}
                                            </span>
                                            <span className={`px-2 py-1 rounded font-medium ${asset.state === 'running'
                                                ? 'bg-emerald-500/10 text-emerald-400'
                                                : 'bg-slate-800/50 text-slate-400'
                                                }`}>
                                                {asset.state}
                                            </span>
                                            {asset.public_ip && (
                                                <span className="px-2 py-1 bg-amber-500/10 text-amber-400 rounded flex items-center gap-1">
                                                    <AlertTriangle className="h-3 w-3" />
                                                    Public IP
                                                </span>
                                            )}
                                        </div>

                                        {/* Tags */}
                                        {Object.keys(asset.tags).length > 0 && (
                                            <div className="flex flex-wrap gap-1">
                                                {Object.entries(asset.tags).slice(0, 2).map(([key, value]) => (
                                                    <span key={key} className="text-[10px] px-1.5 py-0.5 bg-slate-800/30 text-slate-500 rounded">
                                                        {key}: {value}
                                                    </span>
                                                ))}
                                                {Object.keys(asset.tags).length > 2 && (
                                                    <span className="text-[10px] px-1.5 py-0.5 bg-slate-800/30 text-slate-500 rounded">
                                                        +{Object.keys(asset.tags).length - 2} more
                                                    </span>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </button>
                            ))}
                        </div>
                    )}
                </main>
            </div>

            {/* Asset Details Sidebar */}
            {selectedAsset && (
                <div className="w-96 bg-slate-900 border-l border-slate-800 flex flex-col overflow-hidden">
                    <div className="p-4 border-b border-slate-800 flex items-center justify-between">
                        <h2 className="text-sm font-bold text-white flex items-center gap-2">
                            <Shield className="h-4 w-4 text-brand-400" />
                            Asset Details
                        </h2>
                        <button
                            onClick={() => setSelectedAsset(null)}
                            className="text-slate-500 hover:text-white transition-colors"
                        >
                            ✕
                        </button>
                    </div>

                    <div className="flex-1 overflow-y-auto p-4 space-y-4">
                        {/* Provider Badge */}
                        <div className="flex items-center gap-3">
                            <span className="text-4xl">{providerIcons[selectedAsset.provider]}</span>
                            <div>
                                <p className="text-xs text-slate-500">Provider</p>
                                <p className={`text-lg font-bold ${providerColors[selectedAsset.provider]}`}>
                                    {selectedAsset.provider.toUpperCase()}
                                </p>
                            </div>
                        </div>

                        {/* Basic Info */}
                        <div className="space-y-3">
                            <DetailField label="Name" value={selectedAsset.name} />
                            <DetailField label="ID" value={selectedAsset.id} mono />
                            <DetailField label="Type" value={selectedAsset.type} />
                            <DetailField label="Region" value={selectedAsset.region} />
                            <DetailField label="Account" value={selectedAsset.account} mono />
                            <DetailField label="State" value={selectedAsset.state} />
                        </div>

                        {/* Network Info */}
                        {(selectedAsset.private_ip || selectedAsset.public_ip) && (
                            <div className="pt-4 border-t border-slate-800">
                                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Network</h3>
                                <div className="space-y-3">
                                    {selectedAsset.private_ip && <DetailField label="Private IP" value={selectedAsset.private_ip} mono />}
                                    {selectedAsset.public_ip && <DetailField label="Public IP" value={selectedAsset.public_ip} mono />}
                                    {selectedAsset.vpc && <DetailField label="VPC" value={selectedAsset.vpc} mono />}
                                </div>
                            </div>
                        )}

                        {/* Security Groups */}
                        {selectedAsset.security_groups && selectedAsset.security_groups.length > 0 && (
                            <div className="pt-4 border-t border-slate-800">
                                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Security Groups</h3>
                                <div className="space-y-1">
                                    {selectedAsset.security_groups.map((sg) => (
                                        <div key={sg} className="text-xs font-mono bg-slate-950 px-3 py-2 rounded border border-slate-800">
                                            {sg}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {/* Risk Assessment */}
                        <div className="pt-4 border-t border-slate-800">
                            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Risk Assessment</h3>
                            <div className="space-y-3">
                                <div>
                                    <div className="flex items-center justify-between mb-2">
                                        <span className="text-xs text-slate-500">Risk Score</span>
                                        <Badge variant={getRiskBadgeVariant(selectedAsset.risk_score)}>
                                            {selectedAsset.risk_score}/100
                                        </Badge>
                                    </div>
                                    <div className="h-2 bg-slate-950 rounded-full overflow-hidden">
                                        <div
                                            className={`h-full transition-all ${selectedAsset.risk_score >= 70
                                                ? 'bg-red-500'
                                                : selectedAsset.risk_score >= 40
                                                    ? 'bg-amber-500'
                                                    : 'bg-emerald-500'
                                                }`}
                                            style={{ width: `${selectedAsset.risk_score}%` }}
                                        />
                                    </div>
                                </div>

                                {selectedAsset.risk_factors && selectedAsset.risk_factors.length > 0 && (
                                    <div>
                                        <p className="text-xs text-slate-500 mb-2">Risk Factors</p>
                                        <div className="space-y-1">
                                            {selectedAsset.risk_factors.map((factor, idx) => (
                                                <div key={idx} className="flex items-center gap-2 text-xs text-amber-400">
                                                    <AlertTriangle className="h-3 w-3" />
                                                    {factor}
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Tags */}
                        {Object.keys(selectedAsset.tags).length > 0 && (
                            <div className="pt-4 border-t border-slate-800">
                                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">Tags</h3>
                                <div className="space-y-2">
                                    {Object.entries(selectedAsset.tags).map(([key, value]) => (
                                        <div key={key} className="flex items-center justify-between text-xs">
                                            <span className="text-slate-500">{key}</span>
                                            <span className="text-white font-medium">{value}</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

function DetailField({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
    return (
        <div>
            <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-1">{label}</p>
            <p className={`text-sm text-white ${mono ? 'font-mono' : 'font-medium'}`}>{value}</p>
        </div>
    );
}
