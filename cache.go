package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// NewCache interval是清理的最小间隔时间
func NewCache(interval time.Duration) *Cache {
	a := Cache{
		mp:            make(map[string]*cacheItem),
		clearInterval: int64(interval),
		lastClearTime: new(int64),
	}
	*a.lastClearTime = current()
	return &a
}

type Cache struct {
	mp            map[string]*cacheItem
	m             sync.RWMutex

	lastClearTime *int64
	clearInterval int64
}

func (c *Cache) Get(key string) (any, bool) {
	defer c.clear()
	c.m.RLock()
	cc, ok := c.mp[key]
	c.m.RUnlock()
	if !ok {
		return nil, false
	}
	if cc.Expired() {
		// 顺手删掉
		c.m.Lock()
		delete(c.mp, key)
		c.m.Unlock()
		return nil, false
	}
	return cc.data, true
}

func (c *Cache) Set(key string, v any, d time.Duration) {
	defer c.clear()
	c.m.Lock()
	c.mp[key] = &cacheItem{
		data:        v,
		expiredTime: time.Now().Add(d).UnixNano(),
	}
	c.m.Unlock()
}

func (c *Cache) Delete(key string) {
	defer c.clear()
	c.m.Lock()
	delete(c.mp, key)
	c.m.Unlock()
}

func (c *Cache) Items() map[string]any {
	mp := make(map[string]any)
	var needDelete []string
	now := current()
	for s, t := range c.mp {
		if now > t.expiredTime {
			needDelete = append(needDelete, s)
		} else {
			mp[s] = t.data
		}
	}
	defer c.deleteKeys(needDelete)
	return mp
}

func (c *Cache) deleteKeys(keys []string)  {
	go func() {
		c.m.Lock()
		for _, key := range keys {
			delete(c.mp, key)
		}
		c.m.Unlock()
	}()
}

func (c *Cache) clear() {
	now := current()
	if now-atomic.LoadInt64(c.lastClearTime) < c.clearInterval {
		return
	}
	atomic.StoreInt64(c.lastClearTime, now)
	go func() {
		c.m.Lock()
		defer c.m.Unlock()
		var needDelete []string
		for s, t := range c.mp {
			if now > t.expiredTime {
				needDelete = append(needDelete, s)
			}
		}
		for _, s := range needDelete {
			delete(c.mp, s)
		}
	}()
}

type cacheItem struct {
	expiredTime int64
	data        any
}

func (item *cacheItem) Expired() bool {
	return current() > item.expiredTime
}

func current() int64 {
	return time.Now().UnixNano()
}