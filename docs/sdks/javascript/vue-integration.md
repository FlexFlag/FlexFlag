# Vue Integration

The FlexFlag JavaScript SDK provides Vue 3 composables and plugins for seamless integration with Vue applications.

## Installation

```bash
npm install flexflag-client
```

## Setup

### 1. Create FlexFlag Client

```javascript
// src/lib/flexflag.js
import { FlexFlagClient } from 'flexflag-client';

export const flexFlagClient = new FlexFlagClient({
  apiKey: import.meta.env.VITE_FLEXFLAG_API_KEY,
  baseUrl: import.meta.env.VITE_FLEXFLAG_BASE_URL,
  environment: import.meta.env.VITE_ENVIRONMENT || 'production',
  
  // Optional: Configure caching
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

### 2. Vue Plugin Setup

```javascript
// src/main.js
import { createApp } from 'vue';
import { flexFlagClient } from './lib/flexflag';
import App from './App.vue';

const app = createApp(App);

// Provide FlexFlag client globally
app.provide('flexflag-client', flexFlagClient);

app.mount('#app');
```

### 3. Alternative: Manual Provider Setup

```vue
<!-- src/App.vue -->
<template>
  <div id="app">
    <Dashboard />
  </div>
</template>

<script setup>
import { provide, reactive } from 'vue';
import { flexFlagClient } from './lib/flexflag';
import Dashboard from './components/Dashboard.vue';

// Provide FlexFlag client to child components
provide('flexflag-client', flexFlagClient);

// Optional: Provide user context
const userContext = reactive({
  userId: 'user-123',
  attributes: {
    plan: 'premium',
    country: 'US'
  }
});

provide('flexflag-context', userContext);
</script>
```

## Using Feature Flags in Components

### Basic Composable Usage

```vue
<!-- src/components/NewFeature.vue -->
<template>
  <div>
    <div v-if="loading">Loading...</div>
    <div v-else-if="error">
      Error loading feature flag: {% raw %}{{ error.message }}{% endraw %}
    </div>
    <div v-else>
      <NewCheckoutFlow v-if="isEnabled" />
      <OldCheckoutFlow v-else />
    </div>
  </div>
</template>

<script setup>
import { useFeatureFlagVue } from 'flexflag-client';
import NewCheckoutFlow from './NewCheckoutFlow.vue';
import OldCheckoutFlow from './OldCheckoutFlow.vue';

const { value: isEnabled, loading, error } = useFeatureFlagVue('new-checkout-flow', false);
</script>
```

### Composable with Context

```vue
<!-- src/components/UserSpecificFeature.vue -->
<template>
  <div>
    <BetaDashboard v-if="showBetaFeature && !loading" :user="user" />
    <StandardDashboard :user="user" />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useFeatureFlagVue } from 'flexflag-client';
import BetaDashboard from './BetaDashboard.vue';
import StandardDashboard from './StandardDashboard.vue';

const props = defineProps({
  user: {
    type: Object,
    required: true
  }
});

// Create context based on user props
const context = computed(() => ({
  userId: props.user.id,
  attributes: {
    plan: props.user.plan,
    country: props.user.country,
    signupDate: props.user.signupDate
  }
}));

const { value: showBetaFeature, loading } = useFeatureFlagVue(
  'beta-dashboard', 
  false, 
  context
);
</script>
```

### Multiple Feature Flags

```vue
<!-- src/components/Dashboard.vue -->
<template>
  <div :class="{ 'dark-theme': features.darkMode, 'light-theme': !features.darkMode }">
    <header>Dashboard</header>
    
    <NewSidebar v-if="features.newSidebar" />
    
    <main>
      <h1>Welcome to your dashboard</h1>
      
      <AnalyticsPanel v-if="features.analyticsPanel" />
      <BetaFeaturesPanel v-if="features.betaFeatures" />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useFlexFlagClient } from 'flexflag-client';
import NewSidebar from './NewSidebar.vue';
import AnalyticsPanel from './AnalyticsPanel.vue';
import BetaFeaturesPanel from './BetaFeaturesPanel.vue';

const client = useFlexFlagClient();
const features = ref({
  darkMode: false,
  newSidebar: false,
  analyticsPanel: false,
  betaFeatures: false
});
const loading = ref(true);

const loadFeatures = async () => {
  try {
    loading.value = true;
    const flags = await client.evaluateBatch([
      'dark-mode',
      'new-sidebar', 
      'analytics-panel',
      'beta-features'
    ]);
    
    features.value = {
      darkMode: flags['dark-mode'],
      newSidebar: flags['new-sidebar'],
      analyticsPanel: flags['analytics-panel'],
      betaFeatures: flags['beta-features']
    };
  } catch (error) {
    console.error('Failed to load features:', error);
  } finally {
    loading.value = false;
  }
};

const handleUpdate = (updatedFlags) => {
  console.log('Features updated:', updatedFlags);
  loadFeatures(); // Reload features
};

onMounted(() => {
  loadFeatures();
  client.on('update', handleUpdate);
});

onUnmounted(() => {
  client.off('update', handleUpdate);
});
</script>

