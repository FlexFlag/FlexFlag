/**
 * FlexFlag JavaScript SDK - Basic Usage Example
 */

const { FlexFlagClient, MemoryCache } = require('@flexflag/client');

async function basicExample() {
  console.log('🚀 FlexFlag SDK Basic Example\n');
  
  // Initialize FlexFlag client with configuration
  const client = new FlexFlagClient({
    apiKey: 'ff_production_your_api_key_here',
    baseUrl: 'http://localhost:8080', // or https://api.flexflag.io
    environment: 'production',
    
    // Cache configuration
    cache: {
      enabled: true,
      ttl: 300000, // 5 minutes
      maxSize: 500,
      storage: 'memory'
    },
    
    // Connection settings
    connection: {
      mode: 'streaming', // Server-Sent Events (SSE) for real-time updates (default)
      pollingInterval: 30000, // Only used if mode is 'polling'
      timeout: 5000,
      retryAttempts: 3
    },
    
    // Offline support
    offline: {
      enabled: true,
      defaultFlags: {
        'dark-mode': false,
        'beta-features': false,
        'premium-plan': 'standard'
      }
    },
    
    // Event handlers
    events: {
      onReady: () => console.log('✅ FlexFlag SDK ready!'),
      onCacheHit: (flagKey) => console.log(`🎯 Cache hit for: ${flagKey}`),
      onCacheMiss: (flagKey) => console.log(`💥 Cache miss for: ${flagKey}`),
      onError: (error) => console.error('❌ FlexFlag error:', error.message)
    },
    
    // Logging
    logging: {
      level: 'info'
    }
  });
  
  try {
    // Wait for SDK to be ready
    await client.waitForReady();
    
    // Set user context
    client.setContext({
      userId: 'user_12345',
      attributes: {
        plan: 'premium',
        region: 'us-east',
        signupDate: '2023-01-15'
      }
    });
    
    console.log('\n📊 Evaluating feature flags...\n');
    
    // 1. Boolean flag evaluation
    const darkModeEnabled = await client.evaluate('dark-mode', undefined, false);
    console.log(`🌓 Dark mode: ${darkModeEnabled ? 'ON' : 'OFF'}`);
    
    // 2. String flag evaluation  
    const theme = await client.evaluate('ui-theme', undefined, 'light');
    console.log(`🎨 UI Theme: ${theme}`);
    
    // 3. Number flag evaluation
    const maxRetries = await client.evaluate('max-retries', undefined, 3);
    console.log(`🔄 Max retries: ${maxRetries}`);
    
    // 4. JSON flag evaluation
    const config = await client.evaluate('app-config', undefined, {});
    console.log(`⚙️  App config:`, config);
    
    // 5. Batch evaluation (more efficient for multiple flags)
    const batchResults = await client.evaluateBatch([
      'dark-mode',
      'ui-theme', 
      'max-retries',
      'beta-features'
    ]);
    console.log('\n📦 Batch evaluation results:', batchResults);
    
    // 6. A/B testing variation
    const variation = await client.getVariation('checkout-flow');
    console.log(`🧪 A/B test variation: ${variation || 'control'}`);
    
    // 7. Detailed evaluation with metadata
    const detailed = await client.evaluateWithDetails('premium-features');
    console.log('\n📋 Detailed evaluation:', {
      value: detailed.value,
      variation: detailed.variation,
      reason: detailed.reason,
      cacheHit: detailed.metadata?.cacheHit,
      source: detailed.metadata?.source,
      evaluationTime: detailed.metadata?.evaluationTime + 'ms'
    });
    
    // 8. Update context and re-evaluate
    console.log('\n🔄 Updating user context...');
    client.updateContext({
      attributes: {
        plan: 'enterprise', // Upgrade plan
        region: 'us-west'
      }
    });
    
    const updatedFeatures = await client.evaluate('premium-features');
    console.log(`💎 Premium features (after upgrade): ${updatedFeatures}`);
    
    // 9. Show cache performance metrics
    const metrics = client.getMetrics();
    console.log('\n📈 SDK Metrics:', {
      evaluations: metrics.evaluations,
      cacheHits: metrics.cacheHits,
      cacheMisses: metrics.cacheMisses,
      cacheHitRate: `${((metrics.cacheHits / metrics.evaluations) * 100).toFixed(1)}%`,
      averageLatency: `${metrics.averageLatency.toFixed(2)}ms`,
      networkRequests: metrics.networkRequests
    });
    
    // 10. Demonstrate real-time updates
    console.log('\n🔄 Listening for real-time flag updates...');
    client.on('update', (updatedFlags) => {
      console.log(`🚨 Flags updated: ${updatedFlags.join(', ')}`);
    });
    
    // Simulate some activity
    setTimeout(async () => {
      console.log('\n🧹 Clearing cache...');
      await client.clearCache();
      
      const flagAfterCacheClear = await client.evaluate('dark-mode');
      console.log(`🌓 Dark mode (after cache clear): ${flagAfterCacheClear}`);
      
      // Final metrics
      const finalMetrics = client.getMetrics();
      console.log('\n📊 Final Metrics:', finalMetrics);
      
      // Cleanup
      await client.close();
      console.log('\n👋 FlexFlag SDK closed gracefully');
    }, 2000);
    
  } catch (error) {
    console.error('💥 Error:', error.message);
  }
}

// Advanced usage with custom cache
async function advancedCacheExample() {
  console.log('\n🔧 Advanced Cache Configuration Example\n');
  
  // Custom cache configuration
  const client = new FlexFlagClient({
    apiKey: 'ff_production_your_api_key_here',
    baseUrl: 'http://localhost:8080',
    
    cache: {
      enabled: true,
      provider: new MemoryCache({
        ttl: 600000, // 10 minutes
        maxSize: 1000,
        compression: true
      }),
      keyPrefix: 'myapp:flexflag:'
    },
    
    performance: {
      evaluationMode: 'cached',
      batchRequests: true,
      prefetch: true,
      prefetchFlags: ['critical-feature', 'ui-theme', 'max-retries']
    }
  });
  
  await client.waitForReady();
  console.log('✅ Advanced cache client ready');
  
  // Evaluate flags - should be served from prefetch cache
  const results = await client.evaluateBatch(['critical-feature', 'ui-theme']);
  console.log('⚡ Prefetched flags (served from cache):', results);
  
  await client.close();
}

// Error handling example
async function errorHandlingExample() {
  console.log('\n🛡️  Error Handling Example\n');
  
  const client = new FlexFlagClient({
    apiKey: 'invalid_api_key',
    baseUrl: 'http://invalid-url',
    
    offline: {
      enabled: true,
      defaultFlags: {
        'fallback-feature': true,
        'emergency-mode': 'enabled'
      }
    },
    
    connection: {
      timeout: 1000, // Short timeout to trigger errors
      retryAttempts: 1
    }
  });
  
  // This should fall back to offline defaults
  const fallbackValue = await client.evaluate('fallback-feature', undefined, false);
  console.log(`🆘 Fallback feature (offline mode): ${fallbackValue}`);
  
  const emergencyMode = await client.evaluate('emergency-mode', undefined, 'disabled');
  console.log(`🚨 Emergency mode: ${emergencyMode}`);
  
  await client.close();
}

// Run examples
if (require.main === module) {
  (async () => {
    await basicExample();
    await advancedCacheExample();
    await errorHandlingExample();
  })().catch(console.error);
}

module.exports = {
  basicExample,
  advancedCacheExample,
  errorHandlingExample
};