/**
 * Cache Provider Implementations
 * Multiple cache strategies for optimal performance
 */

import { LRUCache } from 'lru-cache';
import { CacheProvider, CacheConfig, FlagValue } from './types';

/**
 * In-memory cache using LRU eviction
 */
export class MemoryCache implements CacheProvider {
  private cache: LRUCache<string, { value: FlagValue; expires: number }>;
  private config: CacheConfig;

  constructor(config: CacheConfig) {
    this.config = config;
    this.cache = new LRUCache({
      max: config.maxSize || 1000,
      ttl: config.ttl || 300000, // 5 minutes default
      updateAgeOnGet: true,
      updateAgeOnHas: false,
    });
  }

  async get(key: string): Promise<FlagValue | null> {
    const item = this.cache.get(key);
    if (!item) return null;
    
    // Check expiration
    if (item.expires && Date.now() > item.expires) {
      this.cache.delete(key);
      return null;
    }
    
    return item.value;
  }

  async set(key: string, value: FlagValue, ttl?: number): Promise<void> {
    const expires = Date.now() + (ttl || this.config.ttl || 300000);
    this.cache.set(key, { value, expires });
  }

  async delete(key: string): Promise<void> {
    this.cache.delete(key);
  }

  async clear(): Promise<void> {
    this.cache.clear();
  }

  async has(key: string): Promise<boolean> {
    return this.cache.has(key);
  }

  async size(): Promise<number> {
    return this.cache.size;
  }

  async keys(): Promise<string[]> {
    return Array.from(this.cache.keys());
  }
}

/**
 * Browser storage cache (localStorage or sessionStorage)
 */
export class LocalStorageCache implements CacheProvider {
  private storage: Storage;
  private config: CacheConfig;
  private keyPrefix: string;
  private memoryIndex: Map<string, number> = new Map(); // Track keys and sizes

  constructor(config: CacheConfig & { storage?: 'localStorage' | 'sessionStorage' }) {
    this.config = config;
    this.keyPrefix = config.keyPrefix || 'flexflag:';
    
    if (typeof window === 'undefined') {
      throw new Error('LocalStorageCache can only be used in browser environment');
    }
    
    this.storage = config.storage === 'sessionStorage' 
      ? window.sessionStorage 
      : window.localStorage;
    
    // Load existing keys into memory index
    this.loadIndex();
  }

