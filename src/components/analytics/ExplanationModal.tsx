import React from 'react';
import { Button } from '../ui/Button';
import { X, Sparkles, AlertTriangle, HelpCircle } from 'lucide-react';

interface Factor {
    name: string;
    weight: number;
    value: number;
    description: string;
}

interface Explanation {
    entity_id: string;
    anomaly_type: string;
    factors: Factor[];
    confidence: number;
    description: string;
}

interface ExplanationModalProps {
    explanation: Explanation;
    onClose: () => void;
}

export const ExplanationModal: React.FC<ExplanationModalProps> = ({ explanation, onClose }) => {
    return (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
            <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl max-w-2xl w-full flex flex-col max-h-[90vh]">
                <div className="flex items-center justify-between p-4 border-b border-slate-800">
                    <div className="flex items-center gap-2">
                        <Sparkles className="w-5 h-5 text-indigo-400" />
                        <h3 className="text-lg font-semibold text-slate-100">Anomaly Explanation</h3>
                    </div>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        <X className="w-5 h-5" />
                    </Button>
                </div>

                <div className="p-6 overflow-y-auto">
                    <div className="mb-6">
                        <div className="flex items-center gap-2 mb-2">
                            <span className="text-sm font-medium text-slate-400">Analysis for:</span>
                            <code className="text-sm bg-slate-800 text-indigo-300 px-2 py-0.5 rounded">{explanation.entity_id}</code>
                        </div>
                        <p className="text-lg text-slate-200">{explanation.description}</p>
                        <div className="flex items-center gap-2 mt-2">
                            <span className="text-xs text-slate-500">Confidence:</span>
                            <div className="h-1.5 w-24 bg-slate-800 rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-indigo-500 rounded-full"
                                    style={{ width: `${explanation.confidence * 100}%` }}
                                />
                            </div>
                            <span className="text-xs text-slate-400">{(explanation.confidence * 100).toFixed(0)}%</span>
                        </div>
                    </div>

                    <h4 className="text-sm font-semibold text-slate-300 mb-4 uppercase tracking-wider">Contributing Factors</h4>

                    <div className="space-y-4">
                        {explanation.factors.map((factor, idx) => (
                            <div key={idx} className="bg-slate-800/50 rounded-lg p-4 border border-slate-800">
                                <div className="flex justify-between items-start mb-2">
                                    <div>
                                        <h5 className="font-medium text-slate-200">{factor.name}</h5>
                                        <p className="text-sm text-slate-400 mt-1">{factor.description}</p>
                                    </div>
                                    <div className="text-right">
                                        <span className="text-xs font-mono text-indigo-300 bg-indigo-500/10 px-2 py-1 rounded">
                                            Weight: {(factor.weight * 100).toFixed(0)}%
                                        </span>
                                    </div>
                                </div>
                                <div className="mt-3">
                                    <div className="flex justify-between text-xs text-slate-500 mb-1">
                                        <span>Contribution</span>
                                        <span>High Impact</span>
                                    </div>
                                    <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
                                        <div
                                            className="h-full bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full"
                                            style={{ width: `${factor.weight * 100}%` }}
                                        />
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="p-4 border-t border-slate-800 bg-slate-900/50 rounded-b-xl flex justify-end">
                    <Button variant="secondary" onClick={onClose}>
                        Close
                    </Button>
                </div>
            </div>
        </div>
    );
};
