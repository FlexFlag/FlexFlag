"""
FlexFlag FastAPI Integration

Dependencies and utilities for integrating FlexFlag with FastAPI applications.
"""

import asyncio
from typing import Optional, Dict, Any, Callable
from contextlib import asynccontextmanager

try:
    from fastapi import Depends, Request, HTTPException
    from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
    FASTAPI_AVAILABLE = True
except ImportError:
    FASTAPI_AVAILABLE = False

from ..client import FlexFlagClient
from ..types import FlexFlagConfig, EvaluationContext, FlagValue


class FlexFlagDependency:
    """FastAPI dependency for FlexFlag client injection"""
    
    def __init__(self, config: FlexFlagConfig):
        self.config = config
        self._client: Optional[FlexFlagClient] = None
        self._initialization_lock = asyncio.Lock()
    
    async def get_client(self) -> FlexFlagClient:
        """Get or create FlexFlag client instance"""
        if self._client is None:
            async with self._initialization_lock:
                if self._client is None:
                    self._client = FlexFlagClient(self.config)
                    await self._client.wait_for_ready()
        return self._client
    
    async def __call__(self) -> FlexFlagClient:
        """FastAPI dependency callable"""
        return await self.get_client()
    
    async def close(self):
        """Close the client connection"""
        if self._client:
            await self._client.close()


class FlexFlagContext:
    """FastAPI dependency for evaluation context"""
    
    def __init__(self, context_builder: Optional[Callable[[Request], EvaluationContext]] = None):
        self.context_builder = context_builder or self._default_context_builder
    
    def _default_context_builder(self, request: Request) -> EvaluationContext:
        """Default context builder from FastAPI request"""
        # Extract user information from request if available
        user_id = None
        attributes = {}
        
        # Check if user is available in request state
        if hasattr(request.state, 'user') and request.state.user:
            user = request.state.user
            user_id = str(getattr(user, 'id', ''))
            attributes.update({
                'username': getattr(user, 'username', ''),
                'email': getattr(user, 'email', ''),
                'is_active': getattr(user, 'is_active', True),
            })
            
            # Add custom user attributes if available
            if hasattr(user, 'get_flag_attributes'):
                custom_attrs = user.get_flag_attributes()
                if isinstance(custom_attrs, dict):
                    attributes.update(custom_attrs)
        
        # Add request metadata
        client_host = request.client.host if request.client else ''
        user_agent = request.headers.get('user-agent', '')
        
        attributes.update({
            'path': str(request.url.path),
            'method': request.method,
            'user_agent': user_agent,
            'ip_address': client_host,
            'is_secure': request.url.scheme == 'https',
            'query_params': dict(request.query_params),
        })
        
        return EvaluationContext(
            user_id=user_id,
            attributes=attributes
        )
    
    async def __call__(self, request: Request) -> EvaluationContext:
        """FastAPI dependency callable"""
        return self.context_builder(request)


def create_feature_flag_dependency(
    flag_key: str,
    default_value: FlagValue = False,
    required: bool = False
):
    """Create a FastAPI dependency for a specific feature flag"""
    
    async def feature_flag_dependency(
        client: FlexFlagClient = Depends(FlexFlagDependency),
        context: EvaluationContext = Depends(FlexFlagContext())
    ) -> FlagValue:
        """Evaluate specific feature flag"""
        try:
            result = await client.evaluate(flag_key, context, default_value)
            
            if required and not result:
                raise HTTPException(
                    status_code=403,
                    detail=f"Feature '{flag_key}' is not enabled"
                )
            
            return result
        except Exception as e:
            if required:
                raise HTTPException(
                    status_code=500,
                    detail=f"Failed to evaluate feature flag: {e}"
                )
            return default_value
    
    return feature_flag_dependency


def require_feature_flag(
    flag_key: str,
    default_enabled: bool = False,
    error_message: str = "Feature not available"
):
    """Decorator to require a feature flag for route access"""
    
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # Extract dependencies from kwargs
            client = None
            context = None
            request = None
            
            # Find client, context, and request in the function arguments
            for arg in args:
                if isinstance(arg, FlexFlagClient):
                    client = arg
                elif isinstance(arg, EvaluationContext):
                    context = arg
                elif hasattr(arg, 'method') and hasattr(arg, 'url'):  # Request-like object
                    request = arg
            
            # Also check kwargs
            for key, value in kwargs.items():
                if isinstance(value, FlexFlagClient):
                    client = value
                elif isinstance(value, EvaluationContext):
                    context = value
                elif key == 'request' or (hasattr(value, 'method') and hasattr(value, 'url')):
                    request = value
            
            # If we don't have client or context, try to create them
            if client is None:
                raise HTTPException(
                    status_code=500,
                    detail="FlexFlag client not available"
                )
            
            if context is None and request:
                context_builder = FlexFlagContext()
                context = await context_builder(request)
            
            try:
                enabled = await client.evaluate(flag_key, context, default_enabled)
                
                if not enabled:
                    raise HTTPException(
                        status_code=403,
                        detail=error_message
                    )
                
                return await func(*args, **kwargs)
                
            except HTTPException:
                raise
            except Exception as e:
                if default_enabled:
                    return await func(*args, **kwargs)
                else:
                    raise HTTPException(
                        status_code=500,
                        detail=f"Failed to evaluate feature flag: {e}"
                    )
        
        return wrapper
    return decorator


