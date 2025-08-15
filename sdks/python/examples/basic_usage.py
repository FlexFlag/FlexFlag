"""
FlexFlag Python SDK - Basic Usage Example

This example demonstrates how to use the FlexFlag Python SDK
with intelligent caching and offline support.
"""

import asyncio
import time
from typing import Dict, Any

from flexflag import (
    FlexFlagClient, 
    FlexFlagConfig,
    EvaluationContext,
    CacheConfig,
    ConnectionConfig,
    OfflineConfig,
    EventHandlers,
    MemoryCache
)


async def basic_example():
    """Basic FlexFlag SDK usage example"""
    print("🚀 FlexFlag Python SDK Basic Example\n")
    
    # Configure event handlers
    def on_ready():
        print("✅ FlexFlag SDK ready!")
    
    def on_cache_hit(flag_key: str):
        print(f"🎯 Cache hit for: {flag_key}")
    
    def on_cache_miss(flag_key: str):
        print(f"💥 Cache miss for: {flag_key}")
    
    def on_error(error: Exception):
        print(f"❌ FlexFlag error: {error}")
    
    # Initialize FlexFlag client
    config = FlexFlagConfig(
        api_key="ff_production_your_api_key_here",
        base_url="http://localhost:8080",  # or https://api.flexflag.io
        environment="production",
        
        # Cache configuration
        cache=CacheConfig(
            enabled=True,
            ttl=300,  # 5 minutes
            max_size=500,
            storage="memory"
        ),
        
        # Connection settings
        connection=ConnectionConfig(
            mode="polling",  # Use polling for this example
            polling_interval=30,
            timeout=5,
            retry_attempts=3
        ),
        
        # Offline support
        offline=OfflineConfig(
            enabled=True,
            default_flags={
                "dark-mode": False,
                "beta-features": False,
                "max-retries": 3,
                "ui-theme": "light"
            }
        ),
        
        # Event handlers
        events=EventHandlers(
            on_ready=on_ready,
            on_cache_hit=on_cache_hit,
            on_cache_miss=on_cache_miss,
            on_error=on_error
        ),
        
        # Logging
        log_level="INFO"
    )
    
    client = FlexFlagClient(config)
    
    try:
        # Wait for SDK to be ready
        await client.wait_for_ready()
        
        # Set user context
        context = EvaluationContext(
            user_id="user_12345",
            attributes={
                "plan": "premium",
                "region": "us-east",
                "signup_date": "2023-01-15"
            }
        )
        
        client.set_context(context)
        
        print("\n📊 Evaluating feature flags...\n")
        
        # 1. Boolean flag evaluation
        dark_mode = await client.evaluate("dark-mode", default_value=False)
        print(f"🌓 Dark mode: {'ON' if dark_mode else 'OFF'}")
        
        # 2. String flag evaluation
        theme = await client.evaluate("ui-theme", default_value="light")
        print(f"🎨 UI Theme: {theme}")
        
        # 3. Number flag evaluation
        max_retries = await client.evaluate("max-retries", default_value=3)
        print(f"🔄 Max retries: {max_retries}")
        
        # 4. JSON flag evaluation
        app_config = await client.evaluate("app-config", default_value={})
        print(f"⚙️  App config: {app_config}")
        
        # 5. Batch evaluation (more efficient for multiple flags)
        batch_results = await client.evaluate_batch([
            "dark-mode",
            "ui-theme",
            "max-retries", 
            "beta-features"
        ])
        print(f"\n📦 Batch evaluation results: {batch_results}")
        
        # 6. A/B testing variation
        variation = await client.get_variation("checkout-flow")
        print(f"🧪 A/B test variation: {variation or 'control'}")
        
        # 7. Detailed evaluation with metadata
        detailed = await client.evaluate_with_details("premium-features")
        print(f"\n📋 Detailed evaluation:")
        print(f"  Value: {detailed.value}")
        print(f"  Variation: {detailed.variation}")
        print(f"  Reason: {detailed.reason.value}")
        if detailed.metadata:
            print(f"  Cache hit: {detailed.metadata.cache_hit}")
            print(f"  Source: {detailed.metadata.source}")
            print(f"  Evaluation time: {detailed.metadata.evaluation_time_ms:.2f}ms")
        
        # 8. Update context and re-evaluate
        print("\n🔄 Updating user context...")
        client.update_context({
            "attributes": {
                "plan": "enterprise",  # Upgrade plan
                "region": "us-west"
            }
        })
        
        updated_features = await client.evaluate("premium-features")
        print(f"💎 Premium features (after upgrade): {updated_features}")
        
        # 9. Show cache performance metrics
        metrics = client.get_metrics()
        cache_hit_rate = (metrics.cache_hits / metrics.evaluations * 100) if metrics.evaluations > 0 else 0
        
        print(f"\n📈 SDK Metrics:")
        print(f"  Evaluations: {metrics.evaluations}")
        print(f"  Cache hits: {metrics.cache_hits}")
        print(f"  Cache misses: {metrics.cache_misses}")
        print(f"  Cache hit rate: {cache_hit_rate:.1f}%")
        print(f"  Average latency: {metrics.average_latency_ms:.2f}ms")
        print(f"  Network requests: {metrics.network_requests}")
        
        # 10. Clear cache and re-evaluate
        print("\n🧹 Clearing cache...")
        await client.clear_cache()
        
        flag_after_clear = await client.evaluate("dark-mode")
        print(f"🌓 Dark mode (after cache clear): {'ON' if flag_after_clear else 'OFF'}")
        
        # Final metrics
        final_metrics = client.get_metrics()
        print(f"\n📊 Final Metrics: {final_metrics}")
        
    except Exception as e:
        print(f"💥 Error: {e}")
    finally:
        # Cleanup
        await client.close()
        print("\n👋 FlexFlag SDK closed gracefully")


