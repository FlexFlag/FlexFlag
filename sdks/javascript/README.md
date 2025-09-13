# FlexFlag JavaScript/TypeScript SDK

High-performance feature flag client for JavaScript and TypeScript applications with local caching and offline support.

## Installation

```bash
npm install flexflag-client
```

## Quick Start

```javascript
import { FlexFlagClient } from 'flexflag-client';

// Initialize the client
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.flexflag.io', // or your self-hosted instance
  environment: 'production'
});

// Wait for initialization
await client.initialize();

// Evaluate a feature flag
const isFeatureEnabled = await client.evaluateBoolean('new-feature', false);

if (isFeatureEnabled) {
  // Show new feature
} else {
  // Show default experience
}
```

## TypeScript Support

The SDK is written in TypeScript and provides full type definitions:

```typescript
import { FlexFlagClient, FlexFlagConfig, EvaluationContext } from 'flexflag-client';

const config: FlexFlagConfig = {
  apiKey: 'your-api-key',
  environment: 'production'
};

const client = new FlexFlagClient(config);

const context: EvaluationContext = {
  userId: 'user-123',
  attributes: {
    plan: 'premium',
    country: 'US'
  }
};

const flagValue = await client.evaluate('feature-flag', 'default', context);
```

## React Integration

```jsx
import { FlexFlagProvider, useFeatureFlag } from 'flexflag-client';

function App() {
  return (
    <FlexFlagProvider client={client}>
      <MyComponent />
    </FlexFlagProvider>
  );
}

function MyComponent() {
  const [isEnabled, loading] = useFeatureFlag('new-feature', false);
  
  if (loading) return <div>Loading...</div>;
  
  return isEnabled ? <NewFeature /> : <OldFeature />;
}
```

## Vue Integration

```vue
<template>
  <div v-if="loading">Loading...</div>
  <NewFeature v-else-if="isEnabled" />
  <OldFeature v-else />
</template>

<script setup>
import { useFeatureFlagVue } from 'flexflag-client';

const { value: isEnabled, loading } = useFeatureFlagVue('new-feature', false);
</script>
```

## Configuration Options

```javascript
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.flexflag.io',
  environment: 'production',
  
  // Cache configuration
  cache: {
    storage: 'memory', // 'memory', 'localStorage', 'sessionStorage'
    ttl: 300000, // 5 minutes
    maxSize: 1000
  },
  
  // Connection settings
  connection: {
    mode: 'streaming', // 'streaming', 'polling', 'offline'
    pollingInterval: 30000,
    timeout: 5000
  },
  
  // Offline support
  offline: {
    enabled: true,
    storageKey: 'flexflag-cache',
    defaultFlags: {
      'feature-1': true,
      'feature-2': false
    }
  }
});
```

## API Reference

### FlexFlagClient

#### Methods

- `initialize(): Promise<void>` - Initialize the client and fetch initial flags
- `evaluate(flagKey, defaultValue?, context?): Promise<FlagValue>` - Evaluate any flag type
- `evaluateBoolean(flagKey, defaultValue?, context?): Promise<boolean>` - Evaluate boolean flag
- `evaluateString(flagKey, defaultValue?, context?): Promise<string>` - Evaluate string flag
- `evaluateNumber(flagKey, defaultValue?, context?): Promise<number>` - Evaluate number flag
- `evaluateObject(flagKey, defaultValue?, context?): Promise<object>` - Evaluate object flag
- `close(): Promise<void>` - Close the client and cleanup resources
- `isReady(): boolean` - Check if client is ready

#### Events

- `ready` - Emitted when client is initialized
- `update` - Emitted when flags are updated
- `error` - Emitted on errors

```javascript
client.on('ready', () => {
  console.log('FlexFlag client is ready');
});

client.on('update', (updatedFlags) => {
  console.log('Flags updated:', updatedFlags);
});

client.on('error', (error) => {
  console.error('FlexFlag error:', error);
});
```

## Performance Features

- **Local Caching** - Flags are cached locally for instant evaluation
- **Streaming Updates** - Real-time flag updates via WebSocket/SSE
- **Offline Support** - Works offline with cached flags
- **Batch Evaluation** - Evaluate multiple flags efficiently
- **Smart Prefetching** - Pre-fetch commonly used flags

## Browser Support

- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+
- Node.js 14+

## License

MIT

## Support

- Documentation: https://github.com/flexflag/flexflag/tree/main/docs
- GitHub: https://github.com/flexflag/flexflag
- Issues: https://github.com/flexflag/flexflag/issues
- Discord: https://discord.gg/fpewTJyx9S