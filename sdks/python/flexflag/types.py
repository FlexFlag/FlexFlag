"""
FlexFlag Python SDK Type Definitions
"""

from typing import Any, Dict, List, Optional, Union, Callable, Awaitable
from typing_extensions import TypedDict, Literal
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
import logging

# Type aliases
FlagValue = Union[bool, str, int, float, Dict[str, Any], List[Any], None]
LogLevel = Literal["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
ConnectionMode = Literal["streaming", "polling", "offline"]
CacheStorage = Literal["memory", "disk", "redis", "custom"]

class EvaluationReason(Enum):
    """Reasons for flag evaluation results"""
    TARGETING_MATCH = "TARGETING_MATCH"
    DEFAULT = "DEFAULT"
    DISABLED = "DISABLED"
    CACHED = "CACHED"
    OFFLINE = "OFFLINE"
    ERROR = "ERROR"

@dataclass
class EvaluationContext:
    """Context for flag evaluation"""
    user_id: Optional[str] = None
    attributes: Dict[str, Any] = field(default_factory=dict)
    device: Optional[Dict[str, str]] = None
    session: Optional[Dict[str, Any]] = None

@dataclass
class CacheConfig:
    """Cache configuration"""
    enabled: bool = True
    ttl: int = 300  # 5 minutes in seconds
    max_size: int = 1000
    storage: CacheStorage = "memory"
    key_prefix: str = "flexflag:"
    compression: bool = False
    storage_file: Optional[str] = None  # For disk cache
    
@dataclass
class ConnectionConfig:
    """Connection configuration"""
    mode: ConnectionMode = "streaming"
    polling_interval: int = 30  # seconds
    timeout: int = 5  # seconds
    retry_attempts: int = 3
    retry_delay: int = 1  # seconds
    exponential_backoff: bool = True
    headers: Dict[str, str] = field(default_factory=dict)

@dataclass
class OfflineConfig:
    """Offline mode configuration"""
    enabled: bool = True
    default_flags: Dict[str, FlagValue] = field(default_factory=dict)
    persistence: bool = True
    storage_file: Optional[str] = None

@dataclass
class EventHandlers:
    """Event handler configuration"""
    on_ready: Optional[Callable[[], None]] = None
    on_evaluation: Optional[Callable[[str, FlagValue], None]] = None
    on_update: Optional[Callable[[List[str]], None]] = None
    on_error: Optional[Callable[[Exception], None]] = None
    on_cache_hit: Optional[Callable[[str], None]] = None
    on_cache_miss: Optional[Callable[[str], None]] = None

@dataclass
class FlexFlagConfig:
    """Main FlexFlag configuration"""
    api_key: str
    base_url: str = "https://api.flexflag.io"
    environment: str = "production"
    cache: CacheConfig = field(default_factory=CacheConfig)
    connection: ConnectionConfig = field(default_factory=ConnectionConfig)
    offline: OfflineConfig = field(default_factory=OfflineConfig)
    log_level: LogLevel = "WARNING"
    events: EventHandlers = field(default_factory=EventHandlers)

@dataclass
class EvaluationMetadata:
    """Metadata about flag evaluation"""
    timestamp: datetime
    cache_hit: bool
    evaluation_time_ms: float
    source: str
    request_id: Optional[str] = None

@dataclass
class EvaluationResult:
    """Detailed flag evaluation result"""
    value: FlagValue
    variation: Optional[str] = None
    reason: EvaluationReason = EvaluationReason.DEFAULT
    metadata: Optional[EvaluationMetadata] = None

@dataclass
class Flag:
    """Feature flag definition"""
    key: str
    value: FlagValue
    type: str
    enabled: bool
    variations: List[Dict[str, Any]] = field(default_factory=list)
    targeting: List[Dict[str, Any]] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)

@dataclass
class SDKMetrics:
    """SDK performance metrics"""
    evaluations: int = 0
    cache_hits: int = 0
    cache_misses: int = 0
    errors: int = 0
    network_requests: int = 0
    average_latency_ms: float = 0.0

class BatchEvaluationRequest(TypedDict):
    """Batch evaluation request"""
    flags: List[str]
    context: EvaluationContext

class BatchEvaluationResponse(TypedDict):
    """Batch evaluation response"""
    flags: Dict[str, FlagValue]
    errors: Dict[str, str]

# Abstract base classes for extensibility
class CacheProvider:
    """Abstract cache provider interface"""
    
    async def get(self, key: str) -> Optional[FlagValue]:
        """Get value from cache"""
        raise NotImplementedError
    
    async def set(self, key: str, value: FlagValue, ttl: Optional[int] = None) -> None:
        """Set value in cache with optional TTL"""
        raise NotImplementedError
    
    async def delete(self, key: str) -> None:
        """Delete value from cache"""
        raise NotImplementedError
    
    async def clear(self) -> None:
        """Clear all cached values"""
        raise NotImplementedError
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in cache"""
        raise NotImplementedError
    
    async def size(self) -> int:
        """Get cache size"""
        raise NotImplementedError
    
    async def keys(self) -> List[str]:
        """Get all cache keys"""
        raise NotImplementedError

class Logger:
    """Abstract logger interface"""
    
    def debug(self, message: str, *args: Any, **kwargs: Any) -> None:
        raise NotImplementedError
    
    def info(self, message: str, *args: Any, **kwargs: Any) -> None:
        raise NotImplementedError
    
    def warning(self, message: str, *args: Any, **kwargs: Any) -> None:
        raise NotImplementedError
    
    def error(self, message: str, *args: Any, **kwargs: Any) -> None:
        raise NotImplementedError
    
    def critical(self, message: str, *args: Any, **kwargs: Any) -> None:
        raise NotImplementedError

class DefaultLogger(Logger):
    """Default logger implementation using Python's logging module"""
    
    def __init__(self, level: LogLevel = "WARNING"):
        self.logger = logging.getLogger("flexflag")
        self.logger.setLevel(getattr(logging, level))
        
        if not self.logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '[%(asctime)s] [FlexFlag] [%(levelname)s] %(message)s'
            )
            handler.setFormatter(formatter)
            self.logger.addHandler(handler)
    
    def debug(self, message: str, *args: Any, **kwargs: Any) -> None:
        self.logger.debug(message, *args, **kwargs)
    
    def info(self, message: str, *args: Any, **kwargs: Any) -> None:
        self.logger.info(message, *args, **kwargs)
    
    def warning(self, message: str, *args: Any, **kwargs: Any) -> None:
        self.logger.warning(message, *args, **kwargs)
    
    def error(self, message: str, *args: Any, **kwargs: Any) -> None:
        self.logger.error(message, *args, **kwargs)
    
    def critical(self, message: str, *args: Any, **kwargs: Any) -> None:
        self.logger.critical(message, *args, **kwargs)