"""
FlexFlag Python SDK

High-performance feature flag management with intelligent caching and offline support.
"""

from .client import FlexFlagClient
from .cache import CacheProvider, MemoryCache, DiskCache, RedisCache, TieredCache, create_cache_provider
from .types import (
    FlexFlagConfig,
    EvaluationContext,
    FlagValue,
    CacheConfig,
    ConnectionConfig,
    OfflineConfig,
    LogLevel,
    EvaluationResult,
    EvaluationReason,
    EvaluationMetadata,
    SDKMetrics,
    EventHandlers,
    DefaultLogger,
)

__version__ = "1.0.0"
__author__ = "FlexFlag Team"
__email__ = "support@flexflag.io"

__all__ = [
    "FlexFlagClient",
    "CacheProvider",
    "MemoryCache", 
    "DiskCache",
    "RedisCache",
    "TieredCache",
    "create_cache_provider",
    "FlexFlagConfig",
    "EvaluationContext",
    "FlagValue",
    "CacheConfig",
    "ConnectionConfig",
    "OfflineConfig",
    "LogLevel",
    "EvaluationResult",
    "EvaluationReason",
    "EvaluationMetadata",
    "SDKMetrics",
    "EventHandlers",
    "DefaultLogger",
]