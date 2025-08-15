"""
FlexFlag Python SDK Cache Providers

High-performance cache implementations with TTL support and various storage backends.
"""

import asyncio
import json
import pickle
import sqlite3
import time
from typing import Any, Dict, List, Optional
from cachetools import TTLCache
from pathlib import Path

from .types import FlagValue, CacheConfig, CacheProvider


class MemoryCache(CacheProvider):
    """In-memory cache using TTL cache for maximum performance"""
    
    def __init__(self, config: CacheConfig):
        self.config = config
        self.cache = TTLCache(maxsize=config.max_size, ttl=config.ttl)
        self._lock = asyncio.Lock()
    
    async def get(self, key: str) -> Optional[FlagValue]:
        """Get value from memory cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            return self.cache.get(prefixed_key)
    
    async def set(self, key: str, value: FlagValue, ttl: Optional[int] = None) -> None:
        """Set value in memory cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            # TTLCache doesn't support per-item TTL, so we use the configured TTL
            self.cache[prefixed_key] = value
    
    async def delete(self, key: str) -> None:
        """Delete value from memory cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            self.cache.pop(prefixed_key, None)
    
    async def clear(self) -> None:
        """Clear all cached values"""
        async with self._lock:
            self.cache.clear()
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            return prefixed_key in self.cache
    
    async def size(self) -> int:
        """Get cache size"""
        async with self._lock:
            return len(self.cache)
    
    async def keys(self) -> List[str]:
        """Get all cache keys"""
        async with self._lock:
            prefix = self.config.key_prefix
            return [
                key[len(prefix):] for key in self.cache.keys() 
                if key.startswith(prefix)
            ]


class DiskCache(CacheProvider):
    """Disk-based cache using SQLite for persistence"""
    
    def __init__(self, config: CacheConfig):
        self.config = config
        self.db_path = config.storage_file or "/tmp/flexflag_cache.db"
        self._lock = asyncio.Lock()
        self._init_db()
    
    def _init_db(self):
        """Initialize SQLite database"""
        Path(self.db_path).parent.mkdir(parents=True, exist_ok=True)
        
        conn = sqlite3.connect(self.db_path)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS cache (
                key TEXT PRIMARY KEY,
                value BLOB,
                expires_at REAL
            )
        """)
        conn.execute("CREATE INDEX IF NOT EXISTS idx_expires_at ON cache(expires_at)")
        conn.commit()
        conn.close()
    
    def _serialize_value(self, value: FlagValue) -> bytes:
        """Serialize value for storage"""
        if self.config.compression:
            import gzip
            return gzip.compress(pickle.dumps(value))
        return pickle.dumps(value)
    
    def _deserialize_value(self, data: bytes) -> FlagValue:
        """Deserialize value from storage"""
        if self.config.compression:
            import gzip
            return pickle.loads(gzip.decompress(data))
        return pickle.loads(data)
    
    async def get(self, key: str) -> Optional[FlagValue]:
        """Get value from disk cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            
            conn = sqlite3.connect(self.db_path)
            cursor = conn.execute(
                "SELECT value, expires_at FROM cache WHERE key = ?",
                (prefixed_key,)
            )
            row = cursor.fetchone()
            conn.close()
            
            if row is None:
                return None
            
            value_data, expires_at = row
            
            # Check if expired
            if expires_at < time.time():
                await self.delete(key)
                return None
            
            return self._deserialize_value(value_data)
    
    async def set(self, key: str, value: FlagValue, ttl: Optional[int] = None) -> None:
        """Set value in disk cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            expires_at = time.time() + (ttl or self.config.ttl)
            value_data = self._serialize_value(value)
            
            conn = sqlite3.connect(self.db_path)
            conn.execute(
                "INSERT OR REPLACE INTO cache (key, value, expires_at) VALUES (?, ?, ?)",
                (prefixed_key, value_data, expires_at)
            )
            conn.commit()
            conn.close()
            
            # Clean up expired entries periodically
            if hash(key) % 100 == 0:  # 1% chance to trigger cleanup
                await self._cleanup_expired()
    
    async def delete(self, key: str) -> None:
        """Delete value from disk cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            
            conn = sqlite3.connect(self.db_path)
            conn.execute("DELETE FROM cache WHERE key = ?", (prefixed_key,))
            conn.commit()
            conn.close()
    
    async def clear(self) -> None:
        """Clear all cached values"""
        async with self._lock:
            conn = sqlite3.connect(self.db_path)
            conn.execute("DELETE FROM cache")
            conn.commit()
            conn.close()
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in cache"""
        async with self._lock:
            prefixed_key = f"{self.config.key_prefix}{key}"
            
            conn = sqlite3.connect(self.db_path)
            cursor = conn.execute(
                "SELECT expires_at FROM cache WHERE key = ?",
                (prefixed_key,)
            )
            row = cursor.fetchone()
            conn.close()
            
            if row is None:
                return False
            
            # Check if expired
            expires_at = row[0]
            if expires_at < time.time():
                await self.delete(key)
                return False
            
            return True
    
    async def size(self) -> int:
        """Get cache size"""
        async with self._lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.execute("SELECT COUNT(*) FROM cache WHERE expires_at > ?", (time.time(),))
            count = cursor.fetchone()[0]
            conn.close()
            return count
    
    async def keys(self) -> List[str]:
        """Get all cache keys"""
        async with self._lock:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.execute(
                "SELECT key FROM cache WHERE expires_at > ?", 
                (time.time(),)
            )
            rows = cursor.fetchall()
            conn.close()
            
            prefix = self.config.key_prefix
            return [
                key[len(prefix):] for (key,) in rows 
                if key.startswith(prefix)
            ]
    
    async def _cleanup_expired(self):
        """Remove expired entries"""
        conn = sqlite3.connect(self.db_path)
        conn.execute("DELETE FROM cache WHERE expires_at <= ?", (time.time(),))
        conn.commit()
        conn.close()


