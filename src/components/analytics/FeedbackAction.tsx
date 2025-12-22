import React, { useState } from 'react';
import { Button } from '../ui/Button';
import { ThumbsUp, ThumbsDown, MessageSquare, AlertTriangle } from 'lucide-react';
import { SearchService } from '../../services/searchService';

interface FeedbackActionProps {
    entityId: string;
    alertId?: string;
    onFeedbackSubmitted?: () => void;
}

export const FeedbackAction: React.FC<FeedbackActionProps> = ({ entityId, alertId, onFeedbackSubmitted }) => {
    const [rating, setRating] = useState<number | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [showComment, setShowComment] = useState(false);
    const [comment, setComment] = useState('');
    const [isFalsePositive, setIsFalsePositive] = useState(false);

    const handleSubmit = async (submitRating: number) => {
        setIsSubmitting(true);
        try {
            await fetch('/api/feedback', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    entity_id: entityId,
                    alert_id: alertId || 'manual',
                    rating: submitRating,
                    is_false_positive: isFalsePositive,
                    comments: comment
                })
            });
            setRating(submitRating);
            if (onFeedbackSubmitted) onFeedbackSubmitted();
        } catch (error) {
            console.error("Failed to submit feedback", error);
            alert("Failed to submit feedback. Please try again.");
        } finally {
            setIsSubmitting(false);
        }
    };

    if (rating !== null) {
        return (
            <div className="text-xs text-green-400 flex items-center gap-1 animate-in fade-in">
                <span>Thanks for your feedback!</span>
            </div>
        );
    }

    return (
        <div className="flex items-center gap-1">
            <div className={`flex items-center gap-1 transition-all ${showComment ? 'opacity-50 pointer-events-none' : ''}`}>
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleSubmit(5)}
                    className="h-6 w-6 p-0 hover:text-green-400"
                    title="Helpful / Accurate"
                    disabled={isSubmitting}
                >
                    <ThumbsUp className="w-3 h-3" />
                </Button>

                <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowComment(true)}
                    className="h-6 w-6 p-0 hover:text-red-400"
                    title="Inaccurate / False Positive"
                    disabled={isSubmitting}
                >
                    <ThumbsDown className="w-3 h-3" />
                </Button>
            </div>

            {showComment && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
                    <div className="bg-slate-900 border border-slate-700 rounded-lg p-4 w-full max-w-sm shadow-xl">
                        <h4 className="text-sm font-semibold text-slate-200 mb-3">Provide Feedback</h4>

                        <label className="flex items-center gap-2 mb-3 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={isFalsePositive}
                                onChange={(e) => setIsFalsePositive(e.target.checked)}
                                className="rounded bg-slate-800 border-slate-600 text-indigo-500 focus:ring-0"
                            />
                            <span className="text-sm text-slate-300">This is a False Positive</span>
                        </label>

                        <textarea
                            value={comment}
                            onChange={(e) => setComment(e.target.value)}
                            placeholder="Why was this analysis incorrect?"
                            className="w-full h-24 bg-slate-800 border-slate-700 rounded text-sm text-slate-200 p-2 mb-4 focus:ring-1 focus:ring-indigo-500"
                        />

                        <div className="flex justify-end gap-2">
                            <Button variant="ghost" size="sm" onClick={() => setShowComment(false)}>
                                Cancel
                            </Button>
                            <Button variant="primary" size="sm" onClick={() => handleSubmit(1)} disabled={!comment.trim() && !isFalsePositive}>
                                Submit Feedback
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
