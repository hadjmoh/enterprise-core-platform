// Web Worker for data normalization to avoid blocking the main thread
self.onmessage = (e: MessageEvent) => {
    const { data, keys } = e.data;

    if (!data || !keys) {
        self.postMessage({ datasets: [] });
        return;
    }

    const datasets = keys.map((key: string) => ({
        label: key,
        data: data.map((d: Record<string, unknown>) => ({
            x: new Date(d._time as string).getTime(),
            y: d[key]
        })),
    }));

    self.postMessage({ datasets });
};
