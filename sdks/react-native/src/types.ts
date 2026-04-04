export type FlagValue = boolean | string | number | Record<string, unknown> | null;

export type FlagType = 'boolean' | 'string' | 'number' | 'json' | 'variant';

export interface EvaluationContext {
  userId?: string;
  userKey?: string;
  /** Client platform — "ios", "android", or "web". Used for platform targeting in remote config. */
  platform?: 'ios' | 'android' | 'web';
  attributes?: Record<string, string | number | boolean>;
}

export interface FlexFlagConfig {
  /** FlexFlag API key (required) */
  apiKey: string;
  /** Base URL of your FlexFlag server. Default: 'http://localhost:8080' */
  baseUrl?: string;
  /** Environment to evaluate flags in. Default: 'production' */
  environment?: string;
  /** Polling interval in ms. Default: 30000 (30s). Set to 0 to disable polling. */
  pollingInterval?: number;
  /** HTTP request timeout in ms. Default: 5000 */
  timeout?: number;
  /** Default flag values used when offline or flag not found */
  offlineDefaults?: Record<string, FlagValue>;
  /** AsyncStorage key for persisting the flag cache. Default: 'flexflag_cache' */
  persistenceKey?: string;
  /** Cache TTL in ms. Default: 300000 (5 minutes) */
  cacheTtl?: number;
}

export interface EvaluationResult {
  value: FlagValue;
  variation?: string;
  reason: string;
  default: boolean;
}

export interface ConfigMeta {
  type: FlagType;
  enabled: boolean;
  updated_at: string;
}

export interface ConfigResult {
  config: Record<string, FlagValue>;
  metadata: Record<string, ConfigMeta>;
  environment: string;
  timestamp: string;
}

export interface CachedEvaluation {
  value: FlagValue;
  variation?: string;
  reason: string;
  cachedAt: number;
  ttl: number;
}

export interface PersistedState {
  evaluations: Record<string, CachedEvaluation>;
  config: ConfigResult | null;
  savedAt: number;
}