<style scoped>
.dark-theme {
  background-color: #1a1a1a;
  color: white;
}

.light-theme {
  background-color: white;
  color: black;
}
</style>
```

## Advanced Vue Patterns

### Custom Composable for Multiple Features

```javascript
// src/composables/useFeatures.js
import { ref, onMounted, onUnmounted } from 'vue';
import { useFlexFlagClient } from 'flexflag-client';

export function useFeatures(flagKeys, defaultValues = {}) {
  const client = useFlexFlagClient();
  const features = ref({ ...defaultValues });
  const loading = ref(true);
  const error = ref(null);
  
  const loadFeatures = async () => {
    try {
      loading.value = true;
      error.value = null;
      
      const results = await client.evaluateBatch(flagKeys);
      features.value = { ...defaultValues, ...results };
    } catch (err) {
      error.value = err;
      console.error('Failed to load features:', err);
    } finally {
      loading.value = false;
    }
  };
  
  const handleUpdate = (updatedFlags) => {
    const relevantUpdates = updatedFlags.filter(flag => flagKeys.includes(flag));
    if (relevantUpdates.length > 0) {
      loadFeatures();
    }
  };
  
  onMounted(() => {
    loadFeatures();
    client.on('update', handleUpdate);
  });
  
  onUnmounted(() => {
    client.off('update', handleUpdate);
  });
  
  return {
    features,
    loading,
    error,
    reload: loadFeatures
  };
}
```

Usage:

```vue
<!-- src/components/MyComponent.vue -->
<template>
  <div v-if="loading">Loading features...</div>
  <div v-else>
    <NewFeature v-if="features.newFeature" />
    <BetaPanel v-if="features.betaAccess" />
  </div>
</template>

<script setup>
import { useFeatures } from '../composables/useFeatures';
import NewFeature from './NewFeature.vue';
import BetaPanel from './BetaPanel.vue';

const { features, loading } = useFeatures([
  'new-feature',
  'beta-access'
], {
  newFeature: false,
  betaAccess: false
});
</script>
```

### Feature Flag Directive

```javascript
// src/directives/featureFlag.js
import { useFlexFlagClient } from 'flexflag-client';

export const vFeatureFlag = {
  async mounted(el, binding) {
    const client = useFlexFlagClient();
    const flagKey = binding.value;
    const defaultValue = binding.modifiers.enabled ? true : false;
    
    try {
      const isEnabled = await client.evaluate(flagKey, defaultValue);
      
      if (!isEnabled) {
        el.style.display = 'none';
      }
    } catch (error) {
      console.error(`Feature flag directive error for ${flagKey}:`, error);
      // On error, hide element by default unless explicitly enabled
      if (!binding.modifiers.enabled) {
        el.style.display = 'none';
      }
    }
  }
};

// Register globally in main.js
import { vFeatureFlag } from './directives/featureFlag';
app.directive('feature-flag', vFeatureFlag);
```

Usage:

```vue
<template>
  <div>
    <!-- Element will be hidden if 'beta-feature' flag is false -->
    <div v-feature-flag="'beta-feature'">
      This is a beta feature!
    </div>
    
    <!-- Element will be shown by default if flag fails to load -->
    <div v-feature-flag.enabled="'new-ui'">
      New UI component
    </div>
  </div>
</template>
```

### Feature Flag Component

```vue
<!-- src/components/FeatureFlag.vue -->
<template>
  <div v-if="loading && showLoading">
    <slot name="loading">Loading...</slot>
  </div>
  <div v-else-if="error && showError">
    <slot name="error" :error="error">
      Error: {% raw %}{{ error.message }}{% endraw %}
    </slot>
  </div>
  <div v-else-if="isEnabled">
    <slot />
  </div>
  <div v-else-if="$slots.fallback">
    <slot name="fallback" />
  </div>
</template>

<script setup>
import { useFeatureFlagVue } from 'flexflag-client';

const props = defineProps({
  flag: {
    type: String,
    required: true
  },
  defaultValue: {
    type: Boolean,
    default: false
  },
  context: {
    type: Object,
    default: () => ({})
  },
  showLoading: {
    type: Boolean,
    default: true
  },
  showError: {
    type: Boolean,
    default: false
  }
});

const { value: isEnabled, loading, error } = useFeatureFlagVue(
  props.flag,
  props.defaultValue,
  props.context
);
</script>
```

Usage:

```vue
<template>
  <div>
    <FeatureFlag 
      flag="new-header" 
      :show-loading="false"
    >
      <NewHeader />
      
      <template #fallback>
        <OldHeader />
      </template>
    </FeatureFlag>
    
    <FeatureFlag 
      flag="personalized-dashboard"
      :context="userContext"
      :show-error="true"
    >
      <PersonalizedDashboard />
      
      <template #loading>
        <div class="spinner">Loading dashboard...</div>
      </template>
      
      <template #error="{ error }">
        <div class="error">Failed to load: {% raw %}{{ error.message }}{% endraw %}</div>
      </template>
      
      <template #fallback>
        <StandardDashboard />
      </template>
    </FeatureFlag>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import FeatureFlag from './FeatureFlag.vue';

