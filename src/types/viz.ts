export type ThresholdSeverity = 'info' | 'warning' | 'critical' | 'success';

export interface Threshold {
    value: number;
    color: string;
    label?: string;
    severity?: ThresholdSeverity;
}

export interface FormattingRule {
    field: string;
    condition: 'gt' | 'lt' | 'eq' | 'contains';
    threshold: number | string;
    color: string;
    severity?: ThresholdSeverity;
}
