export interface Flag {
  id?: string;
  project_id?: string;
  key: string;
  name: string;
  description?: string;
  type: 'boolean' | 'string' | 'number' | 'json';
  enabled: boolean;
  default: any;
  variations?: Variation[];
  targeting?: TargetingConfig;
  environments: string[];
  created_at?: string;
  updated_at?: string;
}

export interface Variation {
  id: string;
  name: string;
  value: any;
  description?: string;
  weight?: number;
}

export interface TargetingRule {
  id: string;
  attribute: string;
  operator: 'eq' | 'ne' | 'in' | 'nin' | 'contains' | 'startsWith' | 'endsWith' | 'gt' | 'gte' | 'lt' | 'lte';
  values: string[];
  variation: string;
}

export interface TargetingConfig {
  rules: TargetingRule[];
  rollout?: {
    variations: {
      variation_id: string;
      weight: number;
    }[];
  };
  default_rule?: {
    variation: string;
  };
}

export interface EvaluationRequest {
  flag_key: string;
  user_id: string;
  user_key?: string;
  attributes?: Record<string, any>;
}

export interface EvaluationResponse {
  flag_key: string;
  value: any;
  variation?: string;
  reason: string;
  rule_id?: string;
  default: boolean;
  evaluation_time_ms: number;
  timestamp: string;
}

export interface PerformanceStats {
  cache_hits: number;
  cache_misses: number;
  total_requests: number;
  average_evaluation_time_ms: number;
  p95_evaluation_time_ms: number;
  p99_evaluation_time_ms: number;
}

export interface UltraFastStats {
  preloaded_flags: number;
  cached_responses: number;
  preload_complete: boolean;
}

export type Environment = 'production' | 'staging' | 'development';

export interface CreateFlagRequest {
  key: string;
  name: string;
  description?: string;
  type: Flag['type'];
  enabled: boolean;
  default: any;
  project_id?: string;
  variations?: Omit<Variation, 'id'>[];
  targeting?: TargetingConfig;
}