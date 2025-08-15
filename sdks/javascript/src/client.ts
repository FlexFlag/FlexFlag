/**
 * FlexFlag Client Implementation
 * Main SDK client with intelligent caching and offline support
 */

import { EventEmitter } from 'eventemitter3';
import axios, { AxiosInstance } from 'axios';
import {
  FlexFlagConfig,
  EvaluationContext,
  FlagValue,
  Flag,
  EvaluationResult,
  EvaluationReason,
  SDKMetrics,
  BatchEvaluationRequest,
  BatchEvaluationResponse,
  Logger,
  ConnectionMode
} from './types';
import { CacheProvider, MemoryCache, LocalStorageCache } from './cache';
import { DefaultLogger } from './logger';

export class FlexFlagClient extends EventEmitter {
  private config: Required<FlexFlagConfig>;
  private cache: CacheProvider;
  private httpClient: AxiosInstance;
  private logger: Logger;
  private ready: boolean = false;
  private connectionMode: ConnectionMode;
  private ws?: WebSocket;
  private pollingInterval?: NodeJS.Timeout;
  private defaultContext: EvaluationContext = {};
  private metrics: SDKMetrics = {
    evaluations: 0,
    cacheHits: 0,
    cacheMisses: 0,
    errors: 0,
    networkRequests: 0,
    averageLatency: 0
  };
  private batchQueue: Map<string, Array<(value: FlagValue) => void>> = new Map();
  private batchTimer?: NodeJS.Timeout;

  constructor(config: FlexFlagConfig) {
    super();
    
    // Validate required config
    if (!config.apiKey) {
      throw new Error('FlexFlag: API key is required');
    }

    // Apply defaults
    this.config = this.applyDefaults(config);
    
    // Initialize logger
    this.logger = this.config.logging.logger || new DefaultLogger(this.config.logging.level);
    
    // Initialize cache
    this.cache = this.initializeCache();
    
    // Initialize HTTP client
    this.httpClient = this.initializeHttpClient();
    
    // Set connection mode
    this.connectionMode = this.config.connection.mode || 'streaming';
    
    // Initialize SDK
    this.initialize();
  }

  private applyDefaults(config: FlexFlagConfig): Required<FlexFlagConfig> {
    return {
      apiKey: config.apiKey,
      baseUrl: config.baseUrl || 'https://api.flexflag.io',
      environment: config.environment || 'production',
      cache: {
        enabled: config.cache?.enabled !== false,
        ttl: config.cache?.ttl || 300000, // 5 minutes
        maxSize: config.cache?.maxSize || 1000,
        storage: config.cache?.storage || 'memory',
        provider: config.cache?.provider,
        compression: config.cache?.compression || false,
        keyPrefix: config.cache?.keyPrefix || 'flexflag:'
      },
      connection: {
        mode: config.connection?.mode || 'streaming',
        pollingInterval: config.connection?.pollingInterval || 30000,
        timeout: config.connection?.timeout || 5000,
        retryAttempts: config.connection?.retryAttempts || 3,
        retryDelay: config.connection?.retryDelay || 1000,
        exponentialBackoff: config.connection?.exponentialBackoff !== false,
        headers: config.connection?.headers || {}
      },
      offline: {
        enabled: config.offline?.enabled !== false,
        defaultFlags: config.offline?.defaultFlags || {},
        persistence: config.offline?.persistence !== false,
        storageKey: config.offline?.storageKey || 'flexflag_offline'
      },
      performance: {
        evaluationMode: config.performance?.evaluationMode || 'cached',
        batchRequests: config.performance?.batchRequests !== false,
        batchInterval: config.performance?.batchInterval || 100,
        compressionEnabled: config.performance?.compressionEnabled !== false,
        prefetch: config.performance?.prefetch !== false,
        prefetchFlags: config.performance?.prefetchFlags
      },
      logging: {
        level: config.logging?.level || 'warn',
        logger: config.logging?.logger,
        timestamps: config.logging?.timestamps !== false
      },
      events: config.events || {}
    };
  }

