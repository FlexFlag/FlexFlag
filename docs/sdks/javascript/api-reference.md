# API Reference

Complete API reference for the FlexFlag JavaScript SDK.

## FlexFlagClient

### Constructor

```javascript
new FlexFlagClient(config: FlexFlagConfig)
```

Creates a new FlexFlag client instance.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| config | `FlexFlagConfig` | Yes | Configuration object for the client |

### Methods

#### initialize()

```javascript
async initialize(): Promise<void>
```

Initializes the client and establishes connection to FlexFlag server.

**Example:**
```javascript
const client = new FlexFlagClient(config);
await client.initialize();
```

#### evaluate()

```javascript
async evaluate(flagKey: string, defaultValue?: any, context?: EvaluationContext): Promise<any>
```

Evaluates a single feature flag.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| flagKey | `string` | Yes | The key of the flag to evaluate |
| defaultValue | `any` | No | Default value if flag cannot be evaluated |
| context | `EvaluationContext` | No | User context for evaluation |

**Returns:** Promise resolving to the flag value

**Example:**
```javascript
const value = await client.evaluate('new-feature', false, {
  userId: 'user-123',
  attributes: { plan: 'premium' }
});
```

#### evaluateBatch()

```javascript
async evaluateBatch(flagKeys: string[], context?: EvaluationContext): Promise<Record<string, any>>
```

Evaluates multiple feature flags in a single request.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| flagKeys | `string[]` | Yes | Array of flag keys to evaluate |
| context | `EvaluationContext` | No | User context for evaluation |

**Returns:** Promise resolving to an object with flag keys as properties

**Example:**
```javascript
const flags = await client.evaluateBatch([
  'new-feature',
  'dark-mode',
  'beta-access'
], context);

console.log(flags['new-feature']); // true/false
```

#### evaluateWithDetails()

```javascript
async evaluateWithDetails(flagKey: string, context?: EvaluationContext): Promise<EvaluationResult>
```

Evaluates a flag and returns detailed information about the evaluation.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| flagKey | `string` | Yes | The key of the flag to evaluate |
| context | `EvaluationContext` | No | User context for evaluation |

**Returns:** Promise resolving to detailed evaluation result

**Example:**
```javascript
const result = await client.evaluateWithDetails('feature-flag', context);
console.log(result.value);      // Flag value
console.log(result.reason);     // Why this value was returned
console.log(result.variation);  // Which variation was selected
console.log(result.metadata);   // Additional metadata
```

#### getVariation()

```javascript
async getVariation(flagKey: string, context?: EvaluationContext): Promise<string | null>
```

Gets the variation name for A/B testing scenarios.

#### setContext()

```javascript
setContext(context: EvaluationContext): void
```

Sets the default context for all flag evaluations.

**Example:**
```javascript
client.setContext({
  userId: 'user-123',
  attributes: {
    plan: 'premium',
    country: 'US'
  }
});
```

#### updateContext()

```javascript
updateContext(updates: Partial<EvaluationContext>): void
```

Updates the default context by merging with existing context.

#### clearCache()

```javascript
async clearCache(): Promise<void>
```

Clears all cached flag values.

#### getMetrics()

```javascript
getMetrics(): ClientMetrics
```

Returns SDK performance metrics.

**Example:**
```javascript
const metrics = client.getMetrics();
console.log(metrics.evaluations);     // Total evaluations
console.log(metrics.cacheHits);       // Cache hit count
console.log(metrics.cacheMisses);     // Cache miss count
console.log(metrics.averageLatency);  // Average response time
```

#### resetMetrics()

```javascript
resetMetrics(): void
```

Resets all SDK metrics to zero.

#### isReady()

```javascript
isReady(): boolean
```

Returns true if the SDK is initialized and ready to use.

#### waitForReady()

```javascript
async waitForReady(timeout?: number): Promise<void>
```

Waits for the SDK to be ready, with optional timeout.

#### close()

```javascript
async close(): Promise<void>
```

Closes the client, cleans up resources, and saves offline flags.

### Events

The FlexFlagClient extends EventEmitter and emits the following events:

#### ready

Emitted when the client is initialized and ready to use.

```javascript
client.on('ready', () => {
  console.log('FlexFlag client is ready');
});
```

#### update

Emitted when flags are updated in real-time.

```javascript
client.on('update', (updatedFlags) => {
  console.log('Flags updated:', updatedFlags);
});
```

