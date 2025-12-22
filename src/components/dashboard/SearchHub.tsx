import React, { useState } from 'react';
import { Play, Terminal } from 'lucide-react';
import { Button } from '../ui/Button';
import { SearchService, type EstimateResponse } from '../../services/searchService';
import { CostPreview } from '../search/CostPreview';
import { QueryPilot } from '../search/QueryPilot';

interface SearchHubProps {
    onSearch: (spl: string) => void;
    isLoading: boolean;
    initialSpl?: string;
}

const SEARCH_STORAGE_KEY = 'enterprise_core_search_query';

export const SearchHub: React.FC<SearchHubProps> = ({ onSearch, isLoading, initialSpl }) => {
    const [spl, setSpl] = useState(() => {
        return initialSpl || localStorage.getItem(SEARCH_STORAGE_KEY) || '';
    });
    const [estimate, setEstimate] = useState<EstimateResponse | null>(null);
    const [isEstimating, setIsEstimating] = useState(false);

    React.useEffect(() => {
        if (initialSpl !== undefined) {
            setSpl(initialSpl);
        }
    }, [initialSpl]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (spl.trim()) {
            setIsEstimating(true);
            try {
                // Estimate cost first
                const result = await SearchService.estimateCost(spl);
                setEstimate(result);
            } catch (error) {
                console.error("Failed to estimate query cost:", error);
                alert("Failed to estimate query cost. Please try again.");
            } finally {
                setIsEstimating(false);
            }
        }
    };

    const handleConfirm = () => {
        if (spl.trim()) {
            localStorage.setItem(SEARCH_STORAGE_KEY, spl);
            onSearch(spl);
            setEstimate(null);
        }
    };

    const handleCancel = () => {
        setEstimate(null);
    };

    const handlePilotSelect = (suggestedSpl: string) => {
        setSpl(suggestedSpl);
    };

    return (
        <div className="relative flex flex-col gap-4">
            {/* AI Pilot Section */}
            <QueryPilot onSelectQuery={handlePilotSelect} />

            <div className="bg-slate-900/50 backdrop-blur-xl border border-slate-800 rounded-2xl p-2 shadow-2xl flex items-center gap-2">
                <div className="pl-3 text-slate-500">
                    <Terminal className="w-4 h-4" />
                </div>
                <form onSubmit={handleSubmit} className="flex-1 flex items-center gap-2">
                    <input
                        type="text"
                        value={spl}
                        onChange={(e) => setSpl(e.target.value)}
                        placeholder="Enter SPL query (e.g. search | stats count by sourcetype)"
                        className="flex-1 bg-transparent border-none focus:ring-0 text-slate-200 placeholder:text-slate-600 text-sm font-mono"
                    />
                    <Button
                        type="submit"
                        variant="primary"
                        size="sm"
                        className="h-9 px-4 gap-2 shadow-lg shadow-indigo-500/20"
                        disabled={isLoading || isEstimating || !spl.trim()}
                    >
                        {isLoading || isEstimating ? (
                            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        ) : (
                            <Play className="w-3 h-3 fill-current" />
                        )}
                        Run Query
                    </Button>
                </form>
            </div>

            {/* Cost Preview Overlay */}
            {estimate && (
                <div className="relative">
                    <CostPreview
                        estimate={estimate.estimate}
                        allowed={estimate.allowed}
                        reason={estimate.reason}
                        onConfirm={handleConfirm}
                        onCancel={handleCancel}
                    />
                </div>
            )}
        </div>
    );
};
