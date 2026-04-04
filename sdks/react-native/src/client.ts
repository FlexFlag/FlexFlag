import axios, { AxiosInstance } from 'axios';
import {
  FlexFlagConfig,
  FlagValue,
  EvaluationContext,
  EvaluationResult,
  ConfigResult,
  CachedEvaluation,
  PersistedState,
} from './types';

const DEFAULT_POLLING_INTERVAL = 30_000;
const DEFAULT_TIMEOUT = 5_000;
const DEFAULT_CACHE_TTL = 300_000; // 5 minutes
const DEFAULT_PERSISTENCE_KEY = 'flexflag_cache';

type EventType = 'ready' | 'update' | 'error';
type Listener = (...args: unknown[]) => void;

/** Detect platform automatically from React Native's Platform API. Returns undefined outside RN. */
function detectPlatform(): 'ios' | 'android' | undefined {
  try {
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { Platform } = require('react-native');
    if (Platform.OS === 'ios') return 'ios';
    if (Platform.OS === 'android') return 'android';
  } catch {
    // Not in React Native — web or Node environment
  }
  return undefined;
}

export class FlexFlagClient {
  private readonly config: Required<FlexFlagConfig>;
  private readonly http: AxiosInstance;
  private readonly platform: 'ios' | 'android' | undefined;
  private cache: Map<string, CachedEvaluation> = new Map();
  private configCache: ConfigResult | null = null;
  private pollingTimer: ReturnType<typeof setInterval> | null = null;
  private listeners: Map<EventType, Set<Listener>> = new Map();
  private storage: AsyncStorageLike | null = null;
  private ready = false;
  private appStateSubscription: unknown = null;

  constructor(config: FlexFlagConfig) {
    this.config = {
      baseUrl: 'http://localhost:8080',
      environment: 'production',
      pollingInterval: DEFAULT_POLLING_INTERVAL,
      timeout: DEFAULT_TIMEOUT,
      offlineDefaults: {},
      persistenceKey: DEFAULT_PERSISTENCE_KEY,
      cacheTtl: DEFAULT_CACHE_TTL,
      ...config,
    };

    // Auto-detect platform once at construction time.
    // Callers can still override via context.platform in individual calls.
    this.platform = detectPlatform();

    this.http = axios.create({
      baseURL: `${this.config.baseUrl}/api/v1`,
      timeout: this.config.timeout,
      headers: {
        'X-API-Key': this.config.apiKey,
        'Content-Type': 'application/json',
      },
    });
  }

  /**
   * Initialize the client. Must be called before evaluating flags.
   * Loads persisted cache, fetches fresh flags, and starts polling.
   */
  async initialize(storage?: AsyncStorageLike): Promise<void> {
    if (storage) {
      this.storage = storage;
      await this.loadFromStorage();
    }

    this.attachAppStateListener();

    try {
      await this.fetchAllFlags();
      this.ready = true;
      this.emit('ready');
    } catch {
      this.ready = true;
      this.emit('ready');
    }

    if (this.config.pollingInterval > 0) {
      this.startPolling();
    }
  }

  /**
   * Evaluate a flag and return the full result.
   * Platform is automatically included — no need to specify it manually.
   */
  async evaluate(
    flagKey: string,
    context?: EvaluationContext,
    defaultValue: FlagValue = null
  ): Promise<EvaluationResult> {
    const cacheKey = this.buildCacheKey(flagKey, context);
    const cached = this.cache.get(cacheKey);

    if (cached && Date.now() - cached.cachedAt < cached.ttl) {
      return { value: cached.value, variation: cached.variation, reason: cached.reason, default: false };
    }

    try {
      const response = await this.http.post('/evaluate', {
        flag_key: flagKey,
        user_id: context?.userId,
        user_key: context?.userKey,
        // Auto-detected platform; context.platform overrides if explicitly provided
        platform: context?.platform ?? this.platform,
        attributes: context?.attributes,
      }, {
        params: { environment: this.config.environment },
      });

      const result: EvaluationResult = {
        value: response.data.value,
        variation: response.data.variation,
        reason: response.data.reason,
        default: response.data.default ?? false,
      };

      this.cache.set(cacheKey, {
        value: result.value,
        variation: result.variation,
        reason: result.reason,
        cachedAt: Date.now(),
        ttl: this.config.cacheTtl,
      });

      return result;
    } catch {
      const offline = this.config.offlineDefaults[flagKey] ?? defaultValue;
      return { value: offline, reason: 'OFFLINE', default: true };
    }
  }

  async evaluateBoolean(flagKey: string, defaultValue = false, context?: EvaluationContext): Promise<boolean> {
    const result = await this.evaluate(flagKey, context, defaultValue);
    return typeof result.value === 'boolean' ? result.value : defaultValue;
  }

  async evaluateString(flagKey: string, defaultValue = '', context?: EvaluationContext): Promise<string> {
    const result = await this.evaluate(flagKey, context, defaultValue);
    return typeof result.value === 'string' ? result.value : defaultValue;
  }

  async evaluateNumber(flagKey: string, defaultValue = 0, context?: EvaluationContext): Promise<number> {
    const result = await this.evaluate(flagKey, context, defaultValue);
    return typeof result.value === 'number' ? result.value : defaultValue;
  }

  async evaluateJSON<T>(flagKey: string, defaultValue: T, context?: EvaluationContext): Promise<T> {
    const result = await this.evaluate(flagKey, context, defaultValue as FlagValue);
    if (result.value !== null && typeof result.value === 'object') {
      return result.value as unknown as T;
    }
    return defaultValue;
  }

