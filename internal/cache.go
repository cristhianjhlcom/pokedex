package cache

import "time"

type Cache struct {
	cache map[string]CacheEntry
}

type CacheEntry struct {
	value     []byte
	createdAt time.Time
}

func NewCache(interval time.Duration) Cache {
	c := Cache{
		cache: make(map[string]CacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, value []byte) {
	c.cache[key] = CacheEntry{
		value:     value,
		createdAt: time.Now().UTC(),
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	entry, ok := c.cache[key]
	return entry.value, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.reap(interval)
	}
}

func (c *Cache) reap(interval time.Duration) {
	timeAgo := time.Now().UTC().Add(-interval)
	for k, v := range c.cache {
		if v.createdAt.Before(timeAgo) {
			delete(c.cache, k)
		}
	}
}
