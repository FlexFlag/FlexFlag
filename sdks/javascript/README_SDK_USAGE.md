# FlexFlag SDK - Complete Usage Guide

## Authentication

The SDK supports two authentication methods:

### 1. API Key Authentication (Recommended for SDKs)

First, create an API key in your FlexFlag instance:

```bash
# Start the FlexFlag server
cd /path/to/FlexFlag
make run

# In another terminal, create an API key (you'll need to be authenticated first)
# Or use the UI at http://localhost:3000 to create an API key
```

Then use it in your SDK:

```javascript
const { FlexFlagClient } = require('flexflag-client');

const client = new FlexFlagClient({
  apiKey: 'your_actual_api_key_here',  // Get this from the FlexFlag UI or API
  baseUrl: 'http://localhost:8080',
  environment: 'production'
});
```

### 2. JWT Token Authentication (For browser/frontend)

```javascript
const { FlexFlagClient } = require('flexflag-client');

const client = new FlexFlagClient({
  baseUrl: 'http://localhost:8080',
  environment: 'production',
  connection: {
    headers: {
      'Authorization': `Bearer ${yourJwtToken}`
    }
  }
});
```

## Backend Usage (Node.js/TypeScript)

### Basic Setup

```javascript
const { FlexFlagClient } = require('flexflag-client');

// Initialize the client
const client = new FlexFlagClient({
  apiKey: process.env.FLEXFLAG_API_KEY,
  baseUrl: process.env.FLEXFLAG_BASE_URL || 'http://localhost:8080',
  environment: 'production',

  // Enable offline mode for fallback
  offline: {
    enabled: true,
    defaultFlags: {
      'feature-x': false,
      'theme': 'light',
      'max-retries': 3
    }
  },

  // Configure caching
  cache: {
    enabled: true,
    ttl: 300000, // 5 minutes
    storage: 'memory'
  }
});

// Wait for SDK to be ready
await client.waitForReady();
```

### Evaluating Flags

```javascript
// Boolean flags
const isFeatureEnabled = await client.evaluateBoolean('feature-x', {
  userId: 'user123',
  attributes: {
    plan: 'premium',
    region: 'us-east'
  }
});

// String flags
const theme = await client.evaluateString('ui-theme', {
  userId: 'user123'
}, 'light');

// Number flags
const maxRetries = await client.evaluateNumber('max-retries', {
  userId: 'user123'
}, 3);

// JSON/Object flags
const config = await client.evaluateJSON('app-config', {
  userId: 'user123'
}, {});

// Batch evaluation (more efficient for multiple flags)
const results = await client.evaluateBatch(['feature-x', 'ui-theme', 'max-retries'], {
  userId: 'user123',
  attributes: { plan: 'premium' }
});
console.log(results); // { 'feature-x': true, 'ui-theme': 'dark', 'max-retries': 5 }
```

### Express.js Example

```javascript
const express = require('express');
const { FlexFlagClient } = require('flexflag-client');

const app = express();
const flexflag = new FlexFlagClient({
  apiKey: process.env.FLEXFLAG_API_KEY,
  baseUrl: 'http://localhost:8080',
  environment: 'production'
});

app.get('/api/features', async (req, res) => {
  const userId = req.user?.id || 'anonymous';

  const features = await flexflag.evaluateBatch(
    ['dark-mode', 'new-ui', 'beta-features'],
    {
      userId,
      attributes: {
        plan: req.user?.plan || 'free',
        region: req.user?.region || 'us'
      }
    }
  );

  res.json({ features });
});

app.listen(3000);
```

## React Usage