class FlexFlagMiddleware:
    """FastAPI middleware for FlexFlag integration"""
    
    def __init__(self, app, flexflag_config: FlexFlagConfig):
        self.app = app
        self.client = FlexFlagClient(flexflag_config)
        self.context_builder = FlexFlagContext()
    
    async def __call__(self, scope, receive, send):
        if scope["type"] == "http":
            # Add FlexFlag client to scope
            scope["flexflag"] = self.client
            
            # Initialize client if not ready
            if not self.client._ready.is_set():
                await self.client.wait_for_ready()
        
        await self.app(scope, receive, send)


# FastAPI lifespan events
@asynccontextmanager
async def flexflag_lifespan(app, config: FlexFlagConfig):
    """FastAPI lifespan context manager for FlexFlag"""
    
    # Startup
    client = FlexFlagClient(config)
    await client.wait_for_ready()
    
    # Store client in app state
    app.state.flexflag = client
    
    yield
    
    # Shutdown
    await client.close()


# Helper functions for FastAPI integration
async def get_feature_flags(
    flag_keys: list[str],
    client: FlexFlagClient = Depends(FlexFlagDependency),
    context: EvaluationContext = Depends(FlexFlagContext())
) -> Dict[str, FlagValue]:
    """Get multiple feature flags efficiently"""
    return await client.evaluate_batch(flag_keys, context)


async def get_user_experiments(
    experiment_flags: list[str],
    client: FlexFlagClient = Depends(FlexFlagDependency),
    context: EvaluationContext = Depends(FlexFlagContext())
) -> Dict[str, str]:
    """Get user's experiment variations"""
    results = {}
    for flag_key in experiment_flags:
        variation = await client.get_variation(flag_key, context)
        if variation:
            results[flag_key] = variation
    return results


# Response models for FastAPI
if FASTAPI_AVAILABLE:
    from pydantic import BaseModel
    from typing import List
    
    class FeatureFlagResponse(BaseModel):
        """Response model for feature flag evaluation"""
        flag_key: str
        value: FlagValue
        variation: Optional[str] = None
        cached: bool = False
        evaluation_time_ms: Optional[float] = None
    
    class BatchFeatureFlagsResponse(BaseModel):
        """Response model for batch feature flag evaluation"""
        flags: Dict[str, FlagValue]
        evaluation_time_ms: float
        cache_hit_rate: float
    
    class SDKMetricsResponse(BaseModel):
        """Response model for SDK metrics"""
        evaluations: int
        cache_hits: int
        cache_misses: int
        cache_hit_rate: float
        average_latency_ms: float
        network_requests: int
        errors: int


# API routes for FlexFlag management (optional)
if FASTAPI_AVAILABLE:
    from fastapi import APIRouter
    
    def create_flexflag_router(
        dependency: FlexFlagDependency,
        require_auth: bool = True
    ) -> APIRouter:
        """Create FastAPI router with FlexFlag management endpoints"""
        
        router = APIRouter(prefix="/flexflag", tags=["feature-flags"])
        
        auth_dependency = HTTPBearer() if require_auth else lambda: None
        
        @router.get("/metrics", response_model=SDKMetricsResponse)
        async def get_metrics(
            client: FlexFlagClient = Depends(dependency),
            credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)
        ):
            """Get SDK performance metrics"""
            metrics = client.get_metrics()
            cache_hit_rate = (metrics.cache_hits / metrics.evaluations * 100) if metrics.evaluations > 0 else 0
            
            return SDKMetricsResponse(
                evaluations=metrics.evaluations,
                cache_hits=metrics.cache_hits,
                cache_misses=metrics.cache_misses,
                cache_hit_rate=cache_hit_rate,
                average_latency_ms=metrics.average_latency_ms,
                network_requests=metrics.network_requests,
                errors=metrics.errors
            )
        
        @router.post("/cache/clear")
        async def clear_cache(
            client: FlexFlagClient = Depends(dependency),
            credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)
        ):
            """Clear feature flag cache"""
            await client.clear_cache()
            return {"message": "Cache cleared successfully"}
        
        @router.post("/evaluate", response_model=FeatureFlagResponse)
        async def evaluate_flag(
            flag_key: str,
            client: FlexFlagClient = Depends(dependency),
            context: EvaluationContext = Depends(FlexFlagContext()),
            credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)
        ):
            """Evaluate a single feature flag"""
            import time
            start_time = time.time()
            
            result = await client.evaluate_with_details(flag_key, context)
            evaluation_time = (time.time() - start_time) * 1000
            
            return FeatureFlagResponse(
                flag_key=flag_key,
                value=result.value,
                variation=result.variation,
                cached=result.reason.value == "CACHED",
                evaluation_time_ms=evaluation_time
            )
        
        @router.post("/evaluate/batch", response_model=BatchFeatureFlagsResponse)
        async def evaluate_flags_batch(
            flag_keys: List[str],
            client: FlexFlagClient = Depends(dependency),
            context: EvaluationContext = Depends(FlexFlagContext()),
            credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)
        ):
            """Evaluate multiple feature flags"""
            import time
            start_time = time.time()
            
            results = await client.evaluate_batch(flag_keys, context)
            evaluation_time = (time.time() - start_time) * 1000
            
            # Calculate cache hit rate for this batch
            metrics = client.get_metrics()
            cache_hit_rate = (metrics.cache_hits / metrics.evaluations * 100) if metrics.evaluations > 0 else 0
            
            return BatchFeatureFlagsResponse(
                flags=results,
                evaluation_time_ms=evaluation_time,
                cache_hit_rate=cache_hit_rate
            )
        
        return router