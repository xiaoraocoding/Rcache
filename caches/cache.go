package caches

import (
	"sync"
	"sync/atomic"
	"time"
)

// Cache 是代表缓存的结构体。
type Cache struct {

	// segmentSize 是 segment 的数量，这个数量越多，理论上并发的性能就越好。
	segmentSize int

	// segments 存储着所有的 segment 实例。
	segments []*segment

	// options 是缓存配置。
	options *Options

	// dumping 标识当前缓存是否处于持久化状态。1 表示处于持久化状态。
	// 因为现在的 cache 是没有全局锁的，而持久化需要记录下当前的状态，不允许有更新，所以使用一个变量记录着，
	// 如果处于持久化状态，就让所有更新操作进入自旋状态，等待持久化完成再进行。
	dumping int32
}

// NewCache 返回一个默认配置的缓存实例。
func NewCache() *Cache {
	return NewCacheWith(DefaultOptions())
}

// NewCacheWith 返回一个使用 options 初始化过的缓存实例
func NewCacheWith(options Options) *Cache {
	// 尝试从持久化文件中恢复
	if cache, ok := recoverFromDumpFile(options.DumpFile); ok {
		return cache
	}
	return &Cache{
		segmentSize: options.SegmentSize,

		// 初始化所有的 segment
		segments: newSegments(&options),
		options:  &options,
		dumping:  0,
	}
}

// recoverFromDumpFile 从持久化文件中恢复缓存。
func recoverFromDumpFile(dumpFile string) (*Cache, bool) {
	cache, err := newEmptyDump().from(dumpFile)
	if err != nil {
		return nil, false
	}
	return cache, true
}

// newSegments 返回初始化好的 segment 实例列表。
func newSegments(options *Options) []*segment {
	// 根据配置的数量生成 segment
	segments := make([]*segment, options.SegmentSize)
	for i := 0; i < options.SegmentSize; i++ {
		segments[i] = newSegment(options)
	}
	return segments
}

// index 是选择 segment 的“特殊算法”。
// 这里参考了 Java 中的哈希生成逻辑，尽可能避免重复。不用去纠结为什么这么写，因为没有唯一的写法。
// 为了能使用到哈希值的全部数据，这里使用高位和低位进行异或操作。
func index(key string) int {
	index := 0
	keyBytes := []byte(key)
	for _, b := range keyBytes {
		index = 31*index + int(b&0xff)
	}
	return index ^ (index >> 16)
}

// segmentOf 返回 key 对应的 segment。
// 使用 index 生成的哈希值去获取 segment，这里使用 & 运算也是 Java 中的奇淫技巧。
func (c *Cache) segmentOf(key string) *segment {
	return c.segments[index(key)&(c.segmentSize-1)]
}

// Get 返回指定 key 的数据。
func (c *Cache) Get(key string) ([]byte, bool) {
	// 这边会等待持久化完成
	c.waitForDumping()
	return c.segmentOf(key).get(key)
}

// Set 添加指定的数据到缓存中。
func (c *Cache) Set(key string, value []byte) error {
	return c.SetWithTTL(key, value, NeverDie)
}

// SetWithTTL 添加指定的数据到缓存中，并设置相应的有效期。
func (c *Cache) SetWithTTL(key string, value []byte, ttl int64) error {
	// 这边会等待持久化完成
	c.waitForDumping()
	return c.segmentOf(key).set(key, value, ttl)
}

// Delete 从缓存中删除指定 key 的数据。
func (c *Cache) Delete(key string) error {
	// 这边会等待持久化完成
	c.waitForDumping()
	c.segmentOf(key).delete(key)
	return nil
}

// Status 返回缓存当前的情况。
func (c *Cache) Status() Status {
	result := NewStatus()
	for _, segment := range c.segments {
		status := segment.status()
		result.Count += status.Count
		result.KeySize += status.KeySize
		result.ValueSize += status.ValueSize
	}
	return *result
}

// gc 会清理缓存中过期的数据。
func (c *Cache) gc() {
	// 这边会等待持久化完成
	c.waitForDumping()
	wg := &sync.WaitGroup{}
	for _, seg := range c.segments {
		wg.Add(1)
		go func(s *segment) {
			defer wg.Done()
			s.gc()
		}(seg)
	}
	wg.Wait()
}

// AutoGc 会开启一个异步任务去定时清理过期的数据。
func (c *Cache) AutoGc() {
	go func() {
		ticker := time.NewTicker(time.Duration(c.options.GcDuration) * time.Minute)
		for {
			select {
			case <-ticker.C:
				c.gc()
			}
		}
	}()
}

// dump 会将缓存数据持久化到文件中。
func (c *Cache) dump() error {
	// 这边使用 atomic 包中的原子操作完成状态的切换
	atomic.StoreInt32(&c.dumping, 1)
	defer atomic.StoreInt32(&c.dumping, 0)
	return newDump(c).to(c.options.DumpFile)
}

// AutoDump 会开启一个异步任务去定时持久化缓存数据。
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

// waitForDumping 会等待持久化完成才返回。
func (c *Cache) waitForDumping() {
	for atomic.LoadInt32(&c.dumping) != 0 {
		// 每次循环都会等待一定的时间，如果不睡眠，会导致 CPU 空转消耗资源
		time.Sleep(time.Duration(c.options.CasSleepTime) * time.Microsecond)
	}
}
