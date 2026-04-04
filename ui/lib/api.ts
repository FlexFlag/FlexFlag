import { Flag, CreateFlagRequest, EvaluationRequest, EvaluationResponse, PerformanceStats, UltraFastStats } from '@/types';

const API_BASE = 'http://localhost:8080/api/v1';

class ApiClient {
  private getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('token');
    }
    return null;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const token = this.getAuthToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options?.headers as Record<string, string>),
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
    return this.request<Flag>(`/flags?environment=${environment}`, {
      method: 'POST',
      body: JSON.stringify(flag),
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
    return this.request<Flag>(`/flags/${key}?environment=${environment}`, {
      method: 'PUT',
      body: JSON.stringify(flag),
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

  async getProjectMembers(projectId: string): Promise<any[]> {
    const response = await this.request<{members: any[]}>(`/project-members/${projectId}`);
    return response.members || [];
  }

  async addProjectMember(projectId: string, userId: string, role: string): Promise<{message: string}> {
    return this.request<{message: string}>(`/project-members/${projectId}`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, role }),
    });
  }

  async removeProjectMember(projectId: string, userId: string): Promise<{message: string}> {
    return this.request<{message: string}>(`/project-members/${projectId}/${userId}`, {
      method: 'DELETE',
    });
  }

  async updateProjectMemberRole(projectId: string, userId: string, role: string): Promise<{message: string}> {
    return this.request<{message: string}>(`/project-members/${projectId}/${userId}`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    });
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

  // User Management
  async getUsers(): Promise<any[]> {
    const response = await this.request<{users: any[], total: number}>('/users');
    return response.users || [];
  }

  async createUser(user: {email: string, full_name: string, password?: string, role?: string}): Promise<{user: any, password: string, message: string}> {
    return this.request<{user: any, password: string, message: string}>('/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
  }

  async updateUser(userId: string, updates: {full_name?: string, role?: string, is_active?: boolean}): Promise<any> {
    return this.request<any>(`/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  }

  async deleteUser(userId: string): Promise<void> {
    return this.request<void>(`/users/${userId}`, {
      method: 'DELETE',
    });
  }

  async resetUserPassword(userId: string): Promise<{message: string, password: string}> {
    return this.request<{message: string, password: string}>(`/users/${userId}/reset-password`, {
      method: 'POST',
    });
  }

  // Profile Management
  async updateProfile(full_name: string): Promise<any> {
    return this.request<any>('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify({ full_name }),
    });
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

  // Remote Config
  async getRemoteConfig(environment = 'production', projectId?: string): Promise<{
    config: Record<string, unknown>;
    metadata: Record<string, { type: string; enabled: boolean; updated_at: string }>;
    environment: string;
    timestamp: string;
  }> {
    const params = new URLSearchParams({ environment });
    if (projectId) params.append('project_id', projectId);
    return this.request(`/config?${params.toString()}`);
  }

  async evaluateRemoteConfig(
    context: { user_id?: string; user_key?: string; attributes?: Record<string, unknown> },
    environment = 'production',
    projectId?: string
  ): Promise<{
    config: Record<string, unknown>;
    metadata: Record<string, { type: string; enabled: boolean; updated_at: string }>;
    environment: string;
    timestamp: string;
  }> {
    const params = new URLSearchParams({ environment });
    if (projectId) params.append('project_id', projectId);
    return this.request(`/config/evaluate?${params.toString()}`, {
      method: 'POST',
      body: JSON.stringify(context),
    });
  }

  // Analytics
  async getFlagAnalytics(environment = 'production'): Promise<{
    flags: Record<string, {
      total_evaluations: number;
      unique_users: number;
      last_evaluated_at?: string;
      variation_counts: Record<string, number>;
    }>;
    environment: string;
  }> {
    return this.request(`/analytics/flags?environment=${environment}`);
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