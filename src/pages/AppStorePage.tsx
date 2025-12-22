import { useState, useEffect } from 'react';
import { Search, Filter } from 'lucide-react';
import type { AppWithStatus } from '@/services/appStoreService';
import { appStoreService } from '@/services/appStoreService';
import { AppCard } from '@/components/apps/AppCard';
import { AppDetailModal } from '@/components/apps/AppDetailModal';

const CATEGORIES = ['All', 'Security', 'Analytics', 'Integration', 'Compliance'];

const AppStorePage = () => {
    const [apps, setApps] = useState<AppWithStatus[]>([]);
    const [filteredApps, setFilteredApps] = useState<AppWithStatus[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedCategory, setSelectedCategory] = useState('All');
    const [selectedApp, setSelectedApp] = useState<AppWithStatus | null>(null);

    const fetchApps = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await appStoreService.fetchAvailableApps();
            setApps(data);
            setFilteredApps(data);
        } catch (err: any) {
            setError(err.message || 'Failed to load apps');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchApps();
    }, []);

    useEffect(() => {
        let result = apps;

        // Filter by category
        if (selectedCategory !== 'All') {
            result = result.filter(app => app.category === selectedCategory);
        }

        // Filter by search query
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            result = result.filter(app =>
                app.title.toLowerCase().includes(query) ||
                app.description.toLowerCase().includes(query) ||
                app.tags?.some(tag => tag.toLowerCase().includes(query))
            );
        }

        setFilteredApps(result);
    }, [apps, selectedCategory, searchQuery]);

    const handleSearch = async (query: string) => {
        setSearchQuery(query);
        if (query.trim()) {
            try {
                const results = await appStoreService.searchApps(query);
                setFilteredApps(results);
            } catch (err) {
                // Fallback to client-side filtering
                console.error('Search failed, using client-side filter', err);
            }
        }
    };

    return (
        <div className="p-8 space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold text-slate-100 mb-2">App Store</h1>
                <p className="text-slate-400">
                    Discover and install apps to extend your platform capabilities
                </p>
            </div>

            {/* Search and Filters */}
            <div className="flex flex-col md:flex-row gap-4">
                {/* Search Bar */}
                <div className="flex-1 relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={20} />
                    <input
                        type="text"
                        placeholder="Search apps..."
                        value={searchQuery}
                        onChange={(e) => handleSearch(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 bg-slate-900 border border-slate-800 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:border-brand-500"
                    />
                </div>

                {/* Category Filter */}
                <div className="flex items-center gap-2">
                    <Filter size={20} className="text-slate-400" />
                    <select
                        value={selectedCategory}
                        onChange={(e) => setSelectedCategory(e.target.value)}
                        className="px-4 py-2 bg-slate-900 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-brand-500"
                    >
                        {CATEGORIES.map(cat => (
                            <option key={cat} value={cat}>{cat}</option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Stats */}
            <div className="flex items-center gap-6 text-sm text-slate-400">
                <span>{filteredApps.length} apps available</span>
                <span>•</span>
                <span>{apps.filter(a => a.installed).length} installed</span>
            </div>

            {/* Loading State */}
            {loading && (
                <div className="flex items-center justify-center py-20">
                    <div className="text-slate-400">Loading apps...</div>
                </div>
            )}

            {/* Error State */}
            {error && (
                <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400">
                    {error}
                </div>
            )}

            {/* Apps Grid */}
            {!loading && !error && (
                <>
                    {filteredApps.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-20 text-slate-400">
                            <p className="text-lg mb-2">No apps found</p>
                            <p className="text-sm">Try adjusting your search or filters</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {filteredApps.map(app => (
                                <AppCard
                                    key={app.name}
                                    app={app}
                                    onViewDetails={setSelectedApp}
                                />
                            ))}
                        </div>
                    )}
                </>
            )}

            {/* Detail Modal */}
            <AppDetailModal
                app={selectedApp}
                onClose={() => setSelectedApp(null)}
                onInstallSuccess={fetchApps}
            />
        </div>
    );
};

export default AppStorePage;
