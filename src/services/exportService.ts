/**
 * Export Service
 * Handles data conversion and browser-driven downloads for CSV and PDF formats.
 */

const REPORT_HISTORY_KEY = 'enterprise_core_report_history';

export interface ReportRecord {
    id: string;
    name: string;
    type: 'CSV' | 'PDF' | 'JSON';
    timestamp: string;
    size: string;
    status: 'Ready' | 'Expired';
}

const addReportToHistory = (name: string, type: ReportRecord['type'], sizeInBytes: number) => {
    const saved = localStorage.getItem(REPORT_HISTORY_KEY);
    const history: ReportRecord[] = saved ? JSON.parse(saved) : [];

    const sizeStr = sizeInBytes > 1024 * 1024
        ? `${(sizeInBytes / (1024 * 1024)).toFixed(1)} MB`
        : `${(sizeInBytes / 1024).toFixed(0)} KB`;

    const newReport: ReportRecord = {
        id: Math.random().toString(36).substring(2, 9),
        name,
        type,
        timestamp: new Date().toISOString().replace('T', ' ').substring(0, 16),
        size: sizeStr,
        status: 'Ready'
    };

    localStorage.setItem(REPORT_HISTORY_KEY, JSON.stringify([newReport, ...history].slice(0, 50)));
};

export const exportToCSV = (data: Record<string, unknown>[], filename: string) => {
    if (!data || data.length === 0) return;

    // 1. Extract headers
    const headers = Object.keys(data[0]);

    // 2. Map data to rows
    const rows = data.map(row =>
        headers.map(header => {
            const val = row[header];
            // Escape quotes and wrap in quotes if contains comma
            const escaped = String(val).replace(/"/g, '""');
            return `"${escaped}"`;
        }).join(',')
    );

    // 3. Combine headers and rows
    const csvContent = [headers.join(','), ...rows].join('\n');

    // 4. Create Blob and trigger download
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });

    addReportToHistory(filename, 'CSV', blob.size);

    const link = document.createElement('a');
    if (link.download !== undefined) {
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', `${filename}.csv`);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }
};

export const exportToJSON = (data: Record<string, unknown>[], filename: string) => {
    if (!data || data.length === 0) return;

    const jsonContent = JSON.stringify(data, null, 2);
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });

    addReportToHistory(filename, 'JSON', blob.size);

    const link = document.createElement('a');
    if (link.download !== undefined) {
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', `${filename}.json`);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }
};

/**
 * Triggers a focused print operation for a specific element
 * This is a lightweight alternative to jspdf/html2canvas
 */
export const exportToPDF = (elementId: string, title: string) => {
    const element = document.getElementById(elementId);
    if (!element) return;

    // Estimate file size (very rough)
    addReportToHistory(title, 'PDF', element.innerHTML.length * 2);

    // We use a temporary print stylesheet strategy
    const printWindow = window.open('', '_blank');
    if (!printWindow) return;

    const styles = Array.from(document.styleSheets)
        .map(styleSheet => {
            try {
                return Array.from(styleSheet.cssRules)
                    .map(rule => rule.cssText)
                    .join('');
            } catch {
                return '';
            }
        })
        .join('');

    const now = new Date().toLocaleString();

    printWindow.document.write(`
        <html>
            <head>
                <title>${title}</title>
                <style>
                    ${styles}
                    @media print {
                        body { background: white !important; color: black !important; padding: 0 !important; }
                        .no-print { display: none !important; }
                        #print-target { width: 100% !important; height: auto !important; }
                        .report-header { border-bottom: 2px solid #1e293b; margin-bottom: 40px; padding-bottom: 20px; }
                    }
                    body { font-family: 'Inter', sans-serif; padding: 40px; background: #fff; }
                    .report-header { display: flex; justify-content: space-between; align-items: start; border-bottom: 2px solid #f1f5f9; margin-bottom: 30px; padding-bottom: 20px; }
                    .logo { font-weight: 800; font-size: 24px; color: #6366f1; }
                    .meta { text-align: right; color: #64748b; font-size: 12px; }
                </style>
            </head>
            <body>
                <div id="print-target">
                    <div class="report-header">
                        <div>
                            <div class="logo">ENTERPRISE CORE</div>
                            <div style="font-size: 12px; color: #94a3b8; margin-top: 4px;">Intelligence Intelligence Layer</div>
                        </div>
                        <div class="meta">
                            <div style="font-weight: bold; color: #1e293b;">${title}</div>
                            <div>Generated: ${now}</div>
                            <div>Classification: INTERNAL USE ONLY</div>
                        </div>
                    </div>
                    ${element.innerHTML}
                </div>
                <script>
                    window.onload = () => {
                        window.print();
                        setTimeout(() => window.close(), 500);
                    };
                </script>
            </body>
        </html>
    `);
    printWindow.document.close();
};
