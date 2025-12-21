import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    User,
    Server,
    AlertTriangle,
    Clock,
    Activity,
    Shield,
    ChevronLeft,
    Mail,
    Building,
    Tag
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';
import { identityService } from '../services/identityService';
import type { Entity } from '../services/identityService';
import { cn } from '../lib/utils';

export const EntityProfilePage: React.FC = () => {
    const { entityId } = useParams();
    const navigate = useNavigate();
    const [entity, setEntity] = useState<Entity | null>(null);

    useEffect(() => {
        if (entityId) {
            setEntity(identityService.getEntityById(entityId) || null);
        }
    }, [entityId]);

    if (!entity) {
        return (
            <div className="flex flex-col items-center justify-center p-20 text-slate-500">
                <Shield className="h-16 w-16 mb-4 opacity-20" />
                <h2 className="text-xl font-bold">Entity Not Found</h2>
                <Button variant="ghost" className="mt-4" onClick={() => navigate(-1)}>Go Back</Button>
            </div>
        );
    }

    const isUser = entity.type === 'user';

    return (
        <div className="space-y-6 animate-in fade-in duration-500 p-6 bg-slate-900 min-h-screen text-slate-100">
            {/* Header */}
            <div className="flex items-center gap-4 mb-6">
                <Button variant="ghost" size="sm" onClick={() => navigate(-1)}>
                    <ChevronLeft className="h-5 w-5" />
                </Button>
                <div className={cn(
                    "p-3 rounded-xl",
                    isUser ? "bg-brand-500/20 text-brand-400" : "bg-indigo-500/20 text-indigo-400"
                )}>
                    {isUser ? <User className="h-8 w-8" /> : <Server className="h-8 w-8" />}
                </div>
                <div>
                    <h1 className="text-3xl font-bold text-white">{entity.name}</h1>
                    <div className="flex items-center gap-2 mt-1">
                        <Badge variant="outline" className="text-xs uppercase tracking-widest">{entity.type}</Badge>
                        <span className="text-slate-500 font-mono text-xs">{entity.id}</span>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Left Column: Essential Info & Risk */}
                <div className="space-y-6">
                    <Card className="p-6 border-slate-800 bg-slate-800/40">
                        <div className="flex justify-between items-start mb-4">
                            <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest">Entity Risk Profile</h3>
                            <AlertTriangle className={cn(
                                "h-5 w-5",
                                entity.riskScore > 80 ? "text-red-500" : "text-amber-500"
                            )} />
                        </div>
                        <div className="flex items-end gap-3 mb-4">
                            <span className={cn(
                                "text-5xl font-extrabold",
                                entity.riskScore > 80 ? "text-red-500" : "text-amber-500"
                            )}>{entity.riskScore}</span>
                            <span className="text-slate-600 mb-1">/ 100</span>
                        </div>
                        <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                            <div
                                className={cn(
                                    "h-full rounded-full transition-all duration-1000",
                                    entity.riskScore > 80 ? "bg-red-500" : "bg-amber-500"
                                )}
                                style={{ width: `${entity.riskScore}%` }}
                            />
                        </div>
                        <p className="text-xs text-slate-500 mt-4 leading-relaxed">
                            Risk score calculated based on behavioral anomalies, detection signal severity, and asset criticality.
                        </p>
                    </Card>

                    <Card className="p-6 border-slate-800 bg-slate-800/40">
                        <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-4">Identity Details</h3>
                        <div className="space-y-4">
                            {isUser ? (
                                <>
                                    <div className="flex items-center gap-3">
                                        <Mail className="h-4 w-4 text-slate-500" />
                                        <div className="flex flex-col">
                                            <span className="text-[10px] text-slate-500 uppercase">Email Address</span>
                                            <span className="text-sm text-slate-200">{entity.metadata.email}</span>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <Building className="h-4 w-4 text-slate-500" />
                                        <div className="flex flex-col">
                                            <span className="text-[10px] text-slate-500 uppercase">Department</span>
                                            <span className="text-sm text-slate-200">{entity.metadata.department}</span>
                                        </div>
                                    </div>
                                </>
                            ) : (
                                <>
                                    <div className="flex items-center gap-3">
                                        <Activity className="h-4 w-4 text-slate-500" />
                                        <div className="flex flex-col">
                                            <span className="text-[10px] text-slate-500 uppercase">Operating System</span>
                                            <span className="text-sm text-slate-200">{entity.metadata.os}</span>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <Shield className="h-4 w-4 text-slate-500" />
                                        <div className="flex flex-col">
                                            <span className="text-[10px] text-slate-500 uppercase">Region</span>
                                            <span className="text-sm text-slate-200">{entity.metadata.region}</span>
                                        </div>
                                    </div>
                                </>
                            )}
                            <div className="flex items-center gap-3">
                                <Tag className="h-4 w-4 text-slate-500" />
                                <div className="flex flex-wrap gap-1">
                                    {entity.labels.map(label => (
                                        <Badge key={label} variant="outline" className="text-[10px] bg-slate-900 border-slate-700">{label}</Badge>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </Card>
                </div>

                {/* Right Column: Activity Timeline */}
                <div className="lg:col-span-2">
                    <Card className="p-0 border-slate-800 bg-slate-800/40 flex flex-col h-full overflow-hidden">
                        <div className="p-4 border-b border-slate-800 flex justify-between items-center bg-slate-800/20">
                            <h3 className="font-bold text-white flex items-center gap-2">
                                <Activity className="h-4 w-4 text-brand-400" /> Entity Activity Timeline
                            </h3>
                            <Badge variant="outline">24 Hour window</Badge>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6">
                            <div className="relative space-y-8 before:absolute before:inset-0 before:ml-5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-slate-800 before:via-slate-800 before:to-transparent">
                                {entity.activities.map((act, i) => (
                                    <div key={i} className="relative flex items-start gap-4">
                                        <div className="absolute left-0 mt-1 h-10 w-10 rounded-full border-4 border-slate-900 bg-slate-800 flex items-center justify-center translate-x-[-15%] z-10">
                                            <Clock className="h-4 w-4 text-slate-400" />
                                        </div>
                                        <div className="ml-12 flex-1 pt-1">
                                            <div className="flex items-center justify-between mb-1">
                                                <span className="font-bold text-slate-200 capitalize">{act.action.replaceAll('_', ' ')}</span>
                                                <span className="text-[10px] text-slate-500 font-mono">{new Date(act.timestamp).toLocaleTimeString()}</span>
                                            </div>
                                            <div className="text-sm text-slate-400 flex items-center gap-2">
                                                <span className="text-slate-600">Source:</span>
                                                <span className="text-indigo-400 font-mono text-xs">{act.source}</span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </Card>
                </div>
            </div>
        </div>
    );
};
