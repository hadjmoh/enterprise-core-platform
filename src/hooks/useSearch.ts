import { useState, useEffect, useCallback, useRef } from 'react';

export interface SearchMetadata {
    _type: string;
    _span?: number;
    _generated_at: string;
    _pipeline_id: number;
}

export interface SearchEvent {
    _total_count: number;
    batch: Record<string, unknown>[];
}

export interface SearchError {
    code: string;
    message: string;
    fatal: boolean;
}

export interface SearchProgress {
    status: 'idle' | 'loading' | 'streaming' | 'complete' | 'error';
    metadata: SearchMetadata | null;
    results: Record<string, unknown>[];
    error: SearchError | null;
    totalCount: number;
}

export function useSearch() {
    const [progress, setProgress] = useState<SearchProgress>({
        status: 'idle',
        metadata: null,
        results: [],
        error: null,
        totalCount: 0,
    });

    const eventSourceRef = useRef<EventSource | null>(null);

    const execute = useCallback((spl: string) => {
        // Cleanup previous connection
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        setProgress({
            status: 'loading',
            metadata: null,
            results: [],
            error: null,
            totalCount: 0,
        });

        const encodedSpl = encodeURIComponent(spl);
        const url = `http://localhost:8080/services/search?q=${encodedSpl}`;

        const es = new EventSource(url);
        eventSourceRef.current = es;

        es.addEventListener('metadata', (e) => {
            const meta = JSON.parse(e.data) as SearchMetadata;
            setProgress(prev => ({ ...prev, status: 'streaming', metadata: meta }));
        });

        es.addEventListener('event', (e) => {
            const event = JSON.parse(e.data) as SearchEvent;
            setProgress(prev => ({
                ...prev,
                results: [...prev.results, ...event.batch],
                totalCount: event._total_count,
            }));
        });

        es.addEventListener('error', (e) => {
            const eventData = (e as MessageEvent).data;
            const errorData = JSON.parse(eventData || '{}') as SearchError;
            setProgress(prev => ({
                ...prev,
                status: 'error',
                error: {
                    code: errorData.code || 'UNKNOWN',
                    message: errorData.message || 'Connection error',
                    fatal: errorData.fatal ?? true,
                }
            }));
            if (errorData.fatal !== false) {
                es.close();
            }
        });

        es.addEventListener('done', () => {
            setProgress(prev => ({ ...prev, status: 'complete' }));
            es.close();
        });

        return () => es.close();
    }, []);

    useEffect(() => {
        return () => {
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
            }
        };
    }, []);

    return { ...progress, execute };
}