async def advanced_cache_example():
    """Advanced cache configuration example"""
    print("\n🔧 Advanced Cache Configuration Example\n")
    
    # Custom memory cache with specific settings
    custom_cache = MemoryCache(CacheConfig(
        ttl=600,  # 10 minutes
        max_size=1000,
        compression=False  # Disable compression for speed
    ))
    
    config = FlexFlagConfig(
        api_key="ff_production_your_api_key_here",
        base_url="http://localhost:8080",
        cache=CacheConfig(
            enabled=True,
            storage="custom"  # Use custom cache provider
        )
    )
    
    client = FlexFlagClient(config)
    client.cache = custom_cache  # Set custom cache
    
    await client.wait_for_ready()
    print("✅ Advanced cache client ready")
    
    # Evaluate flags - should use custom cache
    results = await client.evaluate_batch(["critical-feature", "ui-theme"])
    print(f"⚡ Custom cache results: {results}")
    
    await client.close()


async def error_handling_example():
    """Error handling and offline mode example"""
    print("\n🛡️  Error Handling and Offline Mode Example\n")
    
    config = FlexFlagConfig(
        api_key="invalid_api_key",
        base_url="http://invalid-url",
        
        # Offline configuration with fallback values
        offline=OfflineConfig(
            enabled=True,
            default_flags={
                "fallback-feature": True,
                "emergency-mode": "enabled",
                "maintenance-page": False
            }
        ),
        
        connection=ConnectionConfig(
            timeout=1,  # Short timeout to trigger errors
            retry_attempts=1
        )
    )
    
    client = FlexFlagClient(config)
    
    try:
        # These should fall back to offline defaults
        fallback_value = await client.evaluate("fallback-feature", default_value=False)
        print(f"🆘 Fallback feature (offline mode): {fallback_value}")
        
        emergency_mode = await client.evaluate("emergency-mode", default_value="disabled")
        print(f"🚨 Emergency mode: {emergency_mode}")
        
        maintenance = await client.evaluate("maintenance-page", default_value=True)
        print(f"🔧 Maintenance page: {'ON' if maintenance else 'OFF'}")
        
    except Exception as e:
        print(f"💥 Handled error: {e}")
    finally:
        await client.close()


