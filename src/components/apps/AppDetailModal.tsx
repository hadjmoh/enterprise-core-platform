import { useState } from 'react';
import type { AppWithStatus } from '@/services/appStoreService';
import { appStoreService } from '@/services/appStoreService';
import { X, Download, Trash2, ExternalLink } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface AppDetailModalProps {
    app: AppWithStatus | null;
    onClose: () => void;
    onInstallSuccess?: () => void;
}

export const AppDetailModal: React.FC<AppDetailModalProps> = ({ app, onClose, onInstallSuccess }) => {
    const [installing, setInstalling] = useState(false);
    const [uninstalling, setUninstalling] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();

    if (!app) return null;

    const handleInstall = async () => {
        // Trigger file input
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.tar.gz,.tgz';
        input.onchange = async (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (!file) return;

            setInstalling(true);
            setError(null);
            setUploadProgress(0);

            try {
                await appStoreService.installApp(file, (progress) => {
                    setUploadProgress(progress);
                });
                onInstallSuccess?.();
                onClose();
            } catch (err: any) {
                setError(err.message || 'Installation failed');
            } finally {
                setInstalling(false);
                setUploadProgress(0);
            }
        };
        input.click();
    };

    const handleUninstall = async () => {
        if (!confirm(`Are you sure you want to uninstall ${app.title}?`)) return;

        setUninstalling(true);
        setError(null);

        try {
            await appStoreService.uninstallApp(app.name);
            onInstallSuccess?.();
            onClose();
        } catch (err: any) {
            setError(err.message || 'Uninstall failed');
        } finally {
            setUninstalling(false);
        }
    };

    const handleOpenDashboard = () => {
        navigate(`/apps/${app.name}/dashboards/main`);
        onClose();
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
            <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
                {/* Header */}
                <div className="flex items-start justify-between p-6 border-b border-slate-800">
                    <div className="flex items-start gap-4">
                        <div className="text-5xl">{app.icon || '📦'}</div>
                        <div>
                            <h2 className="text-2xl font-bold text-slate-100">{app.title}</h2>
                            <p className="text-slate-400">by {app.author} • v{app.version}</p>
                            <div className="flex items-center gap-3 mt-2">
                                {app.rating && (
                                    <span className="text-sm text-slate-400">
                                        ⭐ {app.rating.toFixed(1)}
                                    </span>
                                )}
                                {app.downloads && (
                                    <span className="text-sm text-slate-400">
                                        {app.downloads.toLocaleString()} downloads
                                    </span>
                                )}
                                <span className="px-2 py-1 bg-slate-800 text-slate-300 text-xs rounded">
                                    {app.category}
                                </span>
                            </div>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-slate-400 hover:text-slate-100 transition-colors"
                    >
                        <X size={24} />
                    </button>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6">
                    {/* Description */}
                    <div>
                        <h3 className="text-lg font-semibold text-slate-100 mb-2">Description</h3>
                        <p className="text-slate-300 leading-relaxed">
                            {app.longDescription || app.description}
                        </p>
                    </div>

                    {/* Tags */}
                    {app.tags && app.tags.length > 0 && (
                        <div>
                            <h3 className="text-lg font-semibold text-slate-100 mb-2">Tags</h3>
                            <div className="flex flex-wrap gap-2">
                                {app.tags.map((tag) => (
                                    <span
                                        key={tag}
                                        className="px-3 py-1 bg-slate-800 text-slate-300 text-sm rounded-full"
                                    >
                                        {tag}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Permissions */}
                    {app.permissions && app.permissions.length > 0 && (
                        <div>
                            <h3 className="text-lg font-semibold text-slate-100 mb-2">Permissions Required</h3>
                            <ul className="space-y-1">
                                {app.permissions.map((perm) => (
                                    <li key={perm} className="text-slate-300 text-sm flex items-center gap-2">
                                        <span className="w-1.5 h-1.5 bg-brand-500 rounded-full"></span>
                                        {perm}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}

                    {/* Error Message */}
                    {error && (
                        <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
                            {error}
                        </div>
                    )}

                    {/* Upload Progress */}
                    {installing && uploadProgress > 0 && (
                        <div>
                            <div className="flex justify-between text-sm text-slate-400 mb-2">
                                <span>Uploading...</span>
                                <span>{Math.round(uploadProgress)}%</span>
                            </div>
                            <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-brand-500 transition-all duration-300"
                                    style={{ width: `${uploadProgress}%` }}
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="p-6 border-t border-slate-800 flex items-center justify-between">
                    <div className="flex gap-3">
                        {app.installed ? (
                            <>
                                <button
                                    onClick={handleOpenDashboard}
                                    className="px-4 py-2 bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors flex items-center gap-2"
                                >
                                    <ExternalLink size={18} />
                                    Open Dashboard
                                </button>
                                <button
                                    onClick={handleUninstall}
                                    disabled={uninstalling}
                                    className="px-4 py-2 bg-red-500/20 hover:bg-red-500/30 text-red-400 rounded-lg transition-colors flex items-center gap-2 disabled:opacity-50"
                                >
                                    <Trash2 size={18} />
                                    {uninstalling ? 'Uninstalling...' : 'Uninstall'}
                                </button>
                            </>
                        ) : (
                            <button
                                onClick={handleInstall}
                                disabled={installing}
                                className="px-6 py-2 bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors flex items-center gap-2 disabled:opacity-50"
                            >
                                <Download size={18} />
                                {installing ? 'Installing...' : 'Install'}
                            </button>
                        )}
                    </div>
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-slate-400 hover:text-slate-100 transition-colors"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
};