const props = defineProps(['user']);

const userContext = computed(() => ({
  userId: props.user?.id,
  attributes: {
    plan: props.user?.plan,
    country: props.user?.country
  }
}));
</script>
```

## Pinia Store Integration

```javascript
// src/stores/features.js
import { defineStore } from 'pinia';
import { flexFlagClient } from '../lib/flexflag';

export const useFeatureStore = defineStore('features', {
  state: () => ({
    flags: {},
    loading: false,
    error: null
  }),
  
  getters: {
    isEnabled: (state) => (flagKey) => {
      return state.flags[flagKey] || false;
    },
    
    hasError: (state) => !!state.error,
    isLoading: (state) => state.loading
  },
  
  actions: {
    async loadFlag(flagKey, defaultValue = false, context = {}) {
      try {
        this.loading = true;
        this.error = null;
        
        const value = await flexFlagClient.evaluate(flagKey, defaultValue, context);
        this.flags[flagKey] = value;
        
        return value;
      } catch (error) {
        this.error = error;
        console.error(`Failed to load flag ${flagKey}:`, error);
        return defaultValue;
      } finally {
        this.loading = false;
      }
    },
    
    async loadFlags(flagKeys, defaultValues = {}, context = {}) {
      try {
        this.loading = true;
        this.error = null;
        
        const results = await flexFlagClient.evaluateBatch(flagKeys, context);
        
        flagKeys.forEach(key => {
          this.flags[key] = results[key] ?? defaultValues[key] ?? false;
        });
        
        return results;
      } catch (error) {
        this.error = error;
        console.error('Failed to load flags:', error);
        
        // Apply defaults on error
        flagKeys.forEach(key => {
          this.flags[key] = defaultValues[key] ?? false;
        });
        
        return defaultValues;
      } finally {
        this.loading = false;
      }
    },
    
    initializeListeners() {
      flexFlagClient.on('update', (updatedFlags) => {
        console.log('Flags updated:', updatedFlags);
        // Reload updated flags
        updatedFlags.forEach(flagKey => {
          if (this.flags.hasOwnProperty(flagKey)) {
            this.loadFlag(flagKey);
          }
        });
      });
    }
  }
});
```

Usage in components:

```vue
<!-- src/components/Dashboard.vue -->
<template>
  <div>
    <div v-if="featureStore.isLoading">Loading features...</div>
    <div v-else>
      <NewSidebar v-if="featureStore.isEnabled('new-sidebar')" />
      <AnalyticsPanel v-if="featureStore.isEnabled('analytics-panel')" />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useFeatureStore } from '../stores/features';

const featureStore = useFeatureStore();

onMounted(() => {
  // Load initial flags
  featureStore.loadFlags([
    'new-sidebar',
    'analytics-panel',
    'dark-mode'
  ]);
  
  // Initialize real-time listeners
  featureStore.initializeListeners();
});
</script>
```

## Environment Configuration

### Development Setup

```javascript
// src/config/flexflag.js
const getFlexFlagConfig = () => {
  const isDevelopment = import.meta.env.MODE === 'development';
  
  return {
    apiKey: import.meta.env.VITE_FLEXFLAG_API_KEY,
    baseUrl: isDevelopment 
      ? 'http://localhost:8080' 
      : import.meta.env.VITE_FLEXFLAG_BASE_URL,
    environment: import.meta.env.VITE_ENVIRONMENT || 'development',
    
    // Enable debug logging in development
    logging: {
      level: isDevelopment ? 'debug' : 'warn'
    },
    
    // Faster polling in development
    connection: {
      mode: 'polling',
      pollingInterval: isDevelopment ? 5000 : 30000
    }
  };
};

export const flexFlagClient = new FlexFlagClient(getFlexFlagConfig());
```

### Environment Variables (.env)

```bash
# .env.development
VITE_FLEXFLAG_API_KEY=your-dev-api-key
VITE_FLEXFLAG_BASE_URL=http://localhost:8080
VITE_ENVIRONMENT=development

# .env.production
VITE_FLEXFLAG_API_KEY=your-prod-api-key
VITE_FLEXFLAG_BASE_URL=https://api.yourapp.com
VITE_ENVIRONMENT=production
```

## Performance Tips

1. **Use Batch Evaluation**: Load multiple flags at once with `evaluateBatch()`
2. **Cache Configuration**: Use `localStorage` for persistent caching
3. **Context Reactivity**: Use `computed` for dynamic context values
4. **Store Integration**: Use Pinia/Vuex for centralized flag management
5. **Component Lazy Loading**: Only load feature components when flags are enabled

## Troubleshooting

### Common Issues

1. **Client Not Found**: Ensure client is properly provided at the app level
2. **Reactive Context**: Use `computed` for context that changes with props/state
3. **Memory Leaks**: Always clean up event listeners with `onUnmounted`
4. **SSR Issues**: Handle server-side rendering by providing default values

### Debug Mode

```javascript
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
    onError: (error) => console.error('🚨 FlexFlag error:', error)
  }
});
```