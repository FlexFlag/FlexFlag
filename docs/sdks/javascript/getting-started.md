# FlexFlag JavaScript SDK - Getting Started

The FlexFlag JavaScript SDK provides a high-performance, feature-rich client for integrating feature flags into your JavaScript and TypeScript applications.

## Installation

```bash
npm install flexflag-client
```

## Quick Start

### Basic Usage

```javascript
import { FlexFlagClient } from 'flexflag-client';

// Initialize the client
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://your-flexflag-server.com', // or http://localhost:8080 for local
  environment: 'production'
});

// Wait for initialization
await client.waitForReady();

// Evaluate a feature flag
const isFeatureEnabled = await client.evaluate('new-feature', false);

if (isFeatureEnabled) {
  // Show new feature
  console.log('New feature is enabled!');
} else {
  // Show default experience
  console.log('Using default experience');
}
```

### With Context

```javascript
const context = {
  userId: 'user-123',
  attributes: {
    plan: 'premium',
    country: 'US',
    betaUser: true
  }
};

const flagValue = await client.evaluate('premium-feature', false, context);
```

### TypeScript Usage

```typescript
import { FlexFlagClient, FlexFlagConfig, EvaluationContext } from 'flexflag-client';

const config: FlexFlagConfig = {
  apiKey: 'your-api-key',
  environment: 'production',
  baseUrl: 'https://your-flexflag-server.com'
};

const client = new FlexFlagClient(config);

const context: EvaluationContext = {
  userId: 'user-123',
  attributes: {
    plan: 'premium',
    country: 'US'
  }
};

// Strongly typed flag evaluation
const flagValue: boolean = await client.evaluate('feature-flag', false, context);
```

## Configuration Options

### Basic Configuration

```javascript
const client = new FlexFlagClient({
  apiKey: 'your-api-key',           // Required: Your FlexFlag API key
  baseUrl: 'https://api.example.com', // Required: Your FlexFlag server URL
  environment: 'production',        // Required: Environment (production, staging, development)
});
```

### Advanced Configuration

```javascript
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.example.com',
  environment: 'production',
  
  // Cache configuration
  cache: {
    enabled: true,                  // Enable/disable caching (default: true)
    storage: 'memory',              // 'memory', 'localStorage', 'sessionStorage'
    ttl: 300000,                    // Cache TTL in milliseconds (5 minutes)
    maxSize: 1000,                  // Maximum number of cached flags
    keyPrefix: 'flexflag:'          // Cache key prefix
  },
  
  // Connection settings
  connection: {
    mode: 'streaming',              // 'streaming', 'polling', 'offline'
    pollingInterval: 30000,         // Polling interval in ms (30 seconds)
    timeout: 5000,                  // Request timeout in ms
    retryAttempts: 3,               // Number of retry attempts
    retryDelay: 1000,               // Delay between retries in ms
    exponentialBackoff: true,       // Use exponential backoff for retries
    headers: {}                     // Additional HTTP headers
  },
  
  // Offline support
  offline: {
    enabled: true,                  // Enable offline mode
    persistence: true,              // Persist flags to localStorage
    storageKey: 'flexflag_offline', // localStorage key for offline flags
    defaultFlags: {                 // Default flag values when offline
      'feature-1': true,
      'feature-2': false
    }
  },
  
  // Performance settings
  performance: {
    evaluationMode: 'cached',       // 'cached', 'fresh'
    batchRequests: true,            // Enable request batching
    batchInterval: 100,             // Batch interval in ms
    compressionEnabled: true,       // Enable response compression
    prefetch: true,                 // Prefetch commonly used flags
    prefetchFlags: ['flag1', 'flag2'] // Specific flags to prefetch
  },
  
  // Logging
  logging: {
    level: 'warn',                  // 'debug', 'info', 'warn', 'error', 'none'
    logger: customLogger,           // Custom logger implementation
    timestamps: true                // Include timestamps in logs
  },
  
  // Event callbacks
  events: {
    onReady: () => console.log('SDK ready'),
    onUpdate: (flags) => console.log('Flags updated:', flags),
    onError: (error) => console.error('SDK error:', error),
    onEvaluation: (flag, value) => console.log(`Evaluated ${flag}: ${value}`),
    onCacheHit: (flag) => console.log(`Cache hit for ${flag}`),
    onCacheMiss: (flag) => console.log(`Cache miss for ${flag}`)
  }
});
```

## Next Steps

- [React Integration](./react-integration.md) - Use FlexFlag with React applications
- [Vue Integration](./vue-integration.md) - Use FlexFlag with Vue applications  
- [Advanced Usage](./advanced-usage.md) - Batch evaluation, metrics, and more
- [API Reference](./api-reference.md) - Complete API documentation
- [Performance Guide](./performance.md) - Optimization tips and best practices