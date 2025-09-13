# React Integration

The FlexFlag JavaScript SDK provides powerful React hooks and components for seamless integration with React applications.

## Installation

```bash
npm install flexflag-client
```

## Setup

### 1. Create FlexFlag Client

```jsx
// src/lib/flexflag.js
import { FlexFlagClient } from 'flexflag-client';

export const flexFlagClient = new FlexFlagClient({
  apiKey: process.env.REACT_APP_FLEXFLAG_API_KEY,
  baseUrl: process.env.REACT_APP_FLEXFLAG_BASE_URL,
  environment: process.env.REACT_APP_ENVIRONMENT || 'production',
  
  // Optional: Configure caching for better performance
  cache: {
    storage: 'localStorage',
    ttl: 300000, // 5 minutes
  },
  
  // Optional: Enable real-time updates
  connection: {
    mode: 'streaming'
  }
});
```

### 2. Add FlexFlag Provider

```jsx
// src/App.js
import React from 'react';
import { FlexFlagProvider } from 'flexflag-client';
import { flexFlagClient } from './lib/flexflag';
import Dashboard from './components/Dashboard';

function App() {
  const userContext = {
    userId: 'user-123',
    attributes: {
      plan: 'premium',
      country: 'US'
    }
  };

  return (
    <FlexFlagProvider client={flexFlagClient} context={userContext}>
      <div className="App">
        <Dashboard />
      </div>
    </FlexFlagProvider>
  );
}

export default App;
```

## Using Feature Flags in Components

### Basic Hook Usage

```jsx
// src/components/NewFeature.js
import React from 'react';
import { useFeatureFlag } from 'flexflag-client';

function NewFeature() {
  const { value: isEnabled, loading, error } = useFeatureFlag('new-checkout-flow', false);
  
  if (loading) {
    return <div>Loading...</div>;
  }
  
  if (error) {
    console.error('Feature flag error:', error);
    return <OldCheckoutFlow />; // Fallback to default
  }
  
  return isEnabled ? <NewCheckoutFlow /> : <OldCheckoutFlow />;
}

function NewCheckoutFlow() {
  return <div>🎉 New and improved checkout!</div>;
}

function OldCheckoutFlow() {
  return <div>Standard checkout flow</div>;
}

export default NewFeature;
```

### Hook with Context

```jsx
// src/components/UserSpecificFeature.js
import React from 'react';
import { useFeatureFlag } from 'flexflag-client';

function UserSpecificFeature({ user }) {
  const context = {
    userId: user.id,
    attributes: {
      plan: user.plan,
      country: user.country,
      signupDate: user.signupDate
    }
  };
  
  const { value: showBetaFeature, loading } = useFeatureFlag(
    'beta-dashboard', 
    false, 
    context
  );
  
  if (loading) return <div>Loading...</div>;
  
  return (
    <div>
      {showBetaFeature && <BetaDashboard user={user} />}
      <StandardDashboard user={user} />
    </div>
  );
}

export default UserSpecificFeature;
```

### Multiple Feature Flags

```jsx
// src/components/Dashboard.js
import React from 'react';
import { useFlexFlagClient } from 'flexflag-client';
import { useState, useEffect } from 'react';

function Dashboard() {
  const client = useFlexFlagClient();
  const [features, setFeatures] = useState({});
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const loadFeatures = async () => {
      try {
        const flags = await client.evaluateBatch([
          'dark-mode',
          'new-sidebar',
          'analytics-panel',
          'beta-features'
        ]);
        
        setFeatures(flags);
        setLoading(false);
      } catch (error) {
        console.error('Failed to load features:', error);
        setLoading(false);
      }
    };
    
    loadFeatures();
    
    // Listen for real-time updates
    const handleUpdate = (updatedFlags) => {
      console.log('Features updated:', updatedFlags);
      loadFeatures(); // Reload features
    };
    
    client.on('update', handleUpdate);
    
    return () => {
      client.off('update', handleUpdate);
    };
  }, [client]);
  
  if (loading) return <div>Loading dashboard...</div>;
  
  return (
    <div className={features['dark-mode'] ? 'dark-theme' : 'light-theme'}>
      <header>Dashboard</header>
      
      {features['new-sidebar'] && <NewSidebar />}
      
      <main>
        <h1>Welcome to your dashboard</h1>
        
        {features['analytics-panel'] && <AnalyticsPanel />}
        {features['beta-features'] && <BetaFeaturesPanel />}
      </main>
    </div>
  );
}

export default Dashboard;
```

