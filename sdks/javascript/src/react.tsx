/**
 * React Integration for FlexFlag
 */

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { FlexFlagClient } from './client';
import { FlexFlagConfig, EvaluationContext, FlagValue } from './types';

// Context for FlexFlag client
const FlexFlagContext = createContext<FlexFlagClient | null>(null);

interface FlexFlagProviderProps {
  config: FlexFlagConfig;
  context?: EvaluationContext;
  children: ReactNode;
}

/**
 * FlexFlag Provider component
 */
export function FlexFlagProvider({ config, context, children }: FlexFlagProviderProps) {
  const [client] = useState(() => new FlexFlagClient(config));

  useEffect(() => {
    if (context) {
      client.setContext(context);
    }
  }, [client, context]);

  useEffect(() => {
    return () => {
      client.close();
    };
  }, [client]);

  return (
    <FlexFlagContext.Provider value={client}>
      {children}
    </FlexFlagContext.Provider>
  );
}

/**
 * Hook to get FlexFlag client
 */
export function useFlexFlagClient(): FlexFlagClient {
  const client = useContext(FlexFlagContext);
  if (!client) {
    throw new Error('useFlexFlagClient must be used within a FlexFlagProvider');
  }
  return client;
}

/**
 * Hook to evaluate a feature flag
 */
export function useFeatureFlag(
  flagKey: string,
  defaultValue?: FlagValue,
  context?: EvaluationContext
): {
  value: FlagValue;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const client = useFlexFlagClient();
  const [value, setValue] = useState<FlagValue>(defaultValue ?? null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const evaluateFlag = async () => {
    try {
      setLoading(true);
      setError(null);
      const result = await client.evaluate(flagKey, context, defaultValue);
      setValue(result);
    } catch (err) {
      setError(err as Error);
      setValue(defaultValue ?? null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    evaluateFlag();

    // Listen for flag updates
    const handleUpdate = (updatedFlags: string[]) => {
      if (updatedFlags.includes(flagKey)) {
        evaluateFlag();
      }
    };

    client.on('update', handleUpdate);

    return () => {
      client.off('update', handleUpdate);
    };
  }, [client, flagKey, context, defaultValue]);

  return {
    value,
    loading,
    error,
    reload: evaluateFlag
  };
}

/**
 * Hook for boolean feature flags
 */
export function useBooleanFlag(
  flagKey: string,
  defaultValue: boolean = false,
  context?: EvaluationContext
): {
  enabled: boolean;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const result = useFeatureFlag(flagKey, defaultValue, context);
  
  return {
    enabled: Boolean(result.value),
    loading: result.loading,
    error: result.error,
    reload: result.reload
  };
}

/**
 * Hook for string feature flags
 */
export function useStringFlag(
  flagKey: string,
  defaultValue: string = '',
  context?: EvaluationContext
): {
  value: string;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const result = useFeatureFlag(flagKey, defaultValue, context);
  
  return {
    value: String(result.value || defaultValue),
    loading: result.loading,
    error: result.error,
    reload: result.reload
  };
}

/**
 * Hook for number feature flags
 */
export function useNumberFlag(
  flagKey: string,
  defaultValue: number = 0,
  context?: EvaluationContext
): {
  value: number;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const result = useFeatureFlag(flagKey, defaultValue, context);
  
  return {
    value: Number(result.value || defaultValue),
    loading: result.loading,
    error: result.error,
    reload: result.reload
  };
}

/**
 * Hook for A/B testing variations
 */
export function useVariation(
  flagKey: string,
  defaultVariation: string = 'control',
  context?: EvaluationContext
): {
  variation: string;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const client = useFlexFlagClient();
  const [variation, setVariation] = useState<string>(defaultVariation);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const evaluateVariation = async () => {
    try {
      setLoading(true);
      setError(null);
      const result = await client.getVariation(flagKey, context);
      setVariation(result || defaultVariation);
    } catch (err) {
      setError(err as Error);
      setVariation(defaultVariation);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    evaluateVariation();

    const handleUpdate = (updatedFlags: string[]) => {
      if (updatedFlags.includes(flagKey)) {
        evaluateVariation();
      }
    };

    client.on('update', handleUpdate);

    return () => {
      client.off('update', handleUpdate);
    };
  }, [client, flagKey, context, defaultVariation]);

  return {
    variation,
    loading,
    error,
    reload: evaluateVariation
  };
}

/**
 * Higher-order component for feature flag gating
 */
export function withFeatureFlag<P extends object>(
  flagKey: string,
  options: {
    fallback?: React.ComponentType<P>;
    defaultValue?: boolean;
    context?: EvaluationContext;
  } = {}
) {
  return function (WrappedComponent: React.ComponentType<P>) {
    const WithFeatureFlagComponent = (props: P) => {
      const { enabled, loading } = useBooleanFlag(
        flagKey,
        options.defaultValue,
        options.context
      );

      if (loading) {
        return null; // or loading spinner
      }

      if (!enabled) {
        if (options.fallback) {
          const FallbackComponent = options.fallback;
          return <FallbackComponent {...props} />;
        }
        return null;
      }

      return <WrappedComponent {...props} />;
    };

    WithFeatureFlagComponent.displayName = `withFeatureFlag(${WrappedComponent.displayName || WrappedComponent.name})`;

    return WithFeatureFlagComponent;
  };
}

/**
 * Component for conditional rendering based on feature flags
 */
interface FeatureGateProps {
  flagKey: string;
  defaultValue?: boolean;
  context?: EvaluationContext;
  fallback?: ReactNode;
  loading?: ReactNode;
  children: ReactNode;
}

export function FeatureGate({
  flagKey,
  defaultValue = false,
  context,
  fallback = null,
  loading: loadingComponent = null,
  children
}: FeatureGateProps) {
  const { enabled, loading } = useBooleanFlag(flagKey, defaultValue, context);

  if (loading) {
    return <>{loadingComponent}</>;
  }

  return <>{enabled ? children : fallback}</>;
}