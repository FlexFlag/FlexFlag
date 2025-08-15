# FlexFlag SDKs Documentation

FlexFlag provides intelligent, high-performance SDKs with local caching for all major programming languages and frameworks.

## 🚀 Quick Start Guide

### JavaScript/TypeScript SDK

#### Installation
```bash
npm install @flexflag/client
# or
yarn add @flexflag/client
```

#### Basic Usage
```javascript
import { FlexFlagClient } from '@flexflag/client';

const client = new FlexFlagClient({
  apiKey: 'ff_production_your_api_key_here',
  baseUrl: 'https://api.flexflag.io', // or your edge server URL
  cache: {
    enabled: true,
    ttl: 300000, // 5 minutes
    maxSize: 1000
  }
});

// Evaluate a feature flag
const isEnabled = await client.evaluate('new-feature', undefined, false);

// Batch evaluation (more efficient)
const flags = await client.evaluateBatch(['feature-1', 'feature-2']);

// A/B testing
const variation = await client.getVariation('checkout-test');
```

#### React Integration
```jsx
import { FlexFlagProvider, useFeatureFlag, FeatureGate } from '@flexflag/client/react';

function App() {
  return (
    <FlexFlagProvider
      config={{ apiKey: 'your_api_key' }}
      context={{ userId: 'user123', attributes: { plan: 'premium' } }}
    >
      <MyComponent />
    </FlexFlagProvider>
  );
}

function MyComponent() {
  const { value: showNewUI, loading } = useFeatureFlag('new-ui', false);
  
  return (
    <div>
      <FeatureGate flagKey="beta-features" fallback={<div>Coming soon!</div>}>
        <BetaFeatures />
      </FeatureGate>
      
      {showNewUI && <NewUIComponents />}
    </div>
  );
}
```

#### Vue 3 Integration
```vue
<template>
  <div>
    <div v-if="enabled.value">🚀 New feature is enabled!</div>
    <div v-feature-flag="'beta-features'">Beta content</div>
  </div>
</template>

<script setup>
import { useBooleanFlag } from '@flexflag/client/vue';

const { enabled, loading, error } = useBooleanFlag('new-feature', false);
</script>
```

### Python SDK

#### Installation
```bash
pip install flexflag
# With extras for specific frameworks
pip install flexflag[django,redis,async]
```

#### Basic Usage
```python
import asyncio
from flexflag import FlexFlagClient, FlexFlagConfig, EvaluationContext

async def main():
    config = FlexFlagConfig(
        api_key="ff_production_your_api_key_here",
        base_url="https://api.flexflag.io",
        cache=CacheConfig(enabled=True, ttl=300)
    )
    
    client = FlexFlagClient(config)
    await client.wait_for_ready()
    
    # Set user context
    context = EvaluationContext(
        user_id="user123",
        attributes={"plan": "premium", "region": "us-east"}
    )
    client.set_context(context)
    
    # Evaluate flags
    is_enabled = await client.evaluate("new-feature", default_value=False)
    theme = await client.evaluate("ui-theme", default_value="light")
    
    # Batch evaluation
    flags = await client.evaluate_batch(["feature-1", "feature-2", "feature-3"])
    
    await client.close()

asyncio.run(main())
```

#### Django Integration
```python
# settings.py
FLEXFLAG_CONFIG = {
    'api_key': 'ff_production_your_api_key_here',
    'base_url': 'https://api.flexflag.io',
    'cache': {'enabled': True, 'ttl': 300}
}

# views.py
from flexflag.integrations.django import get_flexflag_client

async def my_view(request):
    client = get_flexflag_client()
    context = EvaluationContext(
        user_id=str(request.user.id) if request.user.is_authenticated else None,
        attributes={
            'is_staff': request.user.is_staff,
            'plan': getattr(request.user, 'plan', 'free')
        }
    )
    
    show_new_dashboard = await client.evaluate(
        'new-dashboard', 
        context=context, 
        default_value=False
    )
    
    return render(request, 'dashboard.html', {
        'show_new_dashboard': show_new_dashboard
    })
```

