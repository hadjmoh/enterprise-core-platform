import { v4 as uuidv4 } from 'uuid';

/**
 * Scheduling Service
 * Manages recurring report automated triggers.
 */

export interface ScheduledReport {
    id: string;
    name: string;
    freq: 'Daily' | 'Weekly' | 'Monthly';
    type: 'PDF' | 'CSV' | 'JSON';
    query: string;
    recipients: string[];
    lastRun?: string;
    status: 'Active' | 'Paused';
}

const SCHEDULE_KEY = 'enterprise_core_reporting_schedules';

export const getSchedules = (): ScheduledReport[] => {
    try {
        const saved = localStorage.getItem(SCHEDULE_KEY);
        return saved ? JSON.parse(saved) : [];
    } catch {
        return [];
    }
};

export const saveSchedule = (schedule: ScheduledReport) => {
    const schedules = getSchedules();
    const existingIndex = schedules.findIndex(s => s.id === schedule.id);
    if (existingIndex > -1) {
        schedules[existingIndex] = schedule;
    } else {
        schedules.push(schedule);
    }
    localStorage.setItem(SCHEDULE_KEY, JSON.stringify(schedules));
};

export const deleteSchedule = (id: string) => {
    const schedules = getSchedules();
    localStorage.setItem(SCHEDULE_KEY, JSON.stringify(schedules.filter(s => s.id !== id)));
};

export const createMockSchedule = (count: number) => {
    const names = ['Security Audit', 'Ingestion Health', 'Top Violations', 'Compliance Snapshot'];
    const freqs: ('Daily' | 'Weekly')[] = ['Daily', 'Weekly'];

    for (let i = 0; i < count; i++) {
        saveSchedule({
            id: uuidv4(),
            name: names[i % names.length],
            freq: freqs[i % freqs.length],
            type: 'PDF',
            query: 'index=main | stats count by host',
            recipients: ['admin@enterprise.local'],
            status: 'Active'
        });
    }
};
