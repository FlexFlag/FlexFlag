# FlexFlag JavaScript/TypeScript SDK

High-performance feature flag SDK for JavaScript and TypeScript with local caching, offline support, and framework integrations.

## Features

- 🚀 **High Performance** - Local caching with configurable TTL
- 🔌 **Multiple Connection Modes** - Server-Sent Events (SSE), polling, or offline
- 📦 **Framework Integrations** - React hooks and Vue 3 composables
- 💾 **Offline Support** - Works without network connection using default flags
- 🎯 **Type Safe** - Full TypeScript support with type definitions
- ⚡ **Batch Evaluation** - Evaluate multiple flags in a single request
- 🔄 **Real-time Updates** - Automatic flag updates via Server-Sent Events (SSE)
- 📊 **Metrics Tracking** - Built-in performance and usage metrics

## Installation

### Backend/Node.js
```bash
npm install flexflag-client
```

### With React
```bash
npm install flexflag-client react
```

### With Vue 3
```bash
npm install flexflag-client vue
```

## Quick Start

### Backend/Node.js

```javascript
const { FlexFlagClient } = require('flexflag-client');

const client = new FlexFlagClient({
  apiKey: process.env.FLEXFLAG_API_KEY,
  baseUrl: 'http://localhost:8080',
  environment: 'production'
});

// Boolean flags
const isEnabled = await client.evaluateBoolean('feature-flag', {
  userId: 'user-123',
  attributes: { plan: 'premium' }
}, false);

// String flags
const theme = await client.evaluateString('ui-theme', { userId: 'user-123' }, 'light');

// Number flags
const maxRetries = await client.evaluateNumber('max-retries', { userId: 'user-123' }, 3);

// JSON flags
const config = await client.evaluateJSON('app-config', { userId: 'user-123' }, {});

// Batch evaluation
const flags = await client.evaluateBatch(['feature-a', 'feature-b'], { userId: 'user-123' });
```

### React

**Important**: Import React integration from `flexflag-client/react`

```javascript
import React from 'react';
import { FlexFlagProvider, useBooleanFlag, FeatureGate } from 'flexflag-client/react';

function App() {
  return (
    <FlexFlagProvider
      config={{
        apiKey: process.env.REACT_APP_FLEXFLAG_API_KEY,
        baseUrl: 'http://localhost:8080',
        environment: 'production'
      }}
      context={{ userId: 'user-123', attributes: { plan: 'premium' } }}
    >
      <Dashboard />
    </FlexFlagProvider>
  );
}

function Dashboard() {
  const { enabled, loading } = useBooleanFlag('dark-mode');

  if (loading) return <div>Loading...</div>;

  return (
    <div className={enabled ? 'dark' : 'light'}>
      <h1>Dashboard</h1>

      <FeatureGate flagKey="new-feature" fallback={<OldFeature />}>
        <NewFeature />
      </FeatureGate>
    </div>
  );
}
```

### Vue 3

**Important**: Import Vue integration from `flexflag-client/vue`

```javascript
import { createApp } from 'vue';
import { createFlexFlag } from 'flexflag-client/vue';
import App from './App.vue';

const app = createApp(App);

app.use(createFlexFlag({
  apiKey: import.meta.env.VITE_FLEXFLAG_API_KEY,
  baseUrl: 'http://localhost:8080',
  environment: 'production'
}));

app.mount('#app');
```

```vue
<template>
  <div :class="{ dark: enabled }">
    <h1 v-if="!loading">Dark Mode: {{ enabled }}</h1>
  </div>
</template>

<script setup>
import { useBooleanFlag } from 'flexflag-client/vue';

const { enabled, loading } = useBooleanFlag('dark-mode');
</script>
```

## API Reference

### Client Methods

#### `evaluate(flagKey, context?, defaultValue?)`
Evaluate any flag type.

#### `evaluateBoolean(flagKey, context?, defaultValue?)`
Evaluate a boolean flag.

#### `evaluateString(flagKey, context?, defaultValue?)`
Evaluate a string flag.

#### `evaluateNumber(flagKey, context?, defaultValue?)`
Evaluate a number flag.

#### `evaluateJSON<T>(flagKey, context?, defaultValue?)`
Evaluate a JSON flag with type safety.

#### `evaluateBatch(flagKeys[], context?)`
Evaluate multiple flags at once.

#### `setContext(context)`
Set default context for all evaluations.

#### `clearCache()`
Clear all cached flag values.

#### `getMetrics()`
Get SDK metrics.

#### `close()`
Close SDK connections and cleanup.

## Configuration

```javascript
const client = new FlexFlagClient({
  // Required
  apiKey: 'your_api_key',

  // Optional
  baseUrl: 'http://localhost:8080',
  environment: 'production',

  // Cache configuration
  cache: {
    enabled: true,
    ttl: 300000,         // 5 minutes
    storage: 'memory',   // 'memory' | 'localStorage' | 'sessionStorage'
  },

  // Connection settings
  connection: {
    mode: 'streaming',   // 'streaming' (SSE, default) | 'polling' | 'offline'
    pollingInterval: 30000, // Used only in polling mode
    timeout: 5000,
  },

  // Offline support
  offline: {
    enabled: true,
    defaultFlags: {
      'feature-a': false,
      'theme': 'light'
    }
  },

  // Performance
  performance: {
    batchRequests: true,
    prefetch: false,
  },

  // Logging
  logging: {
    level: 'warn',  // 'debug' | 'info' | 'warn' | 'error'
  }
});
```

## Migration from v1.0.3

**Breaking Changes**: Update your imports

```javascript
// ❌ Old (v1.0.3)
import { useFeatureFlag } from 'flexflag-client';

// ✅ New (v1.0.4)
// For React
import { useFeatureFlag } from 'flexflag-client/react';

// For Vue
import { useFeatureFlag } from 'flexflag-client/vue';

// Core SDK (backend/Node.js)
import { FlexFlagClient } from 'flexflag-client';
```

## Documentation

- [Complete Usage Guide](./README_SDK_USAGE.md) - Detailed examples
- [Upgrade Guide](./UPGRADE_GUIDE.md) - Migration instructions
- [Changelog](./CHANGELOG.md) - Version history

## Browser & Runtime Support

- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+
- Node.js 14+

## License

MIT

## Support

- GitHub: https://github.com/flexflag/flexflag
- Issues: https://github.com/flexflag/flexflag/issues