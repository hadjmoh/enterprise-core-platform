import type { AppWithStatus } from '@/services/appStoreService';

interface AppCardProps {
    app: AppWithStatus;
    onViewDetails: (app: AppWithStatus) => void;
    
}

export const AppCard: React.FC<AppCardProps> = ({ app, onViewDetails }) => {
    return (
        <div
            className="bg-slate-900/50 border border-slate-800 rounded-xl p-6 hover:border-brand-500/30 transition-all cursor-pointer group"
            onClick={() => onViewDetails(app)}
        >
            {/* Icon and Title */}
            <div className="flex items-start gap-4 mb-4">
                <div className="text-4xl flex-shrink-0">
                    {app.icon || '📦'}
                </div>
                <div className="flex-1 min-w-0">
                    <h3 className="text-lg font-semibold text-slate-100 group-hover:text-brand-400 transition-colors truncate">
                        {app.title}
                    </h3>
                    <p className="text-sm text-slate-400">by {app.author}</p>
                </div>
                {app.installed && (
                    <span className="px-2 py-1 bg-green-500/20 text-green-400 text-xs font-medium rounded">
                        Installed
                    </span>
                )}
            </div>

            {/* Description */}
            <p className="text-sm text-slate-300 mb-4 line-clamp-2">
                {app.description}
            </p>

            {/* Metadata */}
            <div className="flex items-center justify-between text-xs text-slate-500">
                <div className="flex items-center gap-4">
                    {app.rating && (
                        <span className="flex items-center gap-1">
                            ⭐ {app.rating.toFixed(1)}
                        </span>
                    )}
                    {app.downloads && (
                        <span>
                            {app.downloads.toLocaleString()} downloads
                        </span>
                    )}
                </div>
                <span className="px-2 py-1 bg-slate-800 rounded text-slate-400">
                    {app.category}
                </span>
            </div>

            {/* Tags */}
            {app.tags && app.tags.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-3">
                    {app.tags.slice(0, 3).map((tag) => (
                        <span
                            key={tag}
                            className="px-2 py-0.5 bg-slate-800/50 text-slate-400 text-xs rounded"
                        >
                            {tag}
                        </span>
                    ))}
                </div>
            )}
        </div>
    );
};
