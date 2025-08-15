/**
 * Vue 3 Composables for FlexFlag
 */

import { ref, reactive, onMounted, onUnmounted, watch, inject, provide, InjectionKey, App } from 'vue';
import { FlexFlagClient } from './client';
import { FlexFlagConfig, EvaluationContext, FlagValue } from './types';

// Injection key for FlexFlag client
const FlexFlagClientKey: InjectionKey<FlexFlagClient> = Symbol('FlexFlagClient');

/**
 * Vue plugin to install FlexFlag
 */
export function createFlexFlag(config: FlexFlagConfig) {
  const client = new FlexFlagClient(config);
  
  return {
    install(app: App) {
      app.provide(FlexFlagClientKey, client);
      app.config.globalProperties.$flexflag = client;
    }
  };
}

/**
 * Composable to get FlexFlag client
 */
export function useFlexFlagClient(): FlexFlagClient {
  const client = inject(FlexFlagClientKey);
  if (!client) {
    throw new Error('FlexFlag client not found. Make sure to install the FlexFlag plugin.');
  }
  return client;
}

/**
 * Composable to evaluate a feature flag
 */
export function useFeatureFlag(
  flagKey: string,
  defaultValue?: FlagValue,
  context?: EvaluationContext
) {
  const client = useFlexFlagClient();
  const value = ref<FlagValue>(defaultValue ?? null);
  const loading = ref(true);
  const error = ref<Error | null>(null);

  const evaluateFlag = async () => {
    try {
      loading.value = true;
      error.value = null;
      const result = await client.evaluate(flagKey, context, defaultValue);
      value.value = result;
    } catch (err) {
      error.value = err as Error;
      value.value = defaultValue ?? null;
    } finally {
      loading.value = false;
    }
  };

  const reload = () => {
    evaluateFlag();
  };

  onMounted(async () => {
    await evaluateFlag();

    // Listen for flag updates
    const handleUpdate = (updatedFlags: string[]) => {
      if (updatedFlags.includes(flagKey)) {
        evaluateFlag();
      }
    };

    client.on('update', handleUpdate);

    onUnmounted(() => {
      client.off('update', handleUpdate);
    });
  });

  // Watch for context changes
  if (context) {
    watch(
      () => context,
      () => {
        evaluateFlag();
      },
      { deep: true }
    );
  }

  return {
    value,
    loading,
    error,
    reload
  };
}

/**
 * Composable for boolean feature flags
 */
export function useBooleanFlag(
  flagKey: string,
  defaultValue: boolean = false,
  context?: EvaluationContext
) {
  const { value, loading, error, reload } = useFeatureFlag(flagKey, defaultValue, context);

  const enabled = ref(false);

  watch(
    value,
    (newValue) => {
      enabled.value = Boolean(newValue);
    },
    { immediate: true }
  );

  return {
    enabled,
    loading,
    error,
    reload
  };
}

/**
 * Composable for string feature flags
 */
export function useStringFlag(
  flagKey: string,
  defaultValue: string = '',
  context?: EvaluationContext
) {
  const { value, loading, error, reload } = useFeatureFlag(flagKey, defaultValue, context);

  const stringValue = ref(defaultValue);

  watch(
    value,
    (newValue) => {
      stringValue.value = String(newValue || defaultValue);
    },
    { immediate: true }
  );

  return {
    value: stringValue,
    loading,
    error,
    reload
  };
}

/**
 * Composable for number feature flags
 */
export function useNumberFlag(
  flagKey: string,
  defaultValue: number = 0,
  context?: EvaluationContext
) {
  const { value, loading, error, reload } = useFeatureFlag(flagKey, defaultValue, context);

  const numberValue = ref(defaultValue);

  watch(
    value,
    (newValue) => {
      numberValue.value = Number(newValue || defaultValue);
    },
    { immediate: true }
  );

  return {
    value: numberValue,
    loading,
    error,
    reload
  };
}