  private initializeCache(): CacheProvider {
    if (this.config.cache.provider) {
      return this.config.cache.provider;
    }

    switch (this.config.cache.storage) {
      case 'localStorage':
        if (typeof window !== 'undefined' && window.localStorage) {
          return new LocalStorageCache(this.config.cache);
        }
        // Fall through to memory cache
      case 'sessionStorage':
        if (typeof window !== 'undefined' && window.sessionStorage) {
          return new LocalStorageCache({ ...this.config.cache, storage: 'sessionStorage' });
        }
        // Fall through to memory cache
      case 'memory':
      default:
        return new MemoryCache(this.config.cache);
    }
  }

  private initializeHttpClient(): AxiosInstance {
    return axios.create({
      baseURL: this.config.baseUrl,
      timeout: this.config.connection.timeout,
      headers: {
        'X-API-Key': this.config.apiKey,
        'Content-Type': 'application/json',
        'X-SDK-Version': '1.0.0',
        'X-SDK-Language': 'javascript',
        ...this.config.connection.headers
      }
    });
  }

  private async initialize(): Promise<void> {
    try {
      this.logger.info('Initializing FlexFlag SDK...');
      
      // Load offline flags if available
      if (this.config.offline.enabled && this.config.offline.persistence) {
        await this.loadOfflineFlags();
      }
      
      // Prefetch flags if enabled
      if (this.config.performance.prefetch) {
        await this.prefetchFlags();
      }
      
      // Start connection based on mode
      switch (this.connectionMode) {
        case 'streaming':
          await this.startStreaming();
          break;
        case 'polling':
          this.startPolling();
          break;
        case 'offline':
          this.logger.info('Running in offline mode');
          break;
      }
      
      this.ready = true;
      this.emit('ready');
      if (this.config.events.onReady) {
        this.config.events.onReady();
      }
      
      this.logger.info('FlexFlag SDK initialized successfully');
    } catch (error) {
      this.logger.error('Failed to initialize FlexFlag SDK', error);
      this.handleError(error as Error);
    }
  }

  /**
   * Evaluate a feature flag
   */
  public async evaluate(
    flagKey: string,
    context?: EvaluationContext,
    defaultValue?: FlagValue
  ): Promise<FlagValue> {
    const startTime = Date.now();
    this.metrics.evaluations++;
    
    try {
      // Merge context with default context
      const evalContext = { ...this.defaultContext, ...context };
      
      // Check cache first if enabled
      if (this.config.cache.enabled && this.config.performance.evaluationMode === 'cached') {
        const cacheKey = this.getCacheKey(flagKey, evalContext);
        const cachedValue = await this.cache.get(cacheKey);
        
        if (cachedValue !== null) {
          this.metrics.cacheHits++;
          const evaluationTime = Date.now() - startTime;
          this.updateAverageLatency(evaluationTime);
          
          this.logger.debug(`Cache hit for flag: ${flagKey}`);
          if (this.config.events.onCacheHit) {
            this.config.events.onCacheHit(flagKey);
          }
          
          this.emitEvaluation(flagKey, cachedValue);
          return cachedValue;
        }
        
        this.metrics.cacheMisses++;
        if (this.config.events.onCacheMiss) {
          this.config.events.onCacheMiss(flagKey);
        }
      }
      
      // Fetch from server
      const value = await this.fetchFlag(flagKey, evalContext);
      
      // Cache the result
      if (this.config.cache.enabled) {
        const cacheKey = this.getCacheKey(flagKey, evalContext);
        await this.cache.set(cacheKey, value, this.config.cache.ttl);
      }
      
      const evaluationTime = Date.now() - startTime;
      this.updateAverageLatency(evaluationTime);
      
      this.emitEvaluation(flagKey, value);
      return value;
      
    } catch (error) {
      this.logger.error(`Failed to evaluate flag: ${flagKey}`, error);
      this.metrics.errors++;
      
      // Try offline default
      if (this.config.offline.enabled) {
        const offlineValue = this.config.offline.defaultFlags[flagKey];
        if (offlineValue !== undefined) {
          this.logger.info(`Using offline default for flag: ${flagKey}`);
          return offlineValue;
        }
      }
      
      // Return provided default or null
      return defaultValue !== undefined ? defaultValue : null;
    }
  }

