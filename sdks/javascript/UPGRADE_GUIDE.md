# FlexFlag SDK v1.0.3 - Upgrade Guide

## What's New

### Fixed Issues
1. **React/Vue Dependencies**: React and Vue are now optional peer dependencies, allowing the SDK to work in backend environments without installing these frameworks.
2. **API Endpoint Fixes**: Fixed the request payload format for single and batch evaluation to match the backend API expectations.
3. **New Convenience Methods**: Added type-specific evaluation methods for better developer experience.

### Breaking Changes
⚠️ **Important**: If you were previously importing React or Vue integrations from the main package, you must update your imports.

## Migration Guide

### For Backend/Node.js Users

**Before v1.0.3:**
```javascript
// This would fail if React/Vue weren't installed
const { FlexFlagClient } = require('flexflag-client');
```

**After v1.0.3:**
```javascript
// Works without React/Vue installed
const { FlexFlagClient } = require('flexflag-client');

const client = new FlexFlagClient({
  apiKey: 'your_api_key',
  baseUrl: 'http://localhost:8080',
  environment: 'production'
});

// New convenience methods
const isEnabled = await client.evaluateBoolean('feature-flag');
const theme = await client.evaluateString('ui-theme', undefined, 'light');
const maxRetries = await client.evaluateNumber('max-retries', undefined, 3);
const config = await client.evaluateJSON('app-config');
```

### For React Users

**Before v1.0.3:**
```javascript
import { FlexFlagProvider, useFeatureFlag } from 'flexflag-client';
```

**After v1.0.3:**
```javascript
// Import from the React-specific entry point
import { FlexFlagProvider, useFeatureFlag, useBooleanFlag } from 'flexflag-client/react';

function App() {
  return (
    <FlexFlagProvider config={{ apiKey: 'your_key', baseUrl: 'http://localhost:8080' }}>
      <MyComponent />
    </FlexFlagProvider>
  );
}

function MyComponent() {
  const { enabled } = useBooleanFlag('dark-mode');
  return <div>{enabled ? 'Dark' : 'Light'} mode</div>;
}
```

### For Vue 3 Users

**Before v1.0.3:**
```javascript
import { useFeatureFlag } from 'flexflag-client';
```

**After v1.0.3:**
```javascript
// Import from the Vue-specific entry point
import { createFlexFlag, useFeatureFlag, useBooleanFlag } from 'flexflag-client/vue';

const app = createApp(App);
app.use(createFlexFlag({
  apiKey: 'your_key',
  baseUrl: 'http://localhost:8080'
}));
```

## New Features

### Type-Specific Evaluation Methods

```javascript
// Boolean flags
const isEnabled = await client.evaluateBoolean('feature-flag', context, false);

// String flags
const theme = await client.evaluateString('ui-theme', context, 'light');

// Number flags
const timeout = await client.evaluateNumber('timeout-ms', context, 5000);

// JSON flags
const config = await client.evaluateJSON('app-config', context, {});
```

### Improved Error Handling

The SDK now properly falls back to offline defaults when:
- API authentication fails
- Network is unavailable
- Server returns an error

```javascript
const client = new FlexFlagClient({
  apiKey: 'your_key',
  baseUrl: 'http://localhost:8080',
  offline: {
    enabled: true,
    defaultFlags: {
      'feature-flag': false,
      'ui-theme': 'light'
    }
  }
});

// If API fails, returns offline default
const value = await client.evaluateBoolean('feature-flag'); // Returns false
```

## Installation

### Backend/Node.js
```bash
npm install flexflag-client
```

### With React
```bash
npm install flexflag-client react
# or
npm install flexflag-client  # if React is already installed
```

### With Vue 3
```bash
npm install flexflag-client vue
# or
npm install flexflag-client  # if Vue is already installed
```

## Package Exports

The package now provides three entry points:

- `flexflag-client` - Core SDK (no framework dependencies)
- `flexflag-client/react` - React integration
- `flexflag-client/vue` - Vue 3 integration

## Authorization Fix

The SDK now correctly sends evaluation requests with the proper format:

```javascript
// Evaluation request format
{
  flag_key: 'your-flag',
  user_id: 'user123',
  user_key: 'user123',
  attributes: {
    plan: 'premium',
    region: 'us-east'
  }
}
```

Make sure your API key has the necessary permissions to evaluate flags in your environment.

## Questions?

If you encounter any issues during migration, please check:
1. Your import statements match the new entry points
2. Your API key is valid and has evaluation permissions
3. The backend is running and accessible
4. You've installed the correct peer dependencies for your framework (React or Vue)
