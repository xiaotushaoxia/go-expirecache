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
	mp map[string]*cacheItem
	m  sync.RWMutex

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
	c.m.Lock()
	c.mp[key] = &cacheItem{
		data:        v,
		expiredTime: time.Now().Add(d).UnixNano(),
	}
	c.m.Unlock()
	c.clear()
}

func (c *Cache) Delete(key string) {
	c.m.Lock()
	delete(c.mp, key)
	c.m.Unlock()
	c.clear()
}

func (c *Cache) Items() map[string]any {
	mp := make(map[string]any)
	now, ok := c.clearIntervalOK()

	if !ok { // 上次清理还比较近 不清理
		c.m.RLock()
		for s, t := range c.mp {
			if now < t.expiredTime {
				mp[s] = t.data
			}
		}
		c.m.RUnlock()
	} else { // 上次清理比较远 清理
		var needDelete []string
		c.m.RLock()
		for s, t := range c.mp {
			if now < t.expiredTime {
				mp[s] = t.data
			} else {
				needDelete = append(needDelete, s)
			}
		}
		c.m.RUnlock()
		c.deleteKeys(needDelete)
	}

	return mp
}

func (c *Cache) deleteKeys(keys []string) {
	if len(keys) == 0 {
		return
	}
	go func() {
		c.m.Lock()
		for _, key := range keys {
			delete(c.mp, key)
		}
		c.m.Unlock()
	}()
}

func (c *Cache) clear() {
	now, ok := c.clearIntervalOK()
	if !ok {
		return
	}

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

func (c *Cache) clearIntervalOK() (int64, bool) {
	now := current()
	if now-atomic.LoadInt64(c.lastClearTime) < c.clearInterval {
		return now, false
	}
	atomic.StoreInt64(c.lastClearTime, now)
	return now, true
}

type cacheItem struct {
	expiredTime int64
	data        any
}

func (item *cacheItem) Expired() bool {
	return current() >= item.expiredTime
}

func current() int64 {
	return time.Now().UnixNano()
}