  /**
   * Evaluate multiple flags in batch
   */
  public async evaluateBatch(
    flagKeys: string[],
    context?: EvaluationContext
  ): Promise<Record<string, FlagValue>> {
    if (!this.config.performance.batchRequests) {
      // Evaluate individually
      const results: Record<string, FlagValue> = {};
      for (const key of flagKeys) {
        results[key] = await this.evaluate(key, context);
      }
      return results;
    }
    
    // Batch evaluation
    try {
      const evalContext = { ...this.defaultContext, ...context };
      const response = await this.httpClient.post<BatchEvaluationResponse>(
        '/api/v1/evaluate/batch',
        {
          flags: flagKeys,
          context: evalContext
        } as BatchEvaluationRequest
      );
      
      this.metrics.networkRequests++;
      
      // Cache results
      if (this.config.cache.enabled) {
        for (const [key, value] of Object.entries(response.data.flags)) {
          const cacheKey = this.getCacheKey(key, evalContext);
          await this.cache.set(cacheKey, value, this.config.cache.ttl);
        }
      }
      
      return response.data.flags;
    } catch (error) {
      this.logger.error('Batch evaluation failed', error);
      
      // Fallback to individual evaluation
      const results: Record<string, FlagValue> = {};
      for (const key of flagKeys) {
        results[key] = await this.evaluate(key, context);
      }
      return results;
    }
  }

  /**
   * Get variation for A/B testing
   */
  public async getVariation(
    flagKey: string,
    context?: EvaluationContext
  ): Promise<string | null> {
    const result = await this.evaluateWithDetails(flagKey, context);
    return result.variation || null;
  }

  /**
   * Evaluate with detailed result
   */
  public async evaluateWithDetails(
    flagKey: string,
    context?: EvaluationContext
  ): Promise<EvaluationResult> {
    const startTime = Date.now();
    const evalContext = { ...this.defaultContext, ...context };
    
    try {
      // Check cache for detailed result
      const cacheKey = `${this.getCacheKey(flagKey, evalContext)}:details`;
      const cached = await this.cache.get(cacheKey);
      
      if (cached) {
        return cached as EvaluationResult;
      }
      
      // Fetch detailed evaluation from server
      const response = await this.httpClient.post<EvaluationResult>(
        `/api/v1/evaluate/${flagKey}/details`,
        { context: evalContext }
      );
      
      const result = response.data;
      
      // Cache the detailed result
      if (this.config.cache.enabled) {
        await this.cache.set(cacheKey, result, this.config.cache.ttl);
      }
      
      return result;
    } catch (error) {
      this.logger.error(`Failed to evaluate flag with details: ${flagKey}`, error);
      
      return {
        value: null,
        reason: 'ERROR' as EvaluationReason,
        metadata: {
          timestamp: new Date(),
          cacheHit: false,
          evaluationTime: Date.now() - startTime,
          source: 'offline'
        }
      };
    }
  }

  /**
   * Set default context for all evaluations
   */
  public setContext(context: EvaluationContext): void {
    this.defaultContext = context;
    this.logger.debug('Default context updated', context);
  }

  /**
   * Update context attributes
   */
  public updateContext(updates: Partial<EvaluationContext>): void {
    this.defaultContext = { ...this.defaultContext, ...updates };
    this.logger.debug('Context updated', updates);
  }

