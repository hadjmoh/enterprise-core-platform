import React, { useState } from 'react';
import { Sparkles, ArrowRight, AlertTriangle, ShieldCheck } from 'lucide-react';
import { Button } from '../ui/Button';
import { SearchService, type SuggestResponse } from '../../services/searchService';

interface QueryPilotProps {
    onSelectQuery: (spl: string) => void;
}

export const QueryPilot: React.FC<QueryPilotProps> = ({ onSelectQuery }) => {
    const [query, setQuery] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [suggestion, setSuggestion] = useState<SuggestResponse['suggestion'] | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleAskAI = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!query.trim()) return;

        setIsLoading(true);
        setSuggestion(null);
        setError(null);

        try {
            const result = await SearchService.suggestQuery(query);
            if (result.error) {
                setError(result.error);
            } else {
                setSuggestion(result.suggestion);
            }
        } catch (err) {
            setError("Failed to reach AI Pilot service.");
            console.error(err);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="bg-slate-900/80 border border-indigo-500/30 rounded-lg p-4 mb-4">
            <div className="flex items-center gap-2 mb-3">
                <Sparkles className="w-4 h-4 text-indigo-400" />
                <h3 className="text-sm font-semibold text-indigo-100">AI Query Pilot <span className="text-[10px] bg-indigo-500/20 text-indigo-300 px-1.5 py-0.5 rounded ml-2">SAFE MODE</span></h3>
            </div>

            <form onSubmit={handleAskAI} className="flex gap-2">
                <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Ask plain English question (e.g., 'Show me errors from last hour')"
                    className="flex-1 bg-slate-800 border-slate-700 rounded text-sm text-slate-200 placeholder:text-slate-500 focus:ring-1 focus:ring-indigo-500"
                />
                <Button type="submit" variant="primary" size="sm" disabled={isLoading || !query.trim()}>
                    {isLoading ? "Thinking..." : "Generate SPL"}
                </Button>
            </form>

            {error && (
                <div className="mt-3 p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400 flex items-center gap-2">
                    <AlertTriangle className="w-3 h-3" />
                    {error}
                </div>
            )}

            {suggestion && (
                <div className="mt-4 bg-slate-800/50 rounded p-3 border border-slate-700 animate-in fade-in slide-in-from-top-2">
                    <div className="flex justify-between items-start mb-2">
                        <span className="text-xs text-slate-400">Suggested SPL:</span>
                        <div className="flex items-center gap-2">
                            <span className="text-[10px] text-slate-500">Confidence: {(suggestion.confidence_score * 100).toFixed(0)}%</span>
                            {suggestion.allowed ? (
                                <span className="text-[10px] bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded flex items-center gap-1">
                                    <ShieldCheck className="w-3 h-3" /> Safe
                                </span>
                            ) : (
                                <span className="text-[10px] bg-red-500/10 text-red-400 px-1.5 py-0.5 rounded flex items-center gap-1">
                                    <AlertTriangle className="w-3 h-3" /> Blocked
                                </span>
                            )}
                        </div>
                    </div>

                    <div className="bg-black/30 p-2 rounded text-sm font-mono text-emerald-300 mb-2 break-all">
                        {suggestion.suggested_spl}
                    </div>

                    {!suggestion.allowed && (
                        <div className="text-xs text-red-400 mb-2">
                            Blocked: {suggestion.blocking_reason}
                        </div>
                    )}

                    <div className="flex justify-end">
                        <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => onSelectQuery(suggestion.suggested_spl)}
                            disabled={!suggestion.allowed}
                            className={!suggestion.allowed ? "opacity-50 cursor-not-allowed" : ""}
                        >
                            Use Query <ArrowRight className="w-3 h-3 ml-2" />
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
};
