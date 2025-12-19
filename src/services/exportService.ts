import jsPDF from 'jspdf';
import html2canvas from 'html2canvas';

/**
 * Export Service
 * Handles data conversion and professional PDF/CSV generation.
 */

const REPORT_HISTORY_KEY = 'enterprise_core_report_history';

export interface ReportRecord {
    id: string;
    name: string;
    type: 'CSV' | 'PDF' | 'JSON';
    timestamp: string;
    size: string;
    status: 'Ready' | 'Expired';
    query?: string;
    user?: string;
    timeRange?: string;
}

export const getReportHistory = (): ReportRecord[] => {
    try {
        const saved = localStorage.getItem(REPORT_HISTORY_KEY);
        return saved ? JSON.parse(saved) : [];
    } catch (e) {
        console.error('Failed to load report history', e);
        return [];
    }
};

export const clearReportHistory = () => {
    localStorage.removeItem(REPORT_HISTORY_KEY);
};

export const deleteReport = (id: string) => {
    const history = getReportHistory();
    localStorage.setItem(REPORT_HISTORY_KEY, JSON.stringify(history.filter(r => r.id !== id)));
};

interface ExportMetadata {
    query?: string;
    user?: string;
    timeRange?: string;
}

const addReportToHistory = (
    name: string,
    type: ReportRecord['type'],
    sizeInBytes: number,
    meta?: ExportMetadata
) => {
    const history = getReportHistory();

    const sizeStr = sizeInBytes > 1024 * 1024
        ? `${(sizeInBytes / (1024 * 1024)).toFixed(1)} MB`
        : `${(sizeInBytes / 1024).toFixed(0)} KB`;

    const newReport: ReportRecord = {
        id: Math.random().toString(36).substring(2, 9),
        name,
        type,
        timestamp: new Date().toISOString().replace('T', ' ').substring(0, 16),
        size: sizeStr,
        status: 'Ready',
        query: meta?.query,
        user: meta?.user || 'admin',
        timeRange: meta?.timeRange || 'All Time'
    };

    localStorage.setItem(REPORT_HISTORY_KEY, JSON.stringify([newReport, ...history].slice(0, 50)));
};

export const exportToCSV = (data: Record<string, unknown>[], filename: string, meta?: ExportMetadata) => {
    if (!data || data.length === 0) return;

    const headers = Object.keys(data[0]);
    const rows = data.map(row =>
        headers.map(header => {
            const val = row[header];
            const escaped = String(val).replace(/"/g, '""');
            return `"${escaped}"`;
        }).join(',')
    );

    const csvContent = [headers.join(','), ...rows].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });

    addReportToHistory(filename, 'CSV', blob.size, meta);

    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `${filename}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
};

export const exportToJSON = (data: Record<string, unknown>[], filename: string, meta?: ExportMetadata) => {
    if (!data || data.length === 0) return;

    const jsonContent = JSON.stringify(data, null, 2);
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });

    addReportToHistory(filename, 'JSON', blob.size, meta);

    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `${filename}.json`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
};

/**
 * Generates a professional PDF snapshot of a DOM element with watermarking
 */
export const exportToPDF = async (elementId: string, title: string, meta?: ExportMetadata) => {
    const element = document.getElementById(elementId);
    if (!element) return;

    try {
        const canvas = await html2canvas(element, {
            scale: 2,
            backgroundColor: '#0f172a',
            logging: false,
            useCORS: true
        });

        const imgData = canvas.toDataURL('image/png');
        const pdf = new jsPDF({
            orientation: 'landscape',
            unit: 'px',
            format: [canvas.width, canvas.height]
        });

        const width = pdf.internal.pageSize.getWidth();
        const height = pdf.internal.pageSize.getHeight();

        // Background Watermark (diagonal)
        pdf.setTextColor(255, 255, 255, 0.05);
        pdf.setFontSize(80);
        pdf.text('CONFIDENTIAL', 50, height - 100, { angle: 45 });
        pdf.text('CONFIDENTIAL', width / 2, height / 2, { angle: 45 });

        pdf.addImage(imgData, 'PNG', 0, 0, width, height);

        // Add Header Branding
        pdf.setFillColor(30, 41, 59); // slate-800
        pdf.rect(0, 0, width, 40, 'F');
        pdf.setTextColor(255, 255, 255);
        pdf.setFontSize(14);
        pdf.text('ENTERPRISE CORE PLATFORM', 20, 25);
        pdf.setFontSize(10);
        pdf.setTextColor(148, 163, 184); // slate-400
        pdf.text(`Report: ${title} | generated ${new Date().toLocaleString()}`, width - 20, 25, { align: 'right' });

        const pdfBlob = pdf.output('blob');
        addReportToHistory(title, 'PDF', pdfBlob.size, meta);
        pdf.save(`${title.replace(/\s+/g, '_').toLowerCase()}.pdf`);
    } catch (error) {
        console.error('PDF Generation failed:', error);
    }
};