/**
 * Composable for A/B testing variations
 */
export function useVariation(
  flagKey: string,
  defaultVariation: string = 'control',
  context?: EvaluationContext
) {
  const client = useFlexFlagClient();
  const variation = ref<string>(defaultVariation);
  const loading = ref(true);
  const error = ref<Error | null>(null);

  const evaluateVariation = async () => {
    try {
      loading.value = true;
      error.value = null;
      const result = await client.getVariation(flagKey, context);
      variation.value = result || defaultVariation;
    } catch (err) {
      error.value = err as Error;
      variation.value = defaultVariation;
    } finally {
      loading.value = false;
    }
  };

  const reload = () => {
    evaluateVariation();
  };

  onMounted(async () => {
    await evaluateVariation();

    const handleUpdate = (updatedFlags: string[]) => {
      if (updatedFlags.includes(flagKey)) {
        evaluateVariation();
      }
    };

    client.on('update', handleUpdate);

    onUnmounted(() => {
      client.off('update', handleUpdate);
    });
  });

  if (context) {
    watch(
      () => context,
      () => {
        evaluateVariation();
      },
      { deep: true }
    );
  }

  return {
    variation,
    loading,
    error,
    reload
  };
}

/**
 * Composable for batch flag evaluation
 */
export function useBatchFlags(
  flagKeys: string[],
  context?: EvaluationContext
) {
  const client = useFlexFlagClient();
  const flags = reactive<Record<string, FlagValue>>({});
  const loading = ref(true);
  const error = ref<Error | null>(null);

  const evaluateFlags = async () => {
    try {
      loading.value = true;
      error.value = null;
      const results = await client.evaluateBatch(flagKeys, context);
      
      // Update flags reactively
      Object.keys(flags).forEach(key => delete flags[key]);
      Object.assign(flags, results);
    } catch (err) {
      error.value = err as Error;
    } finally {
      loading.value = false;
    }
  };

  const reload = () => {
    evaluateFlags();
  };

  onMounted(async () => {
    await evaluateFlags();

    const handleUpdate = (updatedFlags: string[]) => {
      const shouldReload = updatedFlags.some(flag => flagKeys.includes(flag));
      if (shouldReload) {
        evaluateFlags();
      }
    };

    client.on('update', handleUpdate);

    onUnmounted(() => {
      client.off('update', handleUpdate);
    });
  });

  if (context) {
    watch(
      () => context,
      () => {
        evaluateFlags();
      },
      { deep: true }
    );
  }

  return {
    flags,
    loading,
    error,
    reload
  };
}

/**
 * Composable for SDK metrics
 */
export function useFlexFlagMetrics() {
  const client = useFlexFlagClient();
  const metrics = ref(client.getMetrics());

  const refresh = () => {
    metrics.value = client.getMetrics();
  };

  const reset = () => {
    client.resetMetrics();
    metrics.value = client.getMetrics();
  };

  return {
    metrics,
    refresh,
    reset
  };
}

/**
 * Directive for conditional rendering based on feature flags
 */
export const vFeatureFlag = {
  async mounted(el: HTMLElement, binding: any) {
    const { value: flagKey, modifiers, arg } = binding;
    const client = inject(FlexFlagClientKey);
    
    if (!client) {
      console.error('FlexFlag client not found for v-feature-flag directive');
      return;
    }

    try {
      const result = await client.evaluate(flagKey, undefined, false);
      const enabled = Boolean(result);
      
      // Handle modifiers
      const shouldShow = modifiers.not ? !enabled : enabled;
      
      if (!shouldShow) {
        el.style.display = 'none';
      }
    } catch (error) {
      console.error('Error evaluating feature flag in directive:', error);
      el.style.display = 'none';
    }
  }
};

/**
 * Provide FlexFlag client to child components
 */
export function provideFlexFlag(config: FlexFlagConfig) {
  const client = new FlexFlagClient(config);
  provide(FlexFlagClientKey, client);
  return client;
}