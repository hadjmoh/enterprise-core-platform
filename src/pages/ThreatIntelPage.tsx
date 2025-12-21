import React, { useState, useEffect } from 'react';
import {
    Shield,
    Database,
    Search,
    RefreshCw,
    AlertTriangle,
    CheckCircle,
    Activity,
    Globe,
    FileText,
    Server
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Badge } from '../components/ui/Badge';
import { threatIntelService } from '../services/threatIntelService';
import type { TiFeed, TiIndicator } from '../services/threatIntelService';

export const ThreatIntelPage: React.FC = () => {
    const [feeds, setFeeds] = useState<TiFeed[]>([]);
    const [stats, setStats] = useState({ total: 0, critical: 0, high: 0, types: { ip: 0, domain: 0, hash: 0 } });
    const [lookupQuery, setLookupQuery] = useState('');
    const [lookupResult, setLookupResult] = useState<TiIndicator | null>(null);
    const [isSearching, setIsSearching] = useState(false);
    const [activeTab, setActiveTab] = useState<'overview' | 'browse' | 'lookup'>('overview');
    const [recentIndicators, setRecentIndicators] = useState<TiIndicator[]>([]);

    useEffect(() => {
        refreshData();
        const interval = setInterval(refreshData, 5000);
        return () => clearInterval(interval);
    }, []);

    const refreshData = () => {
        setFeeds(threatIntelService.getFeeds());
        setStats(threatIntelService.getStats());
        setRecentIndicators(threatIntelService.getRecentIndicators(20));
    };

    const handleSync = async (feedId: string) => {
        try {
            // Force UI update immediately to show 'syncing' state
            setFeeds(feeds.map(f => f.id === feedId ? { ...f, status: 'syncing' } : f));
            await threatIntelService.ingestFeed(feedId);
            refreshData();
        } catch (e) {
            console.error(e);
        }
    };

    const handleLookup = () => {
        if (!lookupQuery) return;
        setIsSearching(true);
        // Simulate slight delay for realism
        setTimeout(() => {
            const res = threatIntelService.manualSearch(lookupQuery);
            setLookupResult(res.length > 0 ? res[0] : null); // Just take first match for this demo
            setIsSearching(false);
        }, 600);
    };

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-slate-100 font-sans">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-emerald-400 to-cyan-400 bg-clip-text text-transparent flex items-center gap-2">
                        <Shield className="h-8 w-8 text-emerald-400" />
                        Threat Intelligence Center
                    </h1>
                    <p className="text-slate-400 mt-1">Manage external feeds and hunt for indicators.</p>
                </div>
                <div className="flex gap-2">
                    <Button
                        variant={activeTab === 'overview' ? 'primary' : 'ghost'}
                        onClick={() => setActiveTab('overview')}
                        size="sm"
                    >
                        Overview
                    </Button>
                    <Button
                        variant={activeTab === 'lookup' ? 'primary' : 'ghost'}
                        onClick={() => setActiveTab('lookup')}
                        size="sm"
                    >
                        Indicator Lookup
                    </Button>
                </div>
            </div>

            {activeTab === 'overview' && (
                <>
                    {/* KPI Cards */}
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <Card className="p-4 border-slate-700 bg-slate-800/50 backdrop-blur">
                            <div className="flex justify-between items-start">
                                <div>
                                    <p className="text-slate-400 text-sm font-medium">Total Indicators</p>
                                    <h3 className="text-3xl font-bold text-white mt-2">{stats.total.toLocaleString()}</h3>
                                </div>
                                <Database className="h-5 w-5 text-blue-400" />
                            </div>
                        </Card>
                        <Card className="p-4 border-slate-700 bg-slate-800/50 backdrop-blur">
                            <div className="flex justify-between items-start">
                                <div>
                                    <p className="text-slate-400 text-sm font-medium">Critical Threats</p>
                                    <h3 className="text-3xl font-bold text-red-500 mt-2">{stats.critical.toLocaleString()}</h3>
                                </div>
                                <AlertTriangle className="h-5 w-5 text-red-500" />
                            </div>
                        </Card>
                        <Card className="p-4 border-slate-700 bg-slate-800/50 backdrop-blur">
                            <div className="flex justify-between items-start">
                                <div>
                                    <p className="text-slate-400 text-sm font-medium">High Severity</p>
                                    <h3 className="text-3xl font-bold text-orange-400 mt-2">{stats.high.toLocaleString()}</h3>
                                </div>
                                <Activity className="h-5 w-5 text-orange-400" />
                            </div>
                        </Card>
                        <Card className="p-4 border-slate-700 bg-slate-800/50 backdrop-blur">
                            <div className="flex justify-between items-start">
                                <div>
                                    <p className="text-slate-400 text-sm font-medium">Active Feeds</p>
                                    <h3 className="text-3xl font-bold text-emerald-400 mt-2">
                                        {feeds.filter(f => f.status === 'active').length}
                                        <span className="text-slate-500 text-lg font-normal ml-1">/ {feeds.length}</span>
                                    </h3>
                                </div>
                                <RefreshCw className="h-5 w-5 text-emerald-400" />
                            </div>
                        </Card>
                    </div>

                    {/* Feed Management */}
                    <h2 className="text-lg font-semibold text-white mt-8 mb-4 flex items-center gap-2">
                        <Server className="h-5 w-5 text-slate-400" />
                        Feed Connectors (STIX/TAXII)
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {feeds.map(feed => (
                            <Card key={feed.id} className="border-slate-700 bg-slate-800/30 flex flex-col justify-between">
                                <div className="p-5">
                                    <div className="flex justify-between items-start mb-2">
                                        <Badge variant={feed.status === 'active' ? 'success' : feed.status === 'syncing' ? 'warning' : 'outline'}>
                                            {feed.status.toUpperCase()}
                                        </Badge>
                                        {feed.status === 'active' && <CheckCircle className="h-4 w-4 text-emerald-500" />}
                                    </div>
                                    <h3 className="text-lg font-bold text-white mb-1">{feed.name}</h3>
                                    <p className="text-xs text-slate-400 mb-4">{feed.description}</p>

                                    <div className="space-y-2 text-sm">
                                        <div className="flex justify-between text-slate-300">
                                            <span>Indicators:</span>
                                            <span className="font-mono">{feed.indicatorCount.toLocaleString()}</span>
                                        </div>
                                        <div className="flex justify-between text-slate-300">
                                            <span>Last Sync:</span>
                                            <span>{feed.lastSync ? new Date(feed.lastSync).toLocaleTimeString() : 'Never'}</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="p-4 bg-slate-800/50 border-t border-slate-700">
                                    <Button
                                        className="w-full"
                                        variant="secondary"
                                        onClick={() => handleSync(feed.id)}
                                        disabled={feed.status === 'syncing'}
                                    >
                                        {feed.status === 'syncing' ? (
                                            <><RefreshCw className="h-4 w-4 mr-2 animate-spin" /> Syncing...</>
                                        ) : (
                                            <><RefreshCw className="h-4 w-4 mr-2" /> Force Sync</>
                                        )}
                                    </Button>
                                </div>
                            </Card>
                        ))}
                    </div>

                    {/* Recent Indicators Table */}
                    <h2 className="text-lg font-semibold text-white mt-8 mb-4 flex items-center gap-2">
                        <Activity className="h-5 w-5 text-slate-400" />
                        Recent Ingestions
                    </h2>
                    <div className="rounded-lg border border-slate-700 overflow-hidden bg-slate-800/20">
                        <table className="w-full text-sm text-left">
                            <thead className="bg-slate-800 text-slate-400 font-medium">
                                <tr>
                                    <th className="p-3">Severity</th>
                                    <th className="p-3">Type</th>
                                    <th className="p-3">Pattern / Value</th>
                                    <th className="p-3">Source</th>
                                    <th className="p-3">Time</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-700/50">
                                {recentIndicators.map(ind => (
                                    <tr key={ind.id} className="hover:bg-slate-800/30 transition-colors">
                                        <td className="p-3">
                                            <Badge variant={
                                                ind.severity === 'critical' ? 'error' :
                                                    ind.severity === 'high' ? 'warning' : 'default'
                                            }>
                                                {ind.severity}
                                            </Badge>
                                        </td>
                                        <td className="p-3 text-slate-300">
                                            {ind.pattern.includes('ipv4') ? <span className="flex items-center gap-1"><Globe className="h-3 w-3" /> IP</span> :
                                                ind.pattern.includes('domain') ? <span className="flex items-center gap-1"><Globe className="h-3 w-3" /> Domain</span> :
                                                    <span className="flex items-center gap-1"><FileText className="h-3 w-3" /> Hash</span>}
                                        </td>
                                        <td className="p-3 font-mono text-slate-200">{ind.pattern.replaceAll(/\[\w+:value = '|'\]/g, '')}</td>
                                        <td className="p-3 text-slate-400">{ind.source}</td>
                                        <td className="p-3 text-slate-500">{new Date(ind.created).toLocaleTimeString()}</td>
                                    </tr>
                                ))}
                                {recentIndicators.length === 0 && (
                                    <tr>
                                        <td colSpan={5} className="p-8 text-center text-slate-500 italic">No indicators yet. Sync a feed above.</td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </>
            )}

            {activeTab === 'lookup' && (
                <div className="max-w-2xl mx-auto mt-12">
                    <Card className="p-8 border-slate-600 bg-slate-800/80 shadow-2xl">
                        <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
                            <Search className="h-6 w-6 text-brand-400" />
                            Global Indicator Lookup
                        </h2>
                        <div className="flex gap-4 mb-8">
                            <Input
                                placeholder="Enter IP, Domain, or Hash..."
                                className="flex-1 bg-slate-900 border-slate-600 focus:border-brand-500 h-12 text-lg"
                                value={lookupQuery}
                                onChange={(e) => setLookupQuery(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && handleLookup()}
                            />
                            <Button size="lg" onClick={handleLookup} disabled={isSearching || !lookupQuery} className="h-12 px-8">
                                {isSearching ? <RefreshCw className="animate-spin" /> : 'Scan'}
                            </Button>
                        </div>

                        {isSearching && (
                            <div className="text-center py-12">
                                <div className="animate-pulse flex flex-col items-center">
                                    <div className="h-4 w-48 bg-slate-700 rounded mb-2"></div>
                                    <div className="h-3 w-32 bg-slate-700/50 rounded"></div>
                                </div>
                            </div>
                        )}

                        {!isSearching && lookupResult && (
                            <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
                                <div className={`p-4 rounded-lg border flex items-start gap-4 ${lookupResult.severity === 'critical' ? 'bg-red-500/10 border-red-500/50' :
                                    lookupResult.severity === 'high' ? 'bg-orange-500/10 border-orange-500/50' :
                                        'bg-slate-700/50 border-slate-600'
                                    }`}>
                                    <AlertTriangle className={`h-8 w-8 shrink-0 ${lookupResult.severity === 'critical' ? 'text-red-500' : 'text-orange-500'
                                        }`} />
                                    <div>
                                        <h3 className="text-lg font-bold text-white mb-1">Threat Detected: {lookupResult.severity.toUpperCase()}</h3>
                                        <p className="text-slate-300 font-mono mb-2">{lookupResult.pattern}</p>
                                        <div className="flex gap-2 flex-wrap text-xs">
                                            <Badge variant="outline">Confidence: {lookupResult.confidence}%</Badge>
                                            <Badge variant="outline">Source: {lookupResult.source}</Badge>
                                            {lookupResult.labels.map(l => (
                                                <Badge key={l} className="bg-slate-700">{l}</Badge>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}

                        {!isSearching && !lookupResult && lookupQuery && (
                            <div className="text-center py-8 text-slate-400">
                                <CheckCircle className="h-12 w-12 text-slate-600 mx-auto mb-3" />
                                <p>No active threats found for this indicator.</p>
                                <p className="text-xs text-slate-500 mt-1">(Check query syntax or sync more feeds)</p>
                            </div>
                        )}
                    </Card>
                </div>
            )}
        </div>
    );
};
