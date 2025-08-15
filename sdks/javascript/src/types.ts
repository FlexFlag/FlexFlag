/**
 * FlexFlag SDK Type Definitions
 */

export type FlagValue = boolean | string | number | object | null;

export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'none';

export type ConnectionMode = 'streaming' | 'polling' | 'offline';

export type CacheStorage = 'memory' | 'localStorage' | 'sessionStorage' | 'custom';

export interface FlexFlagConfig {
  /**
   * API key for authentication
   */
  apiKey: string;

  /**
   * Base URL for FlexFlag server or edge server
   * Default: 'https://api.flexflag.io'
   */
  baseUrl?: string;

  /**
   * Environment to use (production, staging, development)
   * Default: 'production'
   */
  environment?: string;

  /**
   * Cache configuration
   */
  cache?: CacheConfig;

  /**
   * Connection configuration
   */
  connection?: ConnectionConfig;

  /**
   * Offline mode configuration
   */
  offline?: OfflineConfig;

  /**
   * Performance optimization settings
   */
  performance?: PerformanceConfig;

  /**
   * Logging configuration
   */
  logging?: LoggingConfig;

  /**
   * Event handlers
   */
  events?: EventHandlers;
}

export interface CacheConfig {
  /**
   * Enable caching
   * Default: true
   */
  enabled?: boolean;

  /**
   * Cache time-to-live in milliseconds
   * Default: 300000 (5 minutes)
   */
  ttl?: number;

  /**
   * Maximum number of flags to cache
   * Default: 1000
   */
  maxSize?: number;

  /**
   * Storage type for cache
   * Default: 'memory'
   */
  storage?: CacheStorage;

  /**
   * Custom cache provider
   */
  provider?: CacheProvider;

  /**
   * Enable cache compression
   * Default: false
   */
  compression?: boolean;

  /**
   * Cache key prefix
   * Default: 'flexflag:'
   */
  keyPrefix?: string;
}

export interface ConnectionConfig {
  /**
   * Connection mode
   * Default: 'streaming'
   */
  mode?: ConnectionMode;

  /**
   * Polling interval in milliseconds (for polling mode)
   * Default: 30000 (30 seconds)
   */
  pollingInterval?: number;

  /**
   * Request timeout in milliseconds
   * Default: 5000 (5 seconds)
   */
  timeout?: number;

  /**
   * Number of retry attempts
   * Default: 3
   */
  retryAttempts?: number;

  /**
   * Delay between retries in milliseconds
   * Default: 1000 (1 second)
   */
  retryDelay?: number;

  /**
   * Enable exponential backoff for retries
   * Default: true
   */
  exponentialBackoff?: boolean;

  /**
   * Custom headers for requests
   */
  headers?: Record<string, string>;
}

export interface OfflineConfig {
  /**
   * Enable offline mode
   * Default: true
   */
  enabled?: boolean;

  /**
   * Default flag values for offline mode
   */
  defaultFlags?: Record<string, FlagValue>;

  /**
   * Persist flags to storage for offline use
   * Default: true
   */
  persistence?: boolean;

  /**
   * Storage key for persisted flags
   * Default: 'flexflag_offline'
   */
  storageKey?: string;
}

export interface PerformanceConfig {
  /**
   * Evaluation mode
   * 'cached': Use cache first (default)
   * 'always-fetch': Always fetch from server
   * 'lazy': Fetch only when needed
   */
  evaluationMode?: 'cached' | 'always-fetch' | 'lazy';

  /**
   * Enable request batching
   * Default: true
   */
  batchRequests?: boolean;

  /**
   * Batch request interval in milliseconds
   * Default: 100
   */
  batchInterval?: number;

  /**
   * Enable response compression
   * Default: true
   */
  compressionEnabled?: boolean;

  /**
   * Prefetch flags on initialization
   * Default: true
   */
  prefetch?: boolean;

  /**
   * List of flags to prefetch (if not all)
   */
  prefetchFlags?: string[];
}