#### error

Emitted when an error occurs.

```javascript
client.on('error', (error) => {
  console.error('FlexFlag error:', error);
});
```

#### evaluation

Emitted after each flag evaluation.

```javascript
client.on('evaluation', (flagKey, value) => {
  console.log(`Flag ${flagKey} evaluated to:`, value);
});
```

## Configuration

### FlexFlagConfig

```typescript
interface FlexFlagConfig {
  apiKey: string;
  baseUrl: string;
  environment: string;
  cache?: CacheConfig;
  connection?: ConnectionConfig;
  offline?: OfflineConfig;
  performance?: PerformanceConfig;
  logging?: LoggingConfig;
  events?: EventConfig;
}
```

#### Required Properties

| Property | Type | Description |
|----------|------|-------------|
| apiKey | `string` | Your FlexFlag API key |
| baseUrl | `string` | FlexFlag server URL |
| environment | `string` | Environment name (production, staging, etc.) |

#### Optional Properties

| Property | Type | Description |
|----------|------|-------------|
| cache | `CacheConfig` | Cache configuration |
| connection | `ConnectionConfig` | Connection settings |
| offline | `OfflineConfig` | Offline mode settings |
| performance | `PerformanceConfig` | Performance optimization settings |
| logging | `LoggingConfig` | Logging configuration |
| events | `EventConfig` | Event callback configuration |

### CacheConfig

