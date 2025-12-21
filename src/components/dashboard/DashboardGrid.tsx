import { ResponsiveGridLayout, useContainerWidth } from 'react-grid-layout';
import type { Layout, LayoutItem } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import type { Widget } from '../../hooks/useDashboard';
import { WidgetContainer } from './WidgetContainer';

interface DashboardGridProps {
    widgets: Widget[];
    isLocked: boolean;
    onLayoutChange: (layouts: LayoutItem[]) => void;
    onRemoveWidget: (id: string) => void;
    onGlobalSearch?: (spl: string) => void;
    onAddWidget?: (widget: { title: string, type: string, spl: string }) => void;
}

export function DashboardGrid({ widgets, isLocked, onLayoutChange, onRemoveWidget, onGlobalSearch, onAddWidget }: DashboardGridProps) {
    const { width, containerRef, mounted } = useContainerWidth({
        measureBeforeMount: false
    });

    const layoutConfig = {
        lg: widgets.map(w => w.layout)
    };

    return (
        <div ref={containerRef} className="w-full min-h-[500px]">
            {mounted && (
                <ResponsiveGridLayout
                    className="layout"
                    layouts={layoutConfig}
                    width={width}
                    breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
                    cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
                    rowHeight={100}
                    dragConfig={{
                        enabled: !isLocked,
                        handle: ".drag-handle"
                    }}
                    resizeConfig={{
                        enabled: !isLocked
                    }}
                    onLayoutChange={(current: Layout) => onLayoutChange(current as LayoutItem[])}
                    margin={[16, 16]}
                >
                    {widgets.map((widget) => (
                        <div key={widget.id}>
                            <WidgetContainer
                                id={widget.id}
                                title={widget.title}
                                spl={widget.spl}
                                isLocked={isLocked}
                                onRemove={onRemoveWidget}
                                onGlobalSearch={onGlobalSearch}
                                onAddWidget={onAddWidget}
                            />
                        </div>
                    ))}
                </ResponsiveGridLayout>
            )}
        </div>
    );
}