export interface LoggingConfig {
  /**
   * Log level
   * Default: 'warn'
   */
  level?: LogLevel;

  /**
   * Custom logger implementation
   */
  logger?: Logger;

  /**
   * Include timestamps in logs
   * Default: true
   */
  timestamps?: boolean;
}

export interface EventHandlers {
  /**
   * Called when SDK is ready
   */
  onReady?: () => void;

  /**
   * Called when a flag is evaluated
   */
  onEvaluation?: (flagKey: string, value: FlagValue) => void;

  /**
   * Called when flags are updated
   */
  onUpdate?: (flags: string[]) => void;

  /**
   * Called on error
   */
  onError?: (error: Error) => void;

  /**
   * Called when cache is hit
   */
  onCacheHit?: (flagKey: string) => void;

  /**
   * Called when cache is missed
   */
  onCacheMiss?: (flagKey: string) => void;
}

export interface EvaluationContext {
  /**
   * User identifier
   */
  userId?: string;

  /**
   * User attributes for targeting
   */
  attributes?: Record<string, any>;

  /**
   * Device information
   */
  device?: DeviceContext;

  /**
   * Session information
   */
  session?: SessionContext;
}

export interface DeviceContext {
  type?: 'mobile' | 'tablet' | 'desktop' | 'tv' | 'other';
  os?: string;
  browser?: string;
  version?: string;
  language?: string;
  timezone?: string;
}

export interface SessionContext {
  id?: string;
  startTime?: Date;
  referrer?: string;
  utmSource?: string;
  utmMedium?: string;
  utmCampaign?: string;
}

export interface Flag {
  key: string;
  value: FlagValue;
  type: 'boolean' | 'string' | 'number' | 'json' | 'variant';
  enabled: boolean;
  variations?: Variation[];
  targeting?: TargetingRule[];
  metadata?: FlagMetadata;
}

export interface Variation {
  name: string;
  value: FlagValue;
  weight?: number;
  description?: string;
}

export interface TargetingRule {
  attribute: string;
  operator: 'equals' | 'not_equals' | 'contains' | 'not_contains' | 'greater_than' | 'less_than' | 'in' | 'not_in';
  value: any;
  variation?: string;
}

export interface FlagMetadata {
  createdAt: Date;
  updatedAt: Date;
  environment: string;
  projectId: string;
  tags?: string[];
}

export interface CacheProvider {
  /**
   * Get a value from cache
   */
  get(key: string): Promise<FlagValue | null>;

  /**
   * Set a value in cache
   */
  set(key: string, value: FlagValue, ttl?: number): Promise<void>;

  /**
   * Delete a value from cache
   */
  delete(key: string): Promise<void>;

  /**
   * Clear all cached values
   */
  clear(): Promise<void>;

  /**
   * Check if key exists in cache
   */
  has(key: string): Promise<boolean>;

  /**
   * Get cache size
   */
  size(): Promise<number>;

  /**
   * Get all cache keys
   */
  keys(): Promise<string[]>;
}

export interface Logger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
}

export interface EvaluationResult {
  value: FlagValue;
  variation?: string;
  reason: EvaluationReason;
  metadata?: EvaluationMetadata;
}

export type EvaluationReason = 
  | 'TARGETING_MATCH'
  | 'DEFAULT'
  | 'DISABLED'
  | 'CACHED'
  | 'OFFLINE'
  | 'ERROR';

export interface EvaluationMetadata {
  timestamp: Date;
  cacheHit: boolean;
  evaluationTime: number;
  source: 'cache' | 'local' | 'edge' | 'main' | 'offline';
}

export interface SDKMetrics {
  evaluations: number;
  cacheHits: number;
  cacheMisses: number;
  errors: number;
  networkRequests: number;
  averageLatency: number;
}

export interface BatchEvaluationRequest {
  flags: string[];
  context: EvaluationContext;
}

export interface BatchEvaluationResponse {
  flags: Record<string, FlagValue>;
  errors?: Record<string, string>;
}