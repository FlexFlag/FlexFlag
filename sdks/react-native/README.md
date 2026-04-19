# flexflag-react-native

React Native SDK for [FlexFlag](https://github.com/flexflag/flexflag) — offline-first feature flags and remote config for iOS and Android apps.

## Features

- Boolean, string, number, and JSON flag evaluation
- Automatic platform detection (`ios` / `android`) for platform-targeted flags
- In-memory cache with configurable TTL
- AsyncStorage persistence — flags survive app restarts
- AppState listener — refreshes flags when the app returns to foreground
- Polling-based updates (30s default, configurable)
- Offline defaults for when the server is unreachable
- React hooks and `FeatureGate` component

## Installation

```sh
npm install flexflag-react-native
# or
yarn add flexflag-react-native
```

AsyncStorage is a peer dependency:

```sh
npm install @react-native-async-storage/async-storage
```

Follow the [AsyncStorage setup guide](https://react-native-async-storage.github.io/async-storage/docs/install) to link it on iOS and Android.

## Quick Start

### 1. Wrap your app with `FlexFlagProvider`

```tsx
import AsyncStorage from '@react-native-async-storage/async-storage';
import { FlexFlagProvider } from 'flexflag-react-native';

export default function App() {
  return (
    <FlexFlagProvider
      config={{
        apiKey: 'your-api-key',
        baseUrl: 'https://your-flexflag-server.com',
        environment: 'production',
      }}
      storage={AsyncStorage}
    >
      <RootNavigator />
    </FlexFlagProvider>
  );
}
```

### 2. Use hooks in your components

```tsx
import { useBooleanFlag, useStringFlag } from 'flexflag-react-native';

function CheckoutScreen() {
  const { value: newCheckoutEnabled, loading } = useBooleanFlag('new-checkout', false);
  const { value: ctaText } = useStringFlag('checkout-cta', 'Buy Now');

  if (loading) return <ActivityIndicator />;

  return newCheckoutEnabled ? (
    <NewCheckout ctaText={ctaText} />
  ) : (
    <LegacyCheckout />
  );
}
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | `string` | required | Your FlexFlag project API key |
| `baseUrl` | `string` | `'http://localhost:8080'` | Base URL of your FlexFlag server |
| `environment` | `string` | `'production'` | Flag environment to evaluate against |
| `pollingInterval` | `number` | `30000` | How often (ms) to refresh flags. Set to `0` to disable |
| `timeout` | `number` | `5000` | HTTP request timeout in ms |
| `cacheTtl` | `number` | `300000` | How long (ms) to keep a cached flag value |
| `offlineDefaults` | `Record<string, FlagValue>` | `{}` | Fallback values used when the server is unreachable |
| `persistenceKey` | `string` | `'flexflag_cache'` | AsyncStorage key for persisting the flag cache |

## Hooks

### `useBooleanFlag`

```tsx
const { value, loading, error, reload } = useBooleanFlag(
  'flag-key',
  false,       // default value
  { userId: 'user-123' }  // optional evaluation context
);
```

### `useStringFlag`

```tsx
const { value } = useStringFlag('app-theme', 'light');
```

### `useNumberFlag`

```tsx
const { value } = useNumberFlag('max-retries', 3);
```

### `useJSONFlag`

```tsx
interface OnboardingConfig {
  steps: string[];
  skipable: boolean;
}

const { value } = useJSONFlag<OnboardingConfig>('onboarding-config', {
  steps: ['welcome'],
  skipable: false,
});
```

### `useRemoteConfig`

Fetches all config values for the environment in one request. Useful at app startup to warm the config cache.

```tsx
const { config, loading } = useRemoteConfig({ userId: 'user-123' });
const buttonColor = config?.config['button-color'] ?? '#007AFF';
```

## `FeatureGate` Component

Conditionally render components based on a boolean flag:

```tsx
import { FeatureGate } from 'flexflag-react-native';

<FeatureGate flagKey="new-profile" fallback={<OldProfile />}>
  <NewProfile />
</FeatureGate>
```

## Evaluation Context

Pass user and attribute context to enable targeting rules:

```tsx
const { value } = useBooleanFlag('premium-feature', false, {
  userId: 'user-123',
  attributes: {
    plan: 'pro',
    country: 'US',
    age: 28,
  },
});
```

The SDK automatically includes the detected platform (`ios` or `android`) in every evaluation request so you can target flags by platform without adding it to the context manually.

## Offline Support

When the server is unreachable, the SDK falls back in this order:

1. **In-memory cache** — values fetched during this session
2. **Persisted cache** — values saved to AsyncStorage from previous sessions
3. **`offlineDefaults`** — values you provide in config
4. **`defaultValue`** — the value passed to the hook

```tsx
<FlexFlagProvider
  config={{
    apiKey: 'your-api-key',
    offlineDefaults: {
      'new-checkout': false,
      'app-theme': 'light',
    },
  }}
  storage={AsyncStorage}
>
```

## Direct Client Usage

If you need to evaluate flags outside of React components:

```tsx
import AsyncStorage from '@react-native-async-storage/async-storage';
import { FlexFlagClient } from 'flexflag-react-native';

const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  environment: 'production',
});

await client.initialize(AsyncStorage);

const enabled = await client.evaluateBoolean('new-feature', false, {
  userId: 'user-123',
});

// Clean up when done
await client.close();
```

## AppState Behaviour

The SDK automatically refreshes flags when the app returns to the foreground (`AppState === 'active'`). This keeps flags current without requiring a full app restart.

## TypeScript

Full TypeScript support is included. All hooks and the client are fully typed — no `@types` package needed.

## License

MIT
