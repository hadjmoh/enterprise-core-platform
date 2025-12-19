import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertCircle, RefreshCw } from 'lucide-react';
import { Button } from '../ui/Button';

interface Props {
    children: ReactNode;
    title?: string;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

export class WidgetErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error("Widget Error:", error, errorInfo);
    }

    private handleReset = () => {
        this.setState({ hasError: false, error: null });
    };

    public render() {
        if (this.state.hasError) {
            return (
                <div className="flex flex-col items-center justify-center h-full p-6 text-center space-y-4 bg-red-500/5 rounded-lg border border-red-500/20">
                    <div className="p-3 bg-red-500/10 rounded-full">
                        <AlertCircle className="h-6 w-6 text-red-500" />
                    </div>
                    <div className="space-y-1">
                        <h4 className="text-sm font-bold text-red-400 uppercase tracking-wider">Widget Crash</h4>
                        <p className="text-xs text-slate-500 line-clamp-2 max-w-[200px]">
                            {this.state.error?.message || "An unexpected error occurred in this visualizer."}
                        </p>
                    </div>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={this.handleReset}
                        className="text-xs hover:bg-red-500/10 hover:text-red-400"
                    >
                        <RefreshCw className="h-3 w-3 mr-2" />
                        Reload Widget
                    </Button>
                </div>
            );
        }

        return this.props.children;
    }
}