## Advanced React Patterns

### Custom Hook for Feature Flags

```jsx
// src/hooks/useFeatures.js
import { useFlexFlagClient } from 'flexflag-client';
import { useState, useEffect } from 'react';

export function useFeatures(flagKeys, defaultValues = {}) {
  const client = useFlexFlagClient();
  const [features, setFeatures] = useState(defaultValues);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    const loadFeatures = async () => {
      try {
        setLoading(true);
        setError(null);
        
        const results = await client.evaluateBatch(flagKeys);
        setFeatures({ ...defaultValues, ...results });
      } catch (err) {
        setError(err);
        console.error('Failed to load features:', err);
      } finally {
        setLoading(false);
      }
    };
    
    loadFeatures();
    
    // Listen for updates
    const handleUpdate = (updatedFlags) => {
      const relevantUpdates = updatedFlags.filter(flag => flagKeys.includes(flag));
      if (relevantUpdates.length > 0) {
        loadFeatures();
      }
    };
    
    client.on('update', handleUpdate);
    
    return () => {
      client.off('update', handleUpdate);
    };
  }, [client, flagKeys]);
  
  return { features, loading, error, reload: () => loadFeatures() };
}

// Usage
function MyComponent() {
  const { features, loading } = useFeatures([
    'new-feature',
    'dark-mode',
    'beta-access'
  ], {
    'new-feature': false,
    'dark-mode': false,
    'beta-access': false
  });
  
  if (loading) return <div>Loading...</div>;
  
  return (
    <div>
      {features['new-feature'] && <NewFeature />}
      {features['beta-access'] && <BetaPanel />}
    </div>
  );
}
```

### Feature Flag Higher-Order Component

```jsx
// src/components/withFeatureFlag.js
import React from 'react';
import { useFeatureFlag } from 'flexflag-client';

export function withFeatureFlag(flagKey, defaultValue = false) {
  return function(Component) {
    return function FeatureFlagWrapper(props) {
      const { value: isEnabled, loading, error } = useFeatureFlag(flagKey, defaultValue);
      
      if (loading) {
        return <div>Loading...</div>;
      }
      
      if (error) {
        console.error(`Feature flag error for ${flagKey}:`, error);
        return null; // Or return a fallback component
      }
      
      if (!isEnabled) {
        return null; // Don't render the component if flag is disabled
      }
      
      return <Component {...props} />;
    };
  };
}

// Usage
const BetaFeature = withFeatureFlag('beta-feature')(function BetaFeature() {
  return <div>This is a beta feature!</div>;
});

// In your component
function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      <BetaFeature /> {/* Only renders if beta-feature flag is enabled */}
    </div>
  );
}
```

### Conditional Rendering Component

```jsx
// src/components/FeatureFlag.js
import React from 'react';
import { useFeatureFlag } from 'flexflag-client';

export function FeatureFlag({ 
  flag, 
  defaultValue = false, 
  context,
  fallback = null,
  loading = null,
  children 
}) {
  const { value: isEnabled, loading: isLoading, error } = useFeatureFlag(
    flag, 
    defaultValue, 
    context
  );
  
  if (isLoading) {
    return loading || <div>Loading...</div>;
  }
  
  if (error) {
    console.error(`Feature flag error for ${flag}:`, error);
    return fallback;
  }
  
  return isEnabled ? children : fallback;
}

// Usage
function App() {
  return (
    <div>
      <h1>My App</h1>
      
      <FeatureFlag 
        flag="new-header" 
        fallback={<OldHeader />}
        loading={<div>Loading header...</div>}
      >
        <NewHeader />
      </FeatureFlag>
      
      <FeatureFlag flag="beta-sidebar">
        <BetaSidebar />
      </FeatureFlag>
      
      <main>
        <FeatureFlag 
          flag="personalized-dashboard" 
          context={% raw %}{{ userId: 'user-123', plan: 'premium' }}{% endraw %}
          fallback={<StandardDashboard />}
        >
          <PersonalizedDashboard />
        </FeatureFlag>
      </main>
    </div>
  );
}
```

