import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';
import { FlexFlagClient, AsyncStorageLike } from './client';
import { FlexFlagConfig, FlagValue, EvaluationContext, ConfigResult } from './types';

// ─── Context ────────────────────────────────────────────────────────────────

interface FlexFlagContextValue {
  client: FlexFlagClient | null;
  ready: boolean;
}

const FlexFlagContext = createContext<FlexFlagContextValue>({ client: null, ready: false });

// ─── Provider ────────────────────────────────────────────────────────────────

interface FlexFlagProviderProps {
  config: FlexFlagConfig;
  /** Pass AsyncStorage from @react-native-async-storage/async-storage for persistence */
  storage?: AsyncStorageLike;
  children: React.ReactNode;
}

export function FlexFlagProvider({ config, storage, children }: FlexFlagProviderProps) {
  const [ready, setReady] = useState(false);
  const clientRef = useRef<FlexFlagClient | null>(null);

  useEffect(() => {
    const client = new FlexFlagClient(config);
    clientRef.current = client;

    client.on('ready', () => setReady(true));

    client.initialize(storage).catch(() => {
      // Even on error, mark ready so the app doesn't hang
      setReady(true);
    });

    return () => {
      client.close();
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <FlexFlagContext.Provider value={{ client: clientRef.current, ready }}>
      {children}
    </FlexFlagContext.Provider>
  );
}

// ─── Hooks ───────────────────────────────────────────────────────────────────

export function useFlexFlagClient(): FlexFlagClient | null {
  return useContext(FlexFlagContext).client;
}

interface FlagHookResult<T> {
  value: T;
  loading: boolean;
  error: Error | null;
  reload: () => void;
}

/**
 * Evaluate a boolean feature flag.
 *
 * @example
 * const { value: isEnabled } = useBooleanFlag('new-checkout', false);
 */
export function useBooleanFlag(
  flagKey: string,
  defaultValue = false,
  context?: EvaluationContext
): FlagHookResult<boolean> {
  return useFlagValue<boolean>(flagKey, defaultValue, context, (v) =>
    typeof v === 'boolean' ? v : defaultValue
  );
}

/**
 * Evaluate a string config flag.
 *
 * @example
 * const { value: theme } = useStringFlag('app-theme', 'light');
 */
export function useStringFlag(
  flagKey: string,
  defaultValue = '',
  context?: EvaluationContext
): FlagHookResult<string> {
  return useFlagValue<string>(flagKey, defaultValue, context, (v) =>
    typeof v === 'string' ? v : defaultValue
  );
}

/**
 * Evaluate a number config flag.
 *
 * @example
 * const { value: timeout } = useNumberFlag('request-timeout', 5000);
 */
export function useNumberFlag(
  flagKey: string,
  defaultValue = 0,
  context?: EvaluationContext
): FlagHookResult<number> {
  return useFlagValue<number>(flagKey, defaultValue, context, (v) =>
    typeof v === 'number' ? v : defaultValue
  );
}

/**
 * Evaluate a JSON config flag.
 *
 * @example
 * const { value: config } = useJSONFlag<OnboardingConfig>('onboarding-config', defaults);
 */
export function useJSONFlag<T>(
  flagKey: string,
  defaultValue: T,
  context?: EvaluationContext
): FlagHookResult<T> {
  return useFlagValue<T>(flagKey, defaultValue, context, (v) =>
    v !== null && typeof v === 'object' ? (v as unknown as T) : defaultValue
  );
}

/**
 * Fetch all remote config values for the current environment.
 * Best called at app startup to warm the config cache.
 *
 * @example
 * const { config } = useRemoteConfig();
 * const buttonColor = config?.config['button-color'] ?? '#007AFF';
 */
export function useRemoteConfig(context?: EvaluationContext): {
  config: ConfigResult | null;
  loading: boolean;
  error: Error | null;
  reload: () => void;
} {
  const { client } = useContext(FlexFlagContext);
  const [config, setConfig] = useState<ConfigResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetch = useCallback(async () => {
    if (!client) return;
    setLoading(true);
    setError(null);
    try {
      const result = await client.getConfig(context);
      setConfig(result);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch config'));
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { config, loading, error, reload: fetch };
}

// ─── FeatureGate component ────────────────────────────────────────────────────

interface FeatureGateProps {
  flagKey: string;
  context?: EvaluationContext;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

/**
 * Conditionally renders children based on a boolean feature flag.
 *
 * @example
 * <FeatureGate flagKey="new-checkout" fallback={<OldCheckout />}>
 *   <NewCheckout />
 * </FeatureGate>
 */
export function FeatureGate({ flagKey, context, fallback = null, children }: FeatureGateProps) {
  const { value, loading } = useBooleanFlag(flagKey, false, context);
  if (loading) return <>{fallback}</>;
  return <>{value ? children : fallback}</>;
}

// ─── Internal helper ─────────────────────────────────────────────────────────

function useFlagValue<T>(
  flagKey: string,
  defaultValue: T,
  context: EvaluationContext | undefined,
  coerce: (v: FlagValue) => T
): FlagHookResult<T> {
  const { client } = useContext(FlexFlagContext);
  const [value, setValue] = useState<T>(defaultValue);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetch = useCallback(async () => {
    if (!client) return;
    setLoading(true);
    setError(null);
    try {
      const result = await client.evaluate(flagKey, context, defaultValue as FlagValue);
      setValue(coerce(result.value));
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Evaluation failed'));
      setValue(defaultValue);
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client, flagKey]);

  useEffect(() => {
    fetch();

    // Re-evaluate when the client broadcasts an update
    const handler = () => { fetch(); };
    client?.on('update', handler);
    return () => { client?.off('update', handler); };
  }, [client, fetch]);

  return { value, loading, error, reload: fetch };
}
