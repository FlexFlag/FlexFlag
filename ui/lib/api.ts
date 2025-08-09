import { Flag, CreateFlagRequest, EvaluationRequest, EvaluationResponse, PerformanceStats, UltraFastStats } from '@/types';

const API_BASE = '/api/v1';

class ApiClient {
  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  // Flag Management
  async createFlag(flag: CreateFlagRequest, environment = 'production'): Promise<Flag> {
    return this.request<Flag>('/flags', {
      method: 'POST',
      body: JSON.stringify(flag),
      headers: {
        'X-Environment': environment,
      },
    });
  }

  async getFlags(environment = 'production'): Promise<Flag[]> {
    const response = await this.request<{flags: Flag[]}>(`/flags?environment=${environment}`);
    return response.flags || [];
  }

  async getFlag(key: string, environment = 'production'): Promise<Flag> {
    return this.request<Flag>(`/flags/${key}?environment=${environment}`);
  }

  async updateFlag(key: string, flag: Partial<Flag>, environment = 'production'): Promise<Flag> {
    return this.request<Flag>(`/flags/${key}`, {
      method: 'PUT',
      body: JSON.stringify(flag),
      headers: {
        'X-Environment': environment,
      },
    });
  }

  async deleteFlag(key: string, environment = 'production'): Promise<void> {
    return this.request<void>(`/flags/${key}?environment=${environment}`, {
      method: 'DELETE',
    });
  }

  async toggleFlag(key: string, environment = 'production'): Promise<Flag> {
    return this.request<Flag>(`/flags/${key}/toggle?environment=${environment}`, {
      method: 'POST',
    });
  }

  // Flag Evaluation
  async evaluateFlag(request: EvaluationRequest, environment = 'production'): Promise<EvaluationResponse> {
    return this.request<EvaluationResponse>(`/evaluate?environment=${environment}`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async evaluateFlagFast(request: EvaluationRequest, environment = 'production'): Promise<EvaluationResponse> {
    return this.request<EvaluationResponse>(`/evaluate/fast?environment=${environment}`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async evaluateFlagUltraFast(request: EvaluationRequest, environment = 'production'): Promise<EvaluationResponse> {
    return this.request<EvaluationResponse>(`/evaluate/ultra?environment=${environment}`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async batchEvaluate(
    requests: { flag_keys: string[]; user_id: string; user_key?: string; attributes?: Record<string, any> },
    environment = 'production'
  ): Promise<Record<string, any>> {
    return this.request<Record<string, any>>(`/evaluate/batch?environment=${environment}`, {
      method: 'POST',
      body: JSON.stringify(requests),
    });
  }

  // Performance & Stats
  async getCacheStats(): Promise<PerformanceStats> {
    const response = await this.request<{cache_stats: any}>('/evaluate/cache/stats');
    // Transform the response to match our interface
    return {
      cache_hits: 0,
      cache_misses: 0,
      total_requests: 0,
      average_evaluation_time_ms: 0,
      p95_evaluation_time_ms: 0,
      p99_evaluation_time_ms: 0,
      ...response.cache_stats,
    };
  }

  async getUltraFastStats(): Promise<UltraFastStats> {
    return this.request<UltraFastStats>('/evaluate/ultra/stats');
  }

  async clearCache(): Promise<void> {
    return this.request<void>('/evaluate/cache/clear', {
      method: 'POST',
    });
  }

  // Health Check
  async healthCheck(): Promise<{ status: string; service: string }> {
    const response = await fetch('http://localhost:8080/health');
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    return response.json();
  }
}

export const apiClient = new ApiClient();
export default apiClient;