#### FastAPI Integration
```python
from fastapi import FastAPI, Depends
from flexflag import FlexFlagClient, FlexFlagConfig
from flexflag.integrations.fastapi import FlexFlagDependency

app = FastAPI()

# Initialize FlexFlag
flexflag_config = FlexFlagConfig(api_key="your_api_key")
flexflag = FlexFlagDependency(flexflag_config)

@app.get("/api/features")
async def get_features(client: FlexFlagClient = Depends(flexflag)):
    features = await client.evaluate_batch([
        'new-api-version',
        'rate-limiting',
        'advanced-search'
    ])
    return features
```

## 📊 Caching Strategies

### Memory Cache (Fastest)
```javascript
// JavaScript
const client = new FlexFlagClient({
  cache: {
    storage: 'memory',
    ttl: 300000, // 5 minutes
    maxSize: 1000
  }
});
```

```python
# Python
from flexflag import MemoryCache, CacheConfig

config = FlexFlagConfig(
    cache=CacheConfig(
        storage="memory",
        ttl=300,  # 5 minutes
        max_size=1000
    )
)
```

### Persistent Cache (Offline Support)
```javascript
// Browser localStorage
const client = new FlexFlagClient({
  cache: {
    storage: 'localStorage',
    ttl: 600000, // 10 minutes
    compression: true
  }
});
```

```python
# Python disk cache
from flexflag import DiskCache

config = FlexFlagConfig(
    cache=CacheConfig(
        storage="disk",
        ttl=600,
        storage_file="/tmp/flexflag_cache.db"
    )
)
```

### Redis Cache (Distributed)
```python
# Python Redis cache
from flexflag import RedisCache

redis_cache = RedisCache(
    host='localhost',
    port=6379,
    db=0,
    ttl=300
)

config = FlexFlagConfig(
    cache=CacheConfig(storage="custom"),
    custom_cache=redis_cache
)
```

### Tiered Cache (Best of Both)
```javascript
import { TieredCache, MemoryCache, LocalStorageCache } from '@flexflag/client';

const tieredCache = new TieredCache([
  new MemoryCache({ ttl: 60000, maxSize: 100 }),      // L1: Memory (1 min)
  new LocalStorageCache({ ttl: 300000, maxSize: 500 }) // L2: Storage (5 min)
]);

const client = new FlexFlagClient({
  cache: { provider: tieredCache }
});
```

## 🎯 Advanced Usage Patterns

### Context-Aware Evaluation
```javascript
// Dynamic context based on user state
const getContextForUser = (user) => ({
  userId: user.id,
  attributes: {
    plan: user.subscription?.plan || 'free',
    region: user.profile?.region || 'us-east',
    signupDate: user.createdAt,
    isVerified: user.emailVerified,
    experimentGroup: user.experiments?.group || 'control'
  },
  device: {
    type: user.lastKnownDevice?.type || 'desktop',
    os: user.lastKnownDevice?.os,
    browser: user.lastKnownDevice?.browser
  }
});

const context = getContextForUser(currentUser);
const features = await client.evaluateBatch([
  'premium-dashboard',
  'advanced-search',
  'beta-features'
], context);
```

### A/B Testing & Experiments
```javascript
// Multi-variate testing
const checkoutVariation = await client.getVariation('checkout-flow');

switch (checkoutVariation) {
  case 'streamlined':
    return <StreamlinedCheckout />;
  case 'detailed':
    return <DetailedCheckout />;
  case 'premium':
    return <PremiumCheckout />;
  default:
    return <StandardCheckout />;
}

// Track experiment exposure
client.on('evaluation', (flagKey, value) => {
  if (flagKey === 'checkout-flow') {
    analytics.track('experiment_exposure', {
      experiment: flagKey,
      variation: value,
      userId: context.userId
    });
  }
});
```

