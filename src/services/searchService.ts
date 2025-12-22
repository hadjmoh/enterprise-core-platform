export interface CostEstimate {
  estimated_rows: number;
  estimated_scan_gb: number;
  estimated_cpu_seconds: number;
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  risk_factors: string[];
  warnings: string[];
}

export interface SuggestResponse {
  suggestion: {
    original_query: string;
    suggested_spl: string;
    confidence_score: number;
    explanation: string;
    cost_estimate: CostEstimate;
    allowed: boolean;
    blocking_reason: string;
  };
  error?: string;
}

export interface EstimateResponse {
  estimate: CostEstimate;
  allowed: boolean;
  reason: string;
}

export class SearchService {
  private static readonly API_BASE = '/api';

  static async estimateCost(query: string, timeRange?: { start: string; end: string }): Promise<EstimateResponse> {
    const response = await fetch(`${this.API_BASE}/search/estimate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // 'Authorization': ... (Handled by interceptor or context)
      },
      body: JSON.stringify({
        query,
        time_start: timeRange?.start,
        time_end: timeRange?.end,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to estimate cost: ${response.statusText}`);
    }

    return response.json();
  }

  static async suggestQuery(query: string): Promise<SuggestResponse> {
    const response = await fetch(`${this.API_BASE}/pilot/suggest`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ query }),
    });

    if (!response.ok) {
      throw new Error(`Failed to get suggestions: ${response.statusText}`);
    }

    return response.json();
  }
}