  /**
   * Fetch all remote config values for this environment.
   * Platform is auto-detected — iOS clients get iOS values, Android gets Android values.
   * Call this at app startup to warm the config cache.
   */
  async getConfig(context?: EvaluationContext): Promise<ConfigResult> {
    try {
      const response = await this.http.post('/config/evaluate', {
        user_id: context?.userId,
        user_key: context?.userKey,
        platform: context?.platform ?? this.platform,
        attributes: context?.attributes,
      }, {
        params: { environment: this.config.environment },
      });

      this.configCache = response.data as ConfigResult;
      await this.persistToStorage();
      return this.configCache;
    } catch {
      if (this.configCache) return this.configCache;
      return { config: {}, metadata: {}, environment: this.config.environment, timestamp: new Date().toISOString() };
    }
  }

  /**
   * Get a single config value from cache (synchronous, no network call).
   * Call getConfig() first to populate the config cache.
   */
  getConfigValue<T extends FlagValue>(key: string, defaultValue: T): T {
    const val = this.configCache?.config[key];
    if (val === undefined || val === null) return defaultValue;
    return val as T;
  }

  /** Returns the auto-detected platform ("ios" | "android" | undefined). */
  getPlatform(): 'ios' | 'android' | undefined {
    return this.platform;
  }

  setContext(_context: EvaluationContext): void {
    this.cache.clear();
  }

  on(event: EventType, listener: Listener): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(listener);
  }

  off(event: EventType, listener: Listener): void {
    this.listeners.get(event)?.delete(listener);
  }

  isReady(): boolean {
    return this.ready;
  }

  async close(): Promise<void> {
    this.stopPolling();
    this.detachAppStateListener();
    await this.persistToStorage();
  }

  // ─── Private ────────────────────────────────────────────────────────────────

  private async fetchAllFlags(): Promise<void> {
    const response = await this.http.get('/flags', {
      params: { environment: this.config.environment },
    });

    const flags: Array<{
      key: string;
      default: FlagValue;
      type: string;
      enabled: boolean;
    }> = response.data?.flags ?? [];

    for (const flag of flags) {
      if (!flag.enabled) continue;
      const cacheKey = this.buildCacheKey(flag.key, undefined);
      if (!this.cache.has(cacheKey)) {
        this.cache.set(cacheKey, {
          value: flag.default,
          reason: 'PRELOADED',
          cachedAt: Date.now(),
          ttl: this.config.cacheTtl,
        });
      }
    }

    await this.persistToStorage();
  }

  private startPolling(): void {
    this.pollingTimer = setInterval(async () => {
      try {
        const hadCache = this.cache.size > 0;
        this.cache.clear();
        await this.fetchAllFlags();
        if (hadCache) this.emit('update', 'all');
      } catch {
        // Polling failure is silent — serve from cache
      }
    }, this.config.pollingInterval);
  }

  private stopPolling(): void {
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer);
      this.pollingTimer = null;
    }
  }

  private attachAppStateListener(): void {
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { AppState } = require('react-native');
      this.appStateSubscription = AppState.addEventListener('change', (state: string) => {
        if (state === 'active') {
          this.cache.clear();
          this.fetchAllFlags().catch(() => {});
        }
      });
    } catch {
      // Not in React Native — skip
    }
  }

  private detachAppStateListener(): void {
    if (this.appStateSubscription && typeof (this.appStateSubscription as { remove?: () => void }).remove === 'function') {
      (this.appStateSubscription as { remove: () => void }).remove();
    }
    this.appStateSubscription = null;
  }

  private async loadFromStorage(): Promise<void> {
    if (!this.storage) return;
    try {
      const raw = await this.storage.getItem(this.config.persistenceKey);
      if (!raw) return;
      const state: PersistedState = JSON.parse(raw);
      const now = Date.now();
      for (const [key, entry] of Object.entries(state.evaluations ?? {})) {
        if (now - entry.cachedAt < entry.ttl) {
          this.cache.set(key, entry);
        }
      }
      if (state.config) {
        this.configCache = state.config;
      }
    } catch {
      // Corrupted storage — start fresh
    }
  }

  private async persistToStorage(): Promise<void> {
    if (!this.storage) return;
    try {
      const evaluations: Record<string, CachedEvaluation> = {};
      for (const [key, val] of this.cache.entries()) {
        evaluations[key] = val;
      }
      const state: PersistedState = {
        evaluations,
        config: this.configCache,
        savedAt: Date.now(),
      };
      await this.storage.setItem(this.config.persistenceKey, JSON.stringify(state));
    } catch {
      // Storage write failure is non-fatal
    }
  }

  private buildCacheKey(flagKey: string, context?: EvaluationContext): string {
    const userId = context?.userId ?? '_';
    const platform = context?.platform ?? this.platform ?? '_';
    return `${this.config.environment}:${flagKey}:${userId}:${platform}`;
  }

  private emit(event: EventType, ...args: unknown[]): void {
    this.listeners.get(event)?.forEach(l => l(...args));
  }
}

/** Minimal AsyncStorage interface — matches @react-native-async-storage/async-storage */
export interface AsyncStorageLike {
  getItem(key: string): Promise<string | null>;
  setItem(key: string, value: string): Promise<void>;
  removeItem(key: string): Promise<void>;
}
