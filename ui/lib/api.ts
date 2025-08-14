import { Flag, CreateFlagRequest, EvaluationRequest, EvaluationResponse, PerformanceStats, UltraFastStats } from '@/types';

const API_BASE = '/api/v1';

class ApiClient {
  private getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('token');
    }
    return null;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const token = this.getAuthToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options?.headers,
    };

    // Add authorization header if token exists
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
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

  async getFlags(environment = 'production', projectId?: string): Promise<Flag[]> {
    const params = new URLSearchParams({ environment });
    if (projectId) {
      params.append('project_id', projectId);
    }
    const response = await this.request<{flags: Flag[]}>(`/flags?${params.toString()}`);
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

  async toggleFlag(key: string, environment = 'production', projectId?: string): Promise<Flag> {
    const params = new URLSearchParams({ environment });
    if (projectId) {
      params.append('project_id', projectId);
    }
    return this.request<Flag>(`/flags/${key}/toggle?${params.toString()}`, {
      method: 'POST',
    });
  }

  // Flag Evaluation
  async evaluateFlag(request: EvaluationRequest, environment = 'production', projectId?: string): Promise<EvaluationResponse> {
    const params = new URLSearchParams({ environment });
    if (projectId) {
      params.append('project_id', projectId);
    }
    return this.request<EvaluationResponse>(`/evaluate?${params.toString()}`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async evaluateFlagFast(request: EvaluationRequest, environment = 'production', projectId?: string): Promise<EvaluationResponse> {
    const params = new URLSearchParams({ environment });
    if (projectId) {
      params.append('project_id', projectId);
    }
    return this.request<EvaluationResponse>(`/evaluate/fast?${params.toString()}`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async evaluateFlagUltraFast(request: EvaluationRequest, environment = 'production', projectId?: string): Promise<EvaluationResponse> {
    const params = new URLSearchParams({ environment });
    if (projectId) {
      params.append('project_id', projectId);
    }
    return this.request<EvaluationResponse>(`/evaluate/ultra?${params.toString()}`, {
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

  // Projects Management
  async getProjects(): Promise<any[]> {
    const response = await this.request<{projects: any[]}>('/projects');
    return response.projects || [];
  }

  async createProject(project: any): Promise<any> {
    return this.request<any>('/projects', {
      method: 'POST',
      body: JSON.stringify(project),
    });
  }

  async getProjectStats(projectId: string): Promise<{flags: number, segments: number, rollouts: number}> {
    return this.request<{flags: number, segments: number, rollouts: number}>(`/project-stats/${projectId}`);
  }

  async getProjectEnvironments(projectSlug: string): Promise<any[]> {
    const response = await this.request<{environments: any[]}>(`/projects/${projectSlug}/environments`);
    return response.environments || [];
  }

  // Audit Logs (placeholder - backend endpoint would be needed)
  async getAuditLogs(projectId?: string): Promise<any[]> {
    const params = new URLSearchParams();
    if (projectId) {
      params.append('project_id', projectId);
    }
    
    const response = await this.request<{logs: any[]}>(`/audit/logs?${params.toString()}`);
    return response.logs || [];
  }

  // User Management (placeholder)
  async getUsers(): Promise<any[]> {
    // Mock data for now
    return Promise.resolve([
      {
        id: '1',
        email: 'admin@example.com',
        full_name: 'Admin User',
        role: 'admin',
        is_active: true,
        created_at: new Date().toISOString(),
        last_login: new Date().toISOString(),
      }
    ]);
  }

  // API Key Management
  async createApiKey(projectId: string, apiKey: any): Promise<any> {
    return this.request<any>(`/project-api-keys/${projectId}`, {
      method: 'POST',
      body: JSON.stringify(apiKey),
    });
  }

  async getApiKeys(projectId: string): Promise<any[]> {
    const response = await this.request<{api_keys: any[]}>(`/project-api-keys/${projectId}`);
    return response.api_keys || [];
  }

  async updateApiKey(projectId: string, keyId: string, updates: any): Promise<void> {
    return this.request<void>(`/project-api-keys/${projectId}/${keyId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteApiKey(projectId: string, keyId: string): Promise<void> {
    return this.request<void>(`/project-api-keys/${projectId}/${keyId}`, {
      method: 'DELETE',
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