  /**
   * Clear all cached flags
   */
  public async clearCache(): Promise<void> {
    await this.cache.clear();
    this.logger.info('Cache cleared');
  }

  /**
   * Get SDK metrics
   */
  public getMetrics(): SDKMetrics {
    return { ...this.metrics };
  }

  /**
   * Reset SDK metrics
   */
  public resetMetrics(): void {
    this.metrics = {
      evaluations: 0,
      cacheHits: 0,
      cacheMisses: 0,
      errors: 0,
      networkRequests: 0,
      averageLatency: 0
    };
  }

  /**
   * Check if SDK is ready
   */
  public isReady(): boolean {
    return this.ready;
  }

  /**
   * Wait for SDK to be ready
   */
  public async waitForReady(timeout: number = 5000): Promise<void> {
    if (this.ready) return;
    
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error('FlexFlag SDK initialization timeout'));
      }, timeout);
      
      this.once('ready', () => {
        clearTimeout(timer);
        resolve();
      });
    });
  }

  /**
   * Close SDK connections and cleanup
   */
  public async close(): Promise<void> {
    this.logger.info('Closing FlexFlag SDK...');
    
    // Stop WebSocket connection
    if (this.ws) {
      this.ws.close();
      this.ws = undefined;
    }
    
    // Stop polling
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
      this.pollingInterval = undefined;
    }
    
    // Clear batch timer
    if (this.batchTimer) {
      clearTimeout(this.batchTimer);
      this.batchTimer = undefined;
    }
    
    // Save offline flags
    if (this.config.offline.enabled && this.config.offline.persistence) {
      await this.saveOfflineFlags();
    }
    
    this.ready = false;
    this.removeAllListeners();
    this.logger.info('FlexFlag SDK closed');
  }

  // Private helper methods

  private async fetchFlag(flagKey: string, context: EvaluationContext): Promise<FlagValue> {
    try {
      const response = await this.httpClient.post(`/api/v1/evaluate`, {
        flag_key: flagKey,
        user_context: context,
        environment: this.config.environment
      });
      
      this.metrics.networkRequests++;
      return response.data.value;
    } catch (error) {
      throw new Error(`Failed to fetch flag ${flagKey}: ${error}`);
    }
  }

  private getCacheKey(flagKey: string, context: EvaluationContext): string {
    const contextKey = context.userId || JSON.stringify(context.attributes || {});
    return `${this.config.cache.keyPrefix}${this.config.environment}:${flagKey}:${contextKey}`;
  }

  private async startStreaming(): Promise<void> {
    if (typeof WebSocket === 'undefined') {
      this.logger.warn('WebSocket not available, falling back to polling');
      this.connectionMode = 'polling';
      this.startPolling();
      return;
    }
    
    const wsUrl = this.config.baseUrl.replace(/^http/, 'ws') + '/api/v1/stream';
    
    try {
      this.ws = new WebSocket(wsUrl, {
        headers: {
          'X-API-Key': this.config.apiKey
        }
      } as any);
      
      this.ws.onopen = () => {
        this.logger.info('WebSocket connection established');
      };
      
      this.ws.onmessage = (event) => {
        this.handleStreamMessage(event.data);
      };
      
      this.ws.onerror = (error) => {
        this.logger.error('WebSocket error', error);
      };
      
      this.ws.onclose = () => {
        this.logger.info('WebSocket connection closed');
        // Attempt reconnection
        setTimeout(() => this.startStreaming(), this.config.connection.retryDelay);
      };
    } catch (error) {
      this.logger.error('Failed to establish WebSocket connection', error);
      // Fall back to polling
      this.connectionMode = 'polling';
      this.startPolling();
    }
  }

  private startPolling(): void {
    this.pollingInterval = setInterval(
      () => this.pollForUpdates(),
      this.config.connection.pollingInterval
    );
    
    // Initial poll
    this.pollForUpdates();
  }

  private async pollForUpdates(): Promise<void> {
    try {
      const response = await this.httpClient.get('/api/v1/flags/changes', {
        params: {
          environment: this.config.environment,
          since: new Date(Date.now() - this.config.connection.pollingInterval).toISOString()
        }
      });
      
      if (response.data.changes && response.data.changes.length > 0) {
        for (const change of response.data.changes) {
          await this.handleFlagUpdate(change);
        }
      }
    } catch (error) {
      this.logger.error('Polling failed', error);
    }
  }

  private async handleStreamMessage(data: string): Promise<void> {
    try {
      const message = JSON.parse(data);
      
      if (message.type === 'flag_update') {
        await this.handleFlagUpdate(message.data);
      }
    } catch (error) {
      this.logger.error('Failed to handle stream message', error);
    }
  }

  private async handleFlagUpdate(update: any): Promise<void> {
    // Clear cache for updated flag
    const keys = await this.cache.keys();
    const pattern = new RegExp(`^${this.config.cache.keyPrefix}${this.config.environment}:${update.flag_key}:`);
    
    for (const key of keys) {
      if (pattern.test(key)) {
        await this.cache.delete(key);
      }
    }
    
    // Emit update event
    this.emit('update', [update.flag_key]);
    if (this.config.events.onUpdate) {
      this.config.events.onUpdate([update.flag_key]);
    }
  }

  private async prefetchFlags(): Promise<void> {
    try {
      let flagKeys = this.config.performance.prefetchFlags;
      
      if (!flagKeys || flagKeys.length === 0) {
        // Fetch all available flags
        const response = await this.httpClient.get('/api/v1/flags', {
          params: { environment: this.config.environment }
        });
        flagKeys = response.data.flags.map((f: Flag) => f.key);
      }
      
      // Batch evaluate all flags
      await this.evaluateBatch(flagKeys);
      
      this.logger.info(`Prefetched ${flagKeys.length} flags`);
    } catch (error) {
      this.logger.error('Failed to prefetch flags', error);
    }
  }

  private async loadOfflineFlags(): Promise<void> {
    if (typeof window === 'undefined' || !window.localStorage) return;
    
    try {
      const stored = localStorage.getItem(this.config.offline.storageKey);
      if (stored) {
        const flags = JSON.parse(stored);
        this.config.offline.defaultFlags = { ...flags, ...this.config.offline.defaultFlags };
        this.logger.info('Loaded offline flags from storage');
      }
    } catch (error) {
      this.logger.error('Failed to load offline flags', error);
    }
  }

  private async saveOfflineFlags(): Promise<void> {
    if (typeof window === 'undefined' || !window.localStorage) return;
    
    try {
      // Get all cached flags
      const keys = await this.cache.keys();
      const flags: Record<string, FlagValue> = {};
      
      for (const key of keys) {
        const match = key.match(new RegExp(`^${this.config.cache.keyPrefix}${this.config.environment}:([^:]+):`));
        if (match) {
          const flagKey = match[1];
          if (!flags[flagKey]) {
            const value = await this.cache.get(key);
            if (value !== null) {
              flags[flagKey] = value;
            }
          }
        }
      }
      
      localStorage.setItem(this.config.offline.storageKey, JSON.stringify(flags));
      this.logger.info('Saved offline flags to storage');
    } catch (error) {
      this.logger.error('Failed to save offline flags', error);
    }
  }

  private emitEvaluation(flagKey: string, value: FlagValue): void {
    this.emit('evaluation', flagKey, value);
    if (this.config.events.onEvaluation) {
      this.config.events.onEvaluation(flagKey, value);
    }
  }

  private handleError(error: Error): void {
    this.emit('error', error);
    if (this.config.events.onError) {
      this.config.events.onError(error);
    }
  }

  private updateAverageLatency(latency: number): void {
    const total = this.metrics.averageLatency * (this.metrics.evaluations - 1) + latency;
    this.metrics.averageLatency = total / this.metrics.evaluations;
  }
}