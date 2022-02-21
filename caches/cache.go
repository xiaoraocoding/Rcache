package caches

import (
	"errors"
	"sync"
	"time"
)

// Cache 代表缓存的结构体。
type Cache struct {

	// data 存储着实际的键值对数据。
	data map[string]*value


	options Options

	// status 表示缓存的状态信息。
	// 这里使用的是指针类型，因为这个 status 会在内部不断地更新，所以想明确地表达出这个值是会被修改的。
	status *Status

	// lock 是保证并发安全的锁。
	lock *sync.RWMutex
}

// NewCache 返回一个默认配置的缓存对象。
func NewCache() *Cache {
	return NewCacheWith(DefaultOptions())
}

// NewCacheWith 返回一个指定配置的缓存对象。
func NewCacheWith(options Options) *Cache {
	// 这是新增代码，主要是先从持久化文件进行恢复，如果恢复不成功，就返回一个空的缓存
	if cache, ok := recoverFromDumpFile(options.DumpFile); ok {
		return cache
	}
	return &Cache{
		// 这里指定 256 的初始容量是为了减少哈希冲突的几率和扩容带来的性能损失
		data:    make(map[string]*value, 256),
		options: options,
		status:  newStatus(),
		lock:    &sync.RWMutex{},
	}
}


func recoverFromDumpFile(dumpFile string) (*Cache, bool) {
	cache, err := newEmptyDump().from(dumpFile)
	if err != nil {
		return nil, false
	}
	return cache, true
}

// Get 返回指定 Key 的数据，如果找不到就返回 false。
func (c *Cache) Get(key string) ([]byte, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if !value.alive() {
		c.lock.RUnlock()
		c.Delete(key)
		c.lock.RLock()
		return nil, false
	}

	// 注意这个 visit 方法会使用 Swap 的形式更新数据的创建时间，用于实现 LRU 过期机制
	return value.visit(), true
}

// Set 添加一个键值对到缓存中，不设定 ttl，也就意味着数据不会过期。
// 返回 error 是 nil 说明添加成功，否则就是添加失败，可能是触发了写满保护机制，拒绝写入数据。
func (c *Cache) Set(key string, value []byte) error {
	return c.SetWithTTL(key, value, NeverDie)
}

// SetWithTTL 添加一个键值对到缓存中，使用给定的 ttl 去设定过期时间。
// 返回 error 是 nil 说明添加成功，否则就是添加失败，可能是触发了写满保护机制，拒绝写入数据。
func (c *Cache) SetWithTTL(key string, value []byte, ttl int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if oldValue, ok := c.data[key]; ok {
		// 如果是已经存在的 key，就不属于新增键值对了，为了方便处理，先把原本的键值对信息去除
		c.status.subEntry(key, oldValue.Data)
	}

	// 这边会判断缓存的容量是否足够，如果不够了，就返回写满保护的错误信息
	if !c.checkEntrySize(key, value) {
		// 注意刚刚把旧的键值对信息去除了，现在要加回去，因为并没有添加新的键值对
		if oldValue, ok := c.data[key]; ok {
			c.status.addEntry(key, oldValue.Data)
		}


		return errors.New("the entry size will exceed if you set this entry")
	}

	// 添加新的键值对，需要先更新缓存信息，然后保存数据
	c.status.addEntry(key, value)
	c.data[key] = newValue(value, ttl)
	return nil
}

// Delete 删除 key 指定的数据。
func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if oldValue, ok := c.data[key]; ok {
		// 如果存在这个 key 才会进行删除，并且需要先把缓存信息更新掉
		c.status.subEntry(key, oldValue.Data)
		delete(c.data, key)
	}
}

// Status 返回缓存信息。
func (c *Cache) Status() Status {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return *c.status
}

// checkEntrySize 会检查要添加的键值对是否满足当前缓存的要求。
func (c *Cache) checkEntrySize(newKey string, newValue []byte) bool {
	// 将当前的键值对占用空间加上要被添加的键值对占用空间，然后和配置中的最大键值对占用空间进行比较
	return c.status.entrySize()+int64(len(newKey))+int64(len(newValue)) <= c.options.MaxEntrySize*1024*1024
}

// gc 会触发数据清理任务，主要是清理过期的数据。
func (c *Cache) gc() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 使用 count 记录当前清理的个数
	count := 0
	for key, value := range c.data {
		if !value.alive() {
			c.status.subEntry(key, value.Data)
			delete(c.data, key)
			count++
			if count >= c.options.MaxGcCount {
				break
			}
		}
	}
}

// AutoGc 会开启一个定时 GC 的异步任务。
func (c *Cache) AutoGc() {
	go func() {
		// 根据配置中的 GcDuration 来设置定时的间隔
		ticker := time.NewTicker(time.Duration(c.options.GcDuration) * time.Minute)
		for {
			select {
			case <-ticker.C:
				c.gc()
			}
		}
	}()
}

// dump 持久化缓存方法。
func (c *Cache) dump() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 创建出 dump 对象并持久化到文件
	return newDump(c).to(c.options.DumpFile)
}

// AutoDump 开启定时任务去持久化缓存。
// 和自动 Gc 的原理是一样的，这里就不再赘述了。
func (c *Cache) AutoDump() {
	go func() {
		ticker := time.NewTicker(time.Duration(c.options.DumpDuration) * time.Minute)
		for {
			select {
			case <-ticker.C:
				c.dump()
			}
		}
	}()
}