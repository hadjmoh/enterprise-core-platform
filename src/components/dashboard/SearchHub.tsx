import React, { useState } from 'react';
import { Play, Terminal } from 'lucide-react';
import { Button } from '../ui/Button';

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

    React.useEffect(() => {
        if (initialSpl !== undefined) {
            setSpl(initialSpl);
        }
    }, [initialSpl]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (spl.trim()) {
            localStorage.setItem(SEARCH_STORAGE_KEY, spl);
            onSearch(spl);
        }
    };

    return (
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
                    disabled={isLoading || !spl.trim()}
                >
                    {isLoading ? (
                        <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : (
                        <Play className="w-3 h-3 fill-current" />
                    )}
                    Run Query
                </Button>
            </form>
        </div>
    );
};