async def django_integration_example():
    """Django integration example"""
    print("\n🌐 Django Integration Example\n")
    
    # This would typically be in Django settings
    FLEXFLAG_CONFIG = {
        "api_key": "ff_production_your_api_key_here",
        "base_url": "http://localhost:8080",
        "environment": "production",
        "cache": {
            "enabled": True,
            "ttl": 300,
            "storage": "memory"
        }
    }
    
    # Initialize client (would be done in Django app initialization)
    config = FlexFlagConfig(**FLEXFLAG_CONFIG)
    client = FlexFlagClient(config)
    
    # Simulate Django request context
    def get_user_context(request):
        """Extract user context from Django request"""
        return EvaluationContext(
            user_id=str(request.user.id) if hasattr(request, 'user') and request.user.is_authenticated else None,
            attributes={
                "is_staff": getattr(request.user, 'is_staff', False),
                "is_premium": getattr(request.user, 'is_premium', False),
                "user_agent": request.META.get('HTTP_USER_AGENT', ''),
                "ip_address": request.META.get('REMOTE_ADDR', '')
            }
        )
    
    # Mock Django request object
    class MockRequest:
        class User:
            id = 123
            is_authenticated = True
            is_staff = False
            is_premium = True
        
        user = User()
        META = {
            'HTTP_USER_AGENT': 'Mozilla/5.0 (example browser)',
            'REMOTE_ADDR': '192.168.1.1'
        }
    
    request = MockRequest()
    context = get_user_context(request)
    
    # Django view example
    await client.wait_for_ready()
    
    # Feature flag checks in Django view
    show_new_dashboard = await client.evaluate("new-dashboard", context=context, default_value=False)
    max_uploads = await client.evaluate("max-file-uploads", context=context, default_value=5)
    
    print(f"👤 User context: {context}")
    print(f"🆕 Show new dashboard: {show_new_dashboard}")
    print(f"📁 Max file uploads: {max_uploads}")
    
    await client.close()


def performance_benchmark():
    """Performance benchmark example"""
    print("\n⚡ Performance Benchmark Example\n")
    
    async def run_benchmark():
        config = FlexFlagConfig(
            api_key="ff_production_your_api_key_here",
            base_url="http://localhost:8080",
            cache=CacheConfig(enabled=True, ttl=600)
        )
        
        client = FlexFlagClient(config)
        await client.wait_for_ready()
        
        # Benchmark single evaluations
        iterations = 1000
        start_time = time.time()
        
        for i in range(iterations):
            await client.evaluate("benchmark-flag", default_value=False)
        
        end_time = time.time()
        total_time = (end_time - start_time) * 1000  # Convert to ms
        avg_time = total_time / iterations
        
        print(f"🏃 Single evaluation benchmark:")
        print(f"  Iterations: {iterations}")
        print(f"  Total time: {total_time:.2f}ms")
        print(f"  Average time per evaluation: {avg_time:.3f}ms")
        print(f"  Evaluations per second: {1000/avg_time:.0f}")
        
        # Benchmark batch evaluations
        batch_flags = [f"batch-flag-{i}" for i in range(10)]
        batch_iterations = 100
        
        start_time = time.time()
        for i in range(batch_iterations):
            await client.evaluate_batch(batch_flags)
        end_time = time.time()
        
        batch_total_time = (end_time - start_time) * 1000
        batch_avg_time = batch_total_time / batch_iterations
        
        print(f"\n📦 Batch evaluation benchmark:")
        print(f"  Batch size: {len(batch_flags)} flags")
        print(f"  Iterations: {batch_iterations}")
        print(f"  Total time: {batch_total_time:.2f}ms")
        print(f"  Average time per batch: {batch_avg_time:.3f}ms")
        print(f"  Batches per second: {1000/batch_avg_time:.0f}")
        
        # Show final metrics
        metrics = client.get_metrics()
        cache_hit_rate = (metrics.cache_hits / metrics.evaluations * 100) if metrics.evaluations > 0 else 0
        print(f"\n📊 Benchmark metrics:")
        print(f"  Cache hit rate: {cache_hit_rate:.1f}%")
        print(f"  Average latency: {metrics.average_latency_ms:.3f}ms")
        
        await client.close()
    
    asyncio.run(run_benchmark())


if __name__ == "__main__":
    # Run examples
    async def main():
        await basic_example()
        await advanced_cache_example() 
        await error_handling_example()
        await django_integration_example()
    
    # Run async examples
    asyncio.run(main())
    
    # Run performance benchmark
    performance_benchmark()