## Environment Configuration

### Development Setup

```jsx
// src/config/flexflag.js
const getFlexFlagConfig = () => {
  const isDevelopment = process.env.NODE_ENV === 'development';
  
  return {
    apiKey: process.env.REACT_APP_FLEXFLAG_API_KEY,
    baseUrl: isDevelopment 
      ? 'http://localhost:8080' 
      : process.env.REACT_APP_FLEXFLAG_BASE_URL,
    environment: process.env.REACT_APP_ENVIRONMENT || 'development',
    
    // Enable debug logging in development
    logging: {
      level: isDevelopment ? 'debug' : 'warn'
    },
    
    // Faster polling in development
    connection: {
      mode: 'polling',
      pollingInterval: isDevelopment ? 5000 : 30000
    },
    
    // Disable caching in development for immediate updates
    cache: {
      enabled: !isDevelopment,
      storage: 'localStorage'
    }
  };
};

export const flexFlagClient = new FlexFlagClient(getFlexFlagConfig());
```

### Environment Variables (.env)

```bash
# .env.development
REACT_APP_FLEXFLAG_API_KEY=your-dev-api-key
REACT_APP_FLEXFLAG_BASE_URL=http://localhost:8080
REACT_APP_ENVIRONMENT=development

# .env.production
REACT_APP_FLEXFLAG_API_KEY=your-prod-api-key
REACT_APP_FLEXFLAG_BASE_URL=https://api.yourapp.com
REACT_APP_ENVIRONMENT=production
```

## Performance Tips

1. **Use the Provider**: Always wrap your app with `FlexFlagProvider` for optimal performance
2. **Batch Evaluations**: Use `evaluateBatch()` for multiple flags
3. **Cache Configuration**: Use `localStorage` caching for persistent storage
4. **Context Optimization**: Avoid creating new context objects on every render
5. **Memoization**: Use `useMemo` for expensive context calculations

```jsx
import React, { useMemo } from 'react';
import { FlexFlagProvider } from 'flexflag-client';

function App({ user }) {
  // Memoize context to prevent unnecessary re-evaluations
  const userContext = useMemo(() => ({
    userId: user.id,
    attributes: {
      plan: user.plan,
      country: user.country
    }
  }), [user.id, user.plan, user.country]);
  
  return (
    <FlexFlagProvider client={flexFlagClient} context={userContext}>
      <YourApp />
    </FlexFlagProvider>
  );
}
```

## Troubleshooting

### Common Issues

1. **Provider Not Found**: Ensure `FlexFlagProvider` wraps your component tree
2. **Stale Values**: Check if context is being updated properly
3. **Performance Issues**: Use batch evaluation for multiple flags
4. **Network Errors**: Implement proper error handling and fallbacks

### Debug Mode

```jsx
const client = new FlexFlagClient({
  apiKey: 'your-api-key',
  baseUrl: 'http://localhost:8080',
  environment: 'development',
  
  // Enable debug logging
  logging: {
    level: 'debug'
  },
  
  // Debug event callbacks
  events: {
    onEvaluation: (flag, value) => console.log(`🚩 ${flag}: ${value}`),
    onCacheHit: (flag) => console.log(`💾 Cache hit: ${flag}`),
    onCacheMiss: (flag) => console.log(`🔍 Cache miss: ${flag}`),
    onError: (error) => console.error('🚨 FlexFlag error:', error)
  }
});
```