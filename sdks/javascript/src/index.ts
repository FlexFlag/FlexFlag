/**
 * FlexFlag JavaScript/TypeScript SDK
 * High-performance feature flag client with local caching
 *
 * For React integration: import from 'flexflag-client/react'
 * For Vue integration: import from 'flexflag-client/vue'
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

// Default export for convenience
import { FlexFlagClient } from './client';
export default FlexFlagClient;