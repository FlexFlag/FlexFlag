/**
 * FlexFlag JavaScript/TypeScript SDK
 * High-performance feature flag client with local caching
 */

export { FlexFlagClient } from './client';
export { CacheProvider, MemoryCache, LocalStorageCache } from './cache';
export { 
  FlexFlagConfig,
  EvaluationContext,
  FlagValue,
  CacheConfig,
  ConnectionMode,
  LogLevel
} from './types';

// React hooks (if React is available)
export { useFeatureFlag, FlexFlagProvider } from './react';

// Vue composables (if Vue is available)
export { useFeatureFlag as useFeatureFlagVue } from './vue';

// Default export for convenience
import { FlexFlagClient } from './client';
export default FlexFlagClient;