```typescript
interface CacheConfig {
  enabled?: boolean;
  storage?: CacheStorage;
  ttl?: number;
  maxSize?: number;
  keyPrefix?: string;
  compression?: boolean;
  provider?: CacheProvider;
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| enabled | `boolean` | `true` | Enable/disable caching |
| storage | `'memory' \| 'localStorage' \| 'sessionStorage'` | `'memory'` | Cache storage type |
| ttl | `number` | `300000` | Cache TTL in milliseconds |
| maxSize | `number` | `1000` | Maximum cached items |
| keyPrefix | `string` | `'flexflag:'` | Cache key prefix |
| compression | `boolean` | `false` | Enable compression for large values |
| provider | `CacheProvider` | `undefined` | Custom cache provider |

### ConnectionConfig

```typescript
interface ConnectionConfig {
  mode?: ConnectionMode;
  pollingInterval?: number;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
  exponentialBackoff?: boolean;
  headers?: Record<string, string>;
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| mode | `'streaming' \| 'polling' \| 'offline'` | `'streaming'` | Connection mode |
| pollingInterval | `number` | `30000` | Polling interval in ms |
| timeout | `number` | `5000` | Request timeout in ms |
| retryAttempts | `number` | `3` | Number of retry attempts |
| retryDelay | `number` | `1000` | Delay between retries in ms |
| exponentialBackoff | `boolean` | `true` | Use exponential backoff |
| headers | `Record<string, string>` | `{}` | Additional HTTP headers |

### OfflineConfig

```typescript
interface OfflineConfig {
  enabled?: boolean;
  persistence?: boolean;
  storageKey?: string;
  defaultFlags?: Record<string, any>;
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| enabled | `boolean` | `true` | Enable offline mode |
| persistence | `boolean` | `true` | Persist flags to storage |
| storageKey | `string` | `'flexflag_offline'` | Storage key for offline flags |
| defaultFlags | `Record<string, any>` | `{}` | Default flag values when offline |

### PerformanceConfig

```typescript
interface PerformanceConfig {
  evaluationMode?: EvaluationMode;
  batchRequests?: boolean;
  batchInterval?: number;
  compressionEnabled?: boolean;
  prefetch?: boolean;
  prefetchFlags?: string[];
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| evaluationMode | `'cached' \| 'fresh'` | `'cached'` | Evaluation strategy |
| batchRequests | `boolean` | `true` | Enable request batching |
| batchInterval | `number` | `100` | Batch interval in ms |
| compressionEnabled | `boolean` | `true` | Enable response compression |
| prefetch | `boolean` | `true` | Enable flag prefetching |
| prefetchFlags | `string[]` | `undefined` | Specific flags to prefetch |

### LoggingConfig

```typescript
interface LoggingConfig {
  level?: LogLevel;
  logger?: Logger;
  timestamps?: boolean;
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| level | `'debug' \| 'info' \| 'warn' \| 'error' \| 'none'` | `'warn'` | Log level |
| logger | `Logger` | `DefaultLogger` | Custom logger implementation |
| timestamps | `boolean` | `true` | Include timestamps in logs |

### EventConfig

```typescript
interface EventConfig {
  onReady?: () => void;
  onUpdate?: (flags: string[]) => void;
  onError?: (error: Error) => void;
  onEvaluation?: (flagKey: string, value: any) => void;
  onCacheHit?: (flagKey: string) => void;
  onCacheMiss?: (flagKey: string) => void;
}
```

## Types

### EvaluationContext

```typescript
interface EvaluationContext {
  userId?: string;
  attributes?: Record<string, any>;
}
```

### EvaluationResult

```typescript
interface EvaluationResult {
  value: any;
  reason: string;
  variation?: string;
  metadata: {
    timestamp: Date;
    cacheHit: boolean;
    evaluationTime: number;
    source: string;
  };
}
```

### ClientMetrics

```typescript
interface ClientMetrics {
  evaluations: number;
  cacheHits: number;
  cacheMisses: number;
  errors: number;
  networkRequests: number;
  averageLatency: number;
}
```

## React Hooks

### useFeatureFlag

```typescript
function useFeatureFlag(
  flagKey: string, 
  defaultValue?: any, 
  context?: EvaluationContext
): {
  value: any;
  loading: boolean;
  error: Error | null;
  reload: () => void;
}
```

### useFlexFlagClient

```typescript
function useFlexFlagClient(): FlexFlagClient
```

Returns the FlexFlag client from context.

## React Components

### FlexFlagProvider

```typescript
interface FlexFlagProviderProps {
  client: FlexFlagClient;
  context?: EvaluationContext;
  children: React.ReactNode;
}

function FlexFlagProvider(props: FlexFlagProviderProps): JSX.Element
```

Provides FlexFlag client to child components.

## Vue Composables

### useFeatureFlagVue

```typescript
function useFeatureFlagVue(
  flagKey: string,
  defaultValue?: any,
  context?: EvaluationContext | ComputedRef<EvaluationContext>
): {
  value: Ref<any>;
  loading: Ref<boolean>;
  error: Ref<Error | null>;
  reload: () => void;
}
```

### useFlexFlagClient

```typescript
function useFlexFlagClient(): FlexFlagClient
```

Returns the FlexFlag client from Vue's provide/inject.

## Cache Providers

### MemoryCache

In-memory LRU cache implementation.

```javascript
const cache = new MemoryCache({
  maxSize: 1000,
  ttl: 300000
});
```

### LocalStorageCache

Browser localStorage/sessionStorage cache implementation.

```javascript
const cache = new LocalStorageCache({
  storage: 'localStorage', // or 'sessionStorage'
  maxSize: 1000,
  ttl: 300000,
  compression: true
});
```

### Custom Cache Provider

Implement your own cache provider:

```typescript
interface CacheProvider {
  get(key: string): Promise<any>;
  set(key: string, value: any, ttl?: number): Promise<void>;
  delete(key: string): Promise<void>;
  clear(): Promise<void>;
  has(key: string): Promise<boolean>;
  size(): Promise<number>;
  keys(): Promise<string[]>;
}

class CustomCache implements CacheProvider {
  // Implement all required methods
}

const client = new FlexFlagClient({
  // ... other config
  cache: {
    provider: new CustomCache()
  }
});
```

## Error Handling

### Error Types

The SDK can throw the following types of errors:

- **Configuration Error**: Invalid configuration
- **Network Error**: Connection or HTTP errors
- **Authentication Error**: Invalid API key
- **Evaluation Error**: Flag evaluation failures

### Best Practices

```javascript
try {
  const value = await client.evaluate('flag-key', false);
  // Use the value
} catch (error) {
  console.error('Flag evaluation failed:', error);
  // Use default value or fallback behavior
}

// Or with error handling in events
client.on('error', (error) => {
  console.error('FlexFlag SDK error:', error);
  // Handle error appropriately
});
```

## Environment Variables

Common environment variable patterns:

```bash
# Development
FLEXFLAG_API_KEY=your-dev-api-key
FLEXFLAG_BASE_URL=http://localhost:8080
FLEXFLAG_ENVIRONMENT=development

# Production
FLEXFLAG_API_KEY=your-prod-api-key
FLEXFLAG_BASE_URL=https://api.yourapp.com
FLEXFLAG_ENVIRONMENT=production
```