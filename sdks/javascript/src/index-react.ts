/**
 * FlexFlag React Integration
 * React hooks and components for FlexFlag
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

// React hooks and components
export {
  useFeatureFlag,
  useBooleanFlag,
  useStringFlag,
  useNumberFlag,
  useVariation,
  useFlexFlagClient,
  FlexFlagProvider,
  withFeatureFlag,
  FeatureGate
} from './react';