```javascript
import React from 'react';
import {
  FlexFlagProvider,
  useBooleanFlag,
  useStringFlag,
  FeatureGate
} from 'flexflag-client/react';

function App() {
  return (
    <FlexFlagProvider
      config={{
        apiKey: process.env.REACT_APP_FLEXFLAG_API_KEY,
        baseUrl: 'http://localhost:8080',
        environment: 'production'
      }}
      context={{
        userId: 'user123',
        attributes: { plan: 'premium' }
      }}
    >
      <Dashboard />
    </FlexFlagProvider>
  );
}

function Dashboard() {
  const { enabled: darkMode, loading } = useBooleanFlag('dark-mode');
  const { value: theme } = useStringFlag('ui-theme', 'light');

  if (loading) return <div>Loading...</div>;

  return (
    <div className={darkMode ? 'dark' : 'light'}>
      <h1>Theme: {theme}</h1>

      {/* Conditional rendering with FeatureGate */}
      <FeatureGate flagKey="new-ui" fallback={<OldUI />}>
        <NewUI />
      </FeatureGate>
    </div>
  );
}
```

## Vue 3 Usage

```javascript
import { createApp } from 'vue';
import { createFlexFlag, useBooleanFlag } from 'flexflag-client/vue';
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
    <h1>Dark Mode: {{ enabled }}</h1>
  </div>
</template>

<script setup>
import { useBooleanFlag } from 'flexflag-client/vue';

const { enabled, loading } = useBooleanFlag('dark-mode');
</script>
```

## Troubleshooting

### 401 Unauthorized Error

If you're getting 401 errors:

1. **Check your API key**: Make sure you're using a valid API key created in FlexFlag
2. **Verify the environment**: Ensure the API key has permissions for the environment you're using
3. **Check the baseUrl**: Make sure the FlexFlag server is running and accessible

```javascript
// For testing/development, you can use offline mode
const client = new FlexFlagClient({
  baseUrl: 'http://localhost:8080',
  connection: {
    mode: 'offline'  // Skip server connection
  },
  offline: {
    enabled: true,
    defaultFlags: {
      'my-flag': true
    }
  }
});
```

### Missing React/Vue Error

If you get errors about React or Vue not being found:

```bash
# For backend/Node.js (no React/Vue needed)
npm install flexflag-client

# For React projects
npm install flexflag-client react

# For Vue projects
npm install flexflag-client vue
```

Then import from the correct entry point:
- Core SDK: `require('flexflag-client')` or `import from 'flexflag-client'`
- React: `import from 'flexflag-client/react'`
- Vue: `import from 'flexflag-client/vue'`

## Environment Variables

Create a `.env` file:

```env
# Backend/Node.js
FLEXFLAG_API_KEY=your_api_key_here
FLEXFLAG_BASE_URL=http://localhost:8080
FLEXFLAG_ENVIRONMENT=production

# React
REACT_APP_FLEXFLAG_API_KEY=your_api_key_here
REACT_APP_FLEXFLAG_BASE_URL=http://localhost:8080

# Vue/Vite
VITE_FLEXFLAG_API_KEY=your_api_key_here
VITE_FLEXFLAG_BASE_URL=http://localhost:8080
```

## API Reference

### Client Methods

- `evaluate(flagKey, context?, defaultValue?)` - Evaluate any flag type
- `evaluateBoolean(flagKey, context?, defaultValue?)` - Evaluate boolean flag
- `evaluateString(flagKey, context?, defaultValue?)` - Evaluate string flag
- `evaluateNumber(flagKey, context?, defaultValue?)` - Evaluate number flag
- `evaluateJSON<T>(flagKey, context?, defaultValue?)` - Evaluate JSON flag
- `evaluateBatch(flagKeys[], context?)` - Evaluate multiple flags at once
- `getVariation(flagKey, context?)` - Get variation for A/B testing
- `setContext(context)` - Set default context for all evaluations
- `clearCache()` - Clear all cached flags
- `getMetrics()` - Get SDK metrics
- `close()` - Close SDK connections and cleanup

### Context Object

```typescript
{
  userId?: string;           // User identifier
  attributes?: {             // Custom attributes
    plan?: string;
    region?: string;
    email?: string;
    // ... any custom attributes
  }
}
```