  private loadIndex(): void {
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i);
      if (key && key.startsWith(this.keyPrefix)) {
        const value = this.storage.getItem(key);
        if (value) {
          this.memoryIndex.set(key, value.length);
        }
      }
    }
  }

  private getFullKey(key: string): string {
    return `${this.keyPrefix}${key}`;
  }

  private enforceMaxSize(): void {
    if (!this.config.maxSize) return;
    
    while (this.memoryIndex.size > this.config.maxSize) {
      // Remove oldest entry (first in map)
      const oldestKey = this.memoryIndex.keys().next().value;
      if (oldestKey) {
        this.storage.removeItem(oldestKey);
        this.memoryIndex.delete(oldestKey);
      }
    }
  }

  async get(key: string): Promise<FlagValue | null> {
    const fullKey = this.getFullKey(key);
    const item = this.storage.getItem(fullKey);
    
    if (!item) return null;
    
    try {
      const parsed = JSON.parse(item);
      
      // Check expiration
      if (parsed.expires && Date.now() > parsed.expires) {
        await this.delete(key);
        return null;
      }
      
      // Decompress if needed
      if (this.config.compression && parsed.compressed) {
        return this.decompress(parsed.value);
      }
      
      return parsed.value;
    } catch (error) {
      console.error('Failed to parse cached value:', error);
      await this.delete(key);
      return null;
    }
  }

  async set(key: string, value: FlagValue, ttl?: number): Promise<void> {
    const fullKey = this.getFullKey(key);
    const expires = Date.now() + (ttl || this.config.ttl || 300000);
    
    let storedValue = value;
    let compressed = false;
    
    // Compress if enabled and value is large
    if (this.config.compression && JSON.stringify(value).length > 1024) {
      storedValue = this.compress(value);
      compressed = true;
    }
    
    const item = JSON.stringify({
      value: storedValue,
      expires,
      compressed,
      timestamp: Date.now()
    });
    
    try {
      this.storage.setItem(fullKey, item);
      this.memoryIndex.set(fullKey, item.length);
      this.enforceMaxSize();
    } catch (error) {
      // Handle quota exceeded error
      if (error instanceof DOMException && error.name === 'QuotaExceededError') {
        // Clear some old items and retry
        await this.clearOldest(5);
        try {
          this.storage.setItem(fullKey, item);
          this.memoryIndex.set(fullKey, item.length);
        } catch (retryError) {
          console.error('Failed to cache value after clearing space:', retryError);
        }
      } else {
        console.error('Failed to cache value:', error);
      }
    }
  }

  async delete(key: string): Promise<void> {
    const fullKey = this.getFullKey(key);
    this.storage.removeItem(fullKey);
    this.memoryIndex.delete(fullKey);
  }

  async clear(): Promise<void> {
    const keysToRemove: string[] = [];
    
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i);
      if (key && key.startsWith(this.keyPrefix)) {
        keysToRemove.push(key);
      }
    }
    
    keysToRemove.forEach(key => this.storage.removeItem(key));
    this.memoryIndex.clear();
  }

  async has(key: string): Promise<boolean> {
    const fullKey = this.getFullKey(key);
    return this.storage.getItem(fullKey) !== null;
  }

  async size(): Promise<number> {
    return this.memoryIndex.size;
  }

  async keys(): Promise<string[]> {
    const keys: string[] = [];
    
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i);
      if (key && key.startsWith(this.keyPrefix)) {
        keys.push(key.substring(this.keyPrefix.length));
      }
    }
    
    return keys;
  }

  private async clearOldest(count: number): Promise<void> {
    const entries = Array.from(this.memoryIndex.entries());
    const toRemove = entries.slice(0, Math.min(count, entries.length));
    
    for (const [key] of toRemove) {
      this.storage.removeItem(key);
      this.memoryIndex.delete(key);
    }
  }

  private compress(value: FlagValue): string {
    // Simple compression using base64 encoding
    // In production, you might use a library like pako for real compression
    const json = JSON.stringify(value);
    if (typeof btoa !== 'undefined') {
      return btoa(json);
    }
    return json;
  }

  private decompress(value: string): FlagValue {
    // Decompress base64
    if (typeof atob !== 'undefined') {
      try {
        const json = atob(value);
        return JSON.parse(json);
      } catch {
        return value;
      }
    }
    return value;
  }
}

/**
 * Tiered cache that uses multiple cache levels
 */
export class TieredCache implements CacheProvider {
  private caches: CacheProvider[];

  constructor(caches: CacheProvider[]) {
    this.caches = caches;
  }

  async get(key: string): Promise<FlagValue | null> {
    for (let i = 0; i < this.caches.length; i++) {
      const value = await this.caches[i].get(key);
      if (value !== null) {
        // Populate higher tier caches
        for (let j = 0; j < i; j++) {
          await this.caches[j].set(key, value);
        }
        return value;
      }
    }
    return null;
  }

  async set(key: string, value: FlagValue, ttl?: number): Promise<void> {
    // Set in all cache tiers
    await Promise.all(
      this.caches.map(cache => cache.set(key, value, ttl))
    );
  }

  async delete(key: string): Promise<void> {
    await Promise.all(
      this.caches.map(cache => cache.delete(key))
    );
  }

  async clear(): Promise<void> {
    await Promise.all(
      this.caches.map(cache => cache.clear())
    );
  }

  async has(key: string): Promise<boolean> {
    for (const cache of this.caches) {
      if (await cache.has(key)) {
        return true;
      }
    }
    return false;
  }

  async size(): Promise<number> {
    // Return size of first cache tier
    return this.caches[0] ? await this.caches[0].size() : 0;
  }

  async keys(): Promise<string[]> {
    // Return keys from first cache tier
    return this.caches[0] ? await this.caches[0].keys() : [];
  }
}

/**
 * No-op cache provider for when caching is disabled
 */
export class NoOpCache implements CacheProvider {
  async get(key: string): Promise<FlagValue | null> {
    return null;
  }

  async set(key: string, value: FlagValue, ttl?: number): Promise<void> {
    // No-op
  }

  async delete(key: string): Promise<void> {
    // No-op
  }

  async clear(): Promise<void> {
    // No-op
  }

  async has(key: string): Promise<boolean> {
    return false;
  }

  async size(): Promise<number> {
    return 0;
  }

  async keys(): Promise<string[]> {
    return [];
  }
}

// Export all cache providers
export { CacheProvider } from './types';