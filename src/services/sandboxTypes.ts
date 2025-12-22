export interface ResourceLimits {
    CPUPercent: number;
    MemoryMB: number;
    DiskQuotaMB: number;
    NetworkRateMB: number;
}

export interface ResourceUsage {
    CPUPercent: number;
    MemoryMB: number;
    DiskUsedMB: number;
    NetworkRateMB: number;
    LastUpdated: string;
}