### Performance Optimization
```javascript
// Prefetch critical flags on app startup
const client = new FlexFlagClient({
  performance: {
    prefetch: true,
    prefetchFlags: [
      'critical-feature',
      'ui-theme',
      'navigation-style',
      'payment-methods'
    ],
    batchRequests: true,
    compressionEnabled: true
  }
});

// Lazy loading for non-critical flags
const nonCriticalFlags = [
  'experimental-feature-1',
  'experimental-feature-2',
  'admin-tools'
];

// Load these only when needed
const loadExperimentalFeatures = async () => {
  return await client.evaluateBatch(nonCriticalFlags);
};
```

### Error Handling & Resilience
```javascript
const client = new FlexFlagClient({
  offline: {
    enabled: true,
    defaultFlags: {
      'critical-feature': true,
      'payment-enabled': true,
      'maintenance-mode': false
    }
  },
  connection: {
    retryAttempts: 3,
    retryDelay: 1000,
    exponentialBackoff: true
  },
  events: {
    onError: (error) => {
      console.error('FlexFlag error:', error);
      // Report to error tracking service
      errorTracker.captureException(error);
    },
    onOffline: () => {
      // Switch to degraded mode
      showNotification('Using offline feature flags');
    }
  }
});

// Graceful degradation
const evaluateWithFallback = async (flagKey, fallbackValue, context) => {
  try {
    return await client.evaluate(flagKey, context, fallbackValue);
  } catch (error) {
    console.warn(`Flag evaluation failed for ${flagKey}, using fallback`);
    return fallbackValue;
  }
};
```

## 📈 Monitoring & Analytics

### Performance Metrics
```javascript
// Monitor SDK performance
setInterval(() => {
  const metrics = client.getMetrics();
  
  console.log('FlexFlag Performance:', {
    evaluations: metrics.evaluations,
    cacheHitRate: `${(metrics.cacheHits / metrics.evaluations * 100).toFixed(1)}%`,
    averageLatency: `${metrics.averageLatency.toFixed(2)}ms`,
    networkRequests: metrics.networkRequests
  });
  
  // Send to monitoring service
  monitoring.histogram('flexflag.evaluation.latency', metrics.averageLatency);
  monitoring.gauge('flexflag.cache.hit_rate', metrics.cacheHits / metrics.evaluations);
}, 60000); // Every minute
```

### Event Tracking
```javascript
const client = new FlexFlagClient({
  events: {
    onCacheHit: (flagKey) => {
      metrics.increment('flexflag.cache.hit', { flag: flagKey });
    },
    onCacheMiss: (flagKey) => {
      metrics.increment('flexflag.cache.miss', { flag: flagKey });
    },
    onEvaluation: (flagKey, value) => {
      // Track feature usage
      analytics.track('feature_flag_evaluated', {
        flag: flagKey,
        value: value,
        timestamp: new Date().toISOString()
      });
    },
    onUpdate: (updatedFlags) => {
      console.log('Flags updated in real-time:', updatedFlags);
      // Refresh UI components that depend on these flags
      updatedFlags.forEach(flag => {
        eventBus.emit(`flag-updated:${flag}`);
      });
    }
  }
});
```

## 🔧 Configuration Reference

### Cache Configuration
```typescript
interface CacheConfig {
  enabled: boolean;           // Enable/disable caching
  ttl: number;               // Time-to-live in milliseconds
  maxSize: number;           // Maximum number of cached flags
  storage: 'memory' | 'localStorage' | 'sessionStorage' | 'custom';
  keyPrefix: string;         // Cache key prefix
  compression: boolean;      // Enable compression for large values
}
```

### Connection Configuration
```typescript
interface ConnectionConfig {
  mode: 'streaming' | 'polling' | 'offline';
  pollingInterval: number;   // Polling interval in milliseconds
  timeout: number;          // Request timeout in milliseconds
  retryAttempts: number;    // Number of retry attempts
  retryDelay: number;       // Delay between retries
  exponentialBackoff: boolean; // Use exponential backoff
}
```

