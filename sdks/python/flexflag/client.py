"""
FlexFlag Python SDK Client

High-performance feature flag client with intelligent caching, offline support,
and real-time updates.
"""

import asyncio
import json
import time
from typing import Any, Dict, List, Optional, Union
from datetime import datetime
from urllib.parse import urljoin
import aiohttp

from .types import (
    FlexFlagConfig, EvaluationContext, FlagValue, EvaluationResult,
    EvaluationReason, EvaluationMetadata, SDKMetrics, DefaultLogger,
    BatchEvaluationRequest, BatchEvaluationResponse
)
from .cache import create_cache_provider, CacheProvider


class FlexFlagClient:
    """
    FlexFlag Python SDK Client
    
    Provides high-performance feature flag evaluation with intelligent caching,
    offline support, and real-time updates.
    """
    
    def __init__(self, config: FlexFlagConfig):
        self.config = config
        self.context: Optional[EvaluationContext] = None
        self.logger = DefaultLogger(config.log_level)
        
        # Initialize cache
        if config.cache.enabled:
            self.cache = create_cache_provider(config.cache)
        else:
            self.cache = None
        
        # Initialize metrics
        self.metrics = SDKMetrics()
        
        # Connection state
        self._session: Optional[aiohttp.ClientSession] = None
        self._ready = asyncio.Event()
        self._offline_mode = False
        self._polling_task: Optional[asyncio.Task] = None
        
        # Event handlers
        self.events = config.events
        
        self.logger.info("FlexFlag SDK initialized")
    
    async def wait_for_ready(self, timeout: float = 10.0) -> bool:
        """Wait for SDK to be ready"""
        try:
            await asyncio.wait_for(self._initialize(), timeout=timeout)
            return True
        except asyncio.TimeoutError:
            self.logger.error("SDK initialization timed out")
            return False
    
    async def _initialize(self):
        """Initialize the SDK"""
        try:
            # Create HTTP session
            self._session = aiohttp.ClientSession(
                timeout=aiohttp.ClientTimeout(total=self.config.connection.timeout),
                headers={
                    "Authorization": f"Bearer {self.config.api_key}",
                    "Content-Type": "application/json",
                    "User-Agent": "FlexFlag-Python-SDK/1.0.0",
                    **self.config.connection.headers
                }
            )
            
            # Test connection if not in offline mode
            if self.config.connection.mode != "offline":
                await self._test_connection()
            
            # Start polling if configured
            if self.config.connection.mode == "polling":
                self._polling_task = asyncio.create_task(self._start_polling())
            
            self._ready.set()
            
            if self.events.on_ready:
                self.events.on_ready()
            
            self.logger.info("FlexFlag SDK ready")
            
        except Exception as e:
            self.logger.error(f"SDK initialization failed: {e}")
            if self.config.offline.enabled:
                self.logger.info("Switching to offline mode")
                self._offline_mode = True
                self._ready.set()
            else:
                raise
    
    async def _test_connection(self):
        """Test connection to FlexFlag API"""
        url = urljoin(self.config.base_url, "/api/v1/health")
        
        for attempt in range(self.config.connection.retry_attempts):
            try:
                async with self._session.get(url) as response:
                    if response.status == 200:
                        return
                    elif response.status == 401:
                        raise Exception("Invalid API key")
                    else:
                        raise Exception(f"API returned status {response.status}")
            
            except Exception as e:
                if attempt == self.config.connection.retry_attempts - 1:
                    raise e
                
                delay = self.config.connection.retry_delay
                if self.config.connection.exponential_backoff:
                    delay *= (2 ** attempt)
                
                self.logger.warning(f"Connection test failed (attempt {attempt + 1}): {e}")
                await asyncio.sleep(delay)
    
    async def _start_polling(self):
        """Start polling for flag updates"""
        while True:
            try:
                await asyncio.sleep(self.config.connection.polling_interval)
                # TODO: Implement flag polling logic
                # This would fetch updated flags and update cache
            except asyncio.CancelledError:
                break
            except Exception as e:
                self.logger.error(f"Polling error: {e}")
    
    def set_context(self, context: EvaluationContext):
        """Set evaluation context"""
        self.context = context
        self.logger.debug(f"Context set: {context}")
    
    def update_context(self, updates: Dict[str, Any]):
        """Update evaluation context"""
        if self.context is None:
            self.context = EvaluationContext()
        
        if "user_id" in updates:
            self.context.user_id = updates["user_id"]
        
        if "attributes" in updates:
            self.context.attributes.update(updates["attributes"])
        
        if "device" in updates:
            self.context.device = updates["device"]
        
        if "session" in updates:
            self.context.session = updates["session"]
        
        self.logger.debug(f"Context updated: {self.context}")
    
    async def evaluate(
        self, 
        flag_key: str, 
        context: Optional[EvaluationContext] = None,
        default_value: Optional[FlagValue] = None
    ) -> FlagValue:
        """Evaluate a feature flag"""
        result = await self.evaluate_with_details(flag_key, context, default_value)
        return result.value
    
    async def evaluate_with_details(
        self,
        flag_key: str,
        context: Optional[EvaluationContext] = None,
        default_value: Optional[FlagValue] = None
    ) -> EvaluationResult:
        """Evaluate a feature flag with detailed result"""
        start_time = time.time()
        evaluation_context = context or self.context
        
        try:
            # Check cache first
            cache_key = self._build_cache_key(flag_key, evaluation_context)
            cached_result = await self._get_from_cache(cache_key)
            
            if cached_result is not None:
                self.metrics.evaluations += 1
                self.metrics.cache_hits += 1
                
                if self.events.on_cache_hit:
                    self.events.on_cache_hit(flag_key)
                
                evaluation_time = (time.time() - start_time) * 1000
                self._update_average_latency(evaluation_time)
                
                return EvaluationResult(
                    value=cached_result,
                    reason=EvaluationReason.CACHED,
                    metadata=EvaluationMetadata(
                        timestamp=datetime.now(),
                        cache_hit=True,
                        evaluation_time_ms=evaluation_time,
                        source="cache"
                    )
                )
            
            # Cache miss - fetch from API
            if self.events.on_cache_miss:
                self.events.on_cache_miss(flag_key)
            
            self.metrics.cache_misses += 1
            
            # Check if offline mode
            if self._offline_mode or self.config.connection.mode == "offline":
                return await self._evaluate_offline(flag_key, default_value, start_time)
            
            # Fetch from API
            result = await self._evaluate_remote(flag_key, evaluation_context, default_value)
            
            # Cache the result
            await self._set_cache(cache_key, result.value)
            
            self.metrics.evaluations += 1
            evaluation_time = (time.time() - start_time) * 1000
            self._update_average_latency(evaluation_time)
            
            if self.events.on_evaluation:
                self.events.on_evaluation(flag_key, result.value)
            
            return result
            
        except Exception as e:
            self.logger.error(f"Flag evaluation failed for {flag_key}: {e}")
            self.metrics.errors += 1
            
            if self.events.on_error:
                self.events.on_error(e)
            
            # Fall back to offline mode
            return await self._evaluate_offline(flag_key, default_value, start_time)
    
    async def evaluate_batch(
        self,
        flag_keys: List[str],
        context: Optional[EvaluationContext] = None
    ) -> Dict[str, FlagValue]:
        """Evaluate multiple flags in a batch"""
        evaluation_context = context or self.context
        results = {}
        
        # Check cache for all flags first
        cache_hits = {}
        cache_misses = []
        
        for flag_key in flag_keys:
            cache_key = self._build_cache_key(flag_key, evaluation_context)
            cached_value = await self._get_from_cache(cache_key)
            
            if cached_value is not None:
                cache_hits[flag_key] = cached_value
                self.metrics.cache_hits += 1
                
                if self.events.on_cache_hit:
                    self.events.on_cache_hit(flag_key)
            else:
                cache_misses.append(flag_key)
                self.metrics.cache_misses += 1
                
                if self.events.on_cache_miss:
                    self.events.on_cache_miss(flag_key)
        
        results.update(cache_hits)
        
        # Fetch cache misses from API
        if cache_misses and not self._offline_mode and self.config.connection.mode != "offline":
            try:
                remote_results = await self._evaluate_batch_remote(cache_misses, evaluation_context)
                
                # Cache the results
                for flag_key, value in remote_results.items():
                    cache_key = self._build_cache_key(flag_key, evaluation_context)
                    await self._set_cache(cache_key, value)
                    
                    if self.events.on_evaluation:
                        self.events.on_evaluation(flag_key, value)
                
                results.update(remote_results)
                
            except Exception as e:
                self.logger.error(f"Batch evaluation failed: {e}")
                self.metrics.errors += 1
                
                if self.events.on_error:
                    self.events.on_error(e)
                
                # Fall back to offline defaults for cache misses
                for flag_key in cache_misses:
                    offline_value = self.config.offline.default_flags.get(flag_key)
                    if offline_value is not None:
                        results[flag_key] = offline_value
        
        # Handle remaining cache misses with offline defaults
        if self._offline_mode or self.config.connection.mode == "offline":
            for flag_key in cache_misses:
                offline_value = self.config.offline.default_flags.get(flag_key)
                if offline_value is not None:
                    results[flag_key] = offline_value
        
        self.metrics.evaluations += len(flag_keys)
        return results
    
    async def get_variation(
        self,
        flag_key: str,
        context: Optional[EvaluationContext] = None
    ) -> Optional[str]:
        """Get variation name for A/B testing"""
        result = await self.evaluate_with_details(flag_key, context)
        return result.variation
    
    async def _evaluate_remote(
        self,
        flag_key: str,
        context: Optional[EvaluationContext],
        default_value: Optional[FlagValue]
    ) -> EvaluationResult:
        """Evaluate flag remotely via API"""
        url = urljoin(self.config.base_url, "/api/v1/evaluate")
        
        payload = {
            "flagKey": flag_key,
            "context": self._context_to_dict(context),
            "environment": self.config.environment
        }
        
        self.metrics.network_requests += 1
        
        for attempt in range(self.config.connection.retry_attempts):
            try:
                async with self._session.post(url, json=payload) as response:
                    if response.status == 200:
                        data = await response.json()
                        return EvaluationResult(
                            value=data.get("value", default_value),
                            variation=data.get("variation"),
                            reason=EvaluationReason.TARGETING_MATCH if data.get("matched") else EvaluationReason.DEFAULT,
                            metadata=EvaluationMetadata(
                                timestamp=datetime.now(),
                                cache_hit=False,
                                evaluation_time_ms=(time.time() * 1000),
                                source="remote",
                                request_id=data.get("requestId")
                            )
                        )
                    elif response.status == 404:
                        return EvaluationResult(
                            value=default_value,
                            reason=EvaluationReason.DEFAULT
                        )
                    else:
                        raise Exception(f"API returned status {response.status}")
            
            except Exception as e:
                if attempt == self.config.connection.retry_attempts - 1:
                    raise e
                
                delay = self.config.connection.retry_delay
                if self.config.connection.exponential_backoff:
                    delay *= (2 ** attempt)
                
                await asyncio.sleep(delay)
        
        # Should not reach here
        raise Exception("All retry attempts exhausted")
    
    async def _evaluate_batch_remote(
        self,
        flag_keys: List[str],
        context: Optional[EvaluationContext]
    ) -> Dict[str, FlagValue]:
        """Evaluate multiple flags remotely via batch API"""
        url = urljoin(self.config.base_url, "/api/v1/evaluate/batch")
        
        payload = {
            "flags": flag_keys,
            "context": self._context_to_dict(context),
            "environment": self.config.environment
        }
        
        self.metrics.network_requests += 1
        
        async with self._session.post(url, json=payload) as response:
            if response.status == 200:
                data = await response.json()
                return data.get("flags", {})
            else:
                raise Exception(f"Batch API returned status {response.status}")
    
    async def _evaluate_offline(
        self,
        flag_key: str,
        default_value: Optional[FlagValue],
        start_time: float
    ) -> EvaluationResult:
        """Evaluate flag in offline mode"""
        offline_value = self.config.offline.default_flags.get(flag_key, default_value)
        evaluation_time = (time.time() - start_time) * 1000
        
        return EvaluationResult(
            value=offline_value,
            reason=EvaluationReason.OFFLINE,
            metadata=EvaluationMetadata(
                timestamp=datetime.now(),
                cache_hit=False,
                evaluation_time_ms=evaluation_time,
                source="offline"
            )
        )
    
    def _build_cache_key(
        self,
        flag_key: str,
        context: Optional[EvaluationContext]
    ) -> str:
        """Build cache key including context"""
        if context is None:
            return f"{flag_key}::{self.config.environment}"
        
        context_str = json.dumps(self._context_to_dict(context), sort_keys=True)
        return f"{flag_key}::{context_str}::{self.config.environment}"
    
    def _context_to_dict(self, context: Optional[EvaluationContext]) -> Dict[str, Any]:
        """Convert context to dictionary"""
        if context is None:
            return {}
        
        return {
            "userId": context.user_id,
            "attributes": context.attributes or {},
            "device": context.device or {},
            "session": context.session or {}
        }
    
    async def _get_from_cache(self, key: str) -> Optional[FlagValue]:
        """Get value from cache"""
        if self.cache is None:
            return None
        
        try:
            return await self.cache.get(key)
        except Exception as e:
            self.logger.warning(f"Cache get failed: {e}")
            return None
    
    async def _set_cache(self, key: str, value: FlagValue):
        """Set value in cache"""
        if self.cache is None:
            return
        
        try:
            await self.cache.set(key, value)
        except Exception as e:
            self.logger.warning(f"Cache set failed: {e}")
    
    async def clear_cache(self):
        """Clear all cached values"""
        if self.cache:
            await self.cache.clear()
            self.logger.info("Cache cleared")
    
    def get_metrics(self) -> SDKMetrics:
        """Get SDK performance metrics"""
        return self.metrics
    
    def _update_average_latency(self, latency_ms: float):
        """Update average latency calculation"""
        if self.metrics.evaluations == 0:
            self.metrics.average_latency_ms = latency_ms
        else:
            # Running average
            total_evaluations = self.metrics.evaluations + 1
            self.metrics.average_latency_ms = (
                (self.metrics.average_latency_ms * self.metrics.evaluations + latency_ms) 
                / total_evaluations
            )
    
    async def close(self):
        """Close SDK and cleanup resources"""
        if self._polling_task:
            self._polling_task.cancel()
            try:
                await self._polling_task
            except asyncio.CancelledError:
                pass
        
        if self._session:
            await self._session.close()
        
        if hasattr(self.cache, 'close'):
            await self.cache.close()
        
        self.logger.info("FlexFlag SDK closed")
    
    async def __aenter__(self):
        """Async context manager entry"""
        await self.wait_for_ready()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        await self.close()