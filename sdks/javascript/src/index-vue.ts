/**
 * FlexFlag Vue 3 Integration
 * Vue composables and directives for FlexFlag
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

// Vue composables and utilities
export {
  useFeatureFlag,
  useBooleanFlag,
  useStringFlag,
  useNumberFlag,
  useVariation,
  useBatchFlags,
  useFlexFlagClient,
  useFlexFlagMetrics,
  createFlexFlag,
  provideFlexFlag,
  vFeatureFlag
} from './vue';