### Performance Configuration
```typescript
interface PerformanceConfig {
  evaluationMode: 'cached' | 'always-fetch' | 'lazy';
  batchRequests: boolean;    // Enable request batching
  batchInterval: number;     // Batch interval in milliseconds
  prefetch: boolean;         // Prefetch flags on initialization
  prefetchFlags: string[];   // Specific flags to prefetch
}
```

## 🚀 Best Practices

### 1. Optimal Cache Configuration
- **Memory cache**: For frequently accessed flags (< 1ms latency)
- **Persistent cache**: For offline support and faster startup
- **TTL setting**: Balance between freshness and performance (5-15 minutes typical)
- **Cache size**: Limit based on available memory (1000 flags ≈ 1MB)

### 2. Context Management
- Set context once at app startup for user-specific flags
- Update context only when user state changes significantly
- Use attributes for targeting, not for dynamic values

### 3. Error Handling
- Always provide fallback values for critical features
- Use offline mode for essential business logic
- Monitor cache hit rates and evaluation latency

### 4. Performance Optimization
- Batch evaluate multiple flags when possible
- Prefetch critical flags on app initialization
- Use streaming mode for real-time updates when available

### 5. Security Considerations
- Never expose API keys in client-side code
- Use environment-specific API keys
- Implement rate limiting for high-traffic applications

## 📚 Framework Integrations

### Express.js Middleware
```javascript
const flexflag = require('@flexflag/client');

const flexflagMiddleware = (config) => {
  const client = new flexflag.FlexFlagClient(config);
  
  return async (req, res, next) => {
    req.flexflag = client;
    req.flagContext = {
      userId: req.user?.id,
      attributes: {
        plan: req.user?.plan,
        role: req.user?.role
      }
    };
    next();
  };
};

app.use(flexflagMiddleware({ apiKey: process.env.FLEXFLAG_API_KEY }));
```

### Django Middleware
```python
# middleware.py
from flexflag.integrations.django import FlexFlagMiddleware

# settings.py
MIDDLEWARE = [
    # ... other middleware
    'your_app.middleware.FlexFlagMiddleware',
]

FLEXFLAG_CONFIG = {
    'api_key': os.environ['FLEXFLAG_API_KEY'],
    'cache': {'enabled': True}
}

# In views
def my_view(request):
    show_feature = await request.flexflag.evaluate(
        'new-feature', 
        context=request.flag_context
    )
```

## 🔍 Troubleshooting

### Common Issues

**1. High Cache Miss Rate**
```javascript
// Check cache configuration
const metrics = client.getMetrics();
if (metrics.cacheHits / metrics.evaluations < 0.8) {
  // Increase TTL or check context consistency
  console.log('Consider increasing cache TTL or reducing context variability');
}
```

**2. Slow Evaluation Times**
```javascript
// Monitor evaluation latency
client.on('evaluation', (flagKey, value, metadata) => {
  if (metadata.evaluationTime > 100) { // > 100ms
    console.warn(`Slow evaluation for ${flagKey}: ${metadata.evaluationTime}ms`);
    // Consider using edge servers or increasing cache TTL
  }
});
```

**3. Memory Issues**
```javascript
// Monitor cache size
setInterval(() => {
  const cacheSize = await client.cache.size();
  if (cacheSize > config.cache.maxSize * 0.9) {
    console.warn('Cache approaching size limit');
    // Consider increasing maxSize or reducing TTL
  }
}, 300000); // Check every 5 minutes
```

**4. Network Issues**
```javascript
// Handle network failures gracefully
const client = new FlexFlagClient({
  offline: { enabled: true },
  events: {
    onError: (error) => {
      if (error.code === 'NETWORK_ERROR') {
        // Switch to offline mode
        client.setOfflineMode(true);
      }
    }
  }
});
```

This comprehensive documentation covers all major SDKs, caching strategies, and real-world usage patterns. The SDKs are designed for maximum performance with intelligent caching that developers can easily configure based on their specific needs.