class RedisCache(CacheProvider):
    """Redis cache for distributed caching"""
    
    def __init__(self, config: CacheConfig, redis_url: str = "redis://localhost:6379/0"):
        self.config = config
        self.redis_url = redis_url
        self._redis = None
        self._lock = asyncio.Lock()
    
    async def _get_redis(self):
        """Get Redis connection"""
        if self._redis is None:
            try:
                import aioredis
                self._redis = aioredis.from_url(self.redis_url)
            except ImportError:
                raise ImportError("aioredis is required for Redis cache. Install with: pip install aioredis")
        return self._redis
    
    def _serialize_value(self, value: FlagValue) -> str:
        """Serialize value for Redis storage"""
        if self.config.compression:
            import gzip
            import base64
            compressed = gzip.compress(json.dumps(value).encode())
            return base64.b64encode(compressed).decode()
        return json.dumps(value)
    
    def _deserialize_value(self, data: str) -> FlagValue:
        """Deserialize value from Redis storage"""
        if self.config.compression:
            import gzip
            import base64
            compressed = base64.b64decode(data.encode())
            return json.loads(gzip.decompress(compressed).decode())
        return json.loads(data)
    
    async def get(self, key: str) -> Optional[FlagValue]:
        """Get value from Redis cache"""
        async with self._lock:
            redis = await self._get_redis()
            prefixed_key = f"{self.config.key_prefix}{key}"
            
            data = await redis.get(prefixed_key)
            if data is None:
                return None
            
            return self._deserialize_value(data)
    
    async def set(self, key: str, value: FlagValue, ttl: Optional[int] = None) -> None:
        """Set value in Redis cache"""
        async with self._lock:
            redis = await self._get_redis()
            prefixed_key = f"{self.config.key_prefix}{key}"
            data = self._serialize_value(value)
            
            await redis.setex(prefixed_key, ttl or self.config.ttl, data)
    
    async def delete(self, key: str) -> None:
        """Delete value from Redis cache"""
        async with self._lock:
            redis = await self._get_redis()
            prefixed_key = f"{self.config.key_prefix}{key}"
            await redis.delete(prefixed_key)
    
    async def clear(self) -> None:
        """Clear all cached values with prefix"""
        async with self._lock:
            redis = await self._get_redis()
            pattern = f"{self.config.key_prefix}*"
            
            keys = await redis.keys(pattern)
            if keys:
                await redis.delete(*keys)
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in Redis cache"""
        async with self._lock:
            redis = await self._get_redis()
            prefixed_key = f"{self.config.key_prefix}{key}"
            return bool(await redis.exists(prefixed_key))
    
    async def size(self) -> int:
        """Get cache size"""
        async with self._lock:
            redis = await self._get_redis()
            pattern = f"{self.config.key_prefix}*"
            keys = await redis.keys(pattern)
            return len(keys)
    
    async def keys(self) -> List[str]:
        """Get all cache keys"""
        async with self._lock:
            redis = await self._get_redis()
            pattern = f"{self.config.key_prefix}*"
            keys = await redis.keys(pattern)
            
            prefix = self.config.key_prefix
            return [
                key.decode()[len(prefix):] for key in keys
                if key.decode().startswith(prefix)
            ]
    
    async def close(self):
        """Close Redis connection"""
        if self._redis:
            await self._redis.close()


class TieredCache(CacheProvider):
    """Multi-tier cache with fallback strategy"""
    
    def __init__(self, caches: List[CacheProvider]):
        if not caches:
            raise ValueError("At least one cache provider is required")
        self.caches = caches
        self._lock = asyncio.Lock()
    
    async def get(self, key: str) -> Optional[FlagValue]:
        """Get value from tiered cache (check each tier)"""
        async with self._lock:
            for i, cache in enumerate(self.caches):
                value = await cache.get(key)
                if value is not None:
                    # Backfill higher tiers
                    for j in range(i):
                        await self.caches[j].set(key, value)
                    return value
            return None
    
    async def set(self, key: str, value: FlagValue, ttl: Optional[int] = None) -> None:
        """Set value in all cache tiers"""
        async with self._lock:
            tasks = [cache.set(key, value, ttl) for cache in self.caches]
            await asyncio.gather(*tasks, return_exceptions=True)
    
    async def delete(self, key: str) -> None:
        """Delete value from all cache tiers"""
        async with self._lock:
            tasks = [cache.delete(key) for cache in self.caches]
            await asyncio.gather(*tasks, return_exceptions=True)
    
    async def clear(self) -> None:
        """Clear all cache tiers"""
        async with self._lock:
            tasks = [cache.clear() for cache in self.caches]
            await asyncio.gather(*tasks, return_exceptions=True)
    
    async def exists(self, key: str) -> bool:
        """Check if key exists in any cache tier"""
        async with self._lock:
            for cache in self.caches:
                if await cache.exists(key):
                    return True
            return False
    
    async def size(self) -> int:
        """Get size of first cache tier"""
        async with self._lock:
            return await self.caches[0].size()
    
    async def keys(self) -> List[str]:
        """Get keys from first cache tier"""
        async with self._lock:
            return await self.caches[0].keys()


def create_cache_provider(config: CacheConfig) -> CacheProvider:
    """Factory function to create cache provider based on config"""
    if config.storage == "memory":
        return MemoryCache(config)
    elif config.storage == "disk":
        return DiskCache(config)
    elif config.storage == "redis":
        return RedisCache(config)
    else:
        raise ValueError(f"Unsupported cache storage: {config.storage}")