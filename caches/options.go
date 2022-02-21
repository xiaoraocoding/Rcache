package caches


type Options struct {

	// MaxEntrySize 是写满保护的一个阈值，当缓存中的键值对占用空间达到这个值，就会触发写满保护。
	// 这个值的单位是 GB。
	MaxEntrySize int64

	// MaxGcCount 是自动淘汰机制的一个阈值，当清理的数据达到了这个值后就会停止清理了。
	MaxGcCount int

	// GcDuration 是自动淘汰机制的时间间隔，每隔固定的 GcDuration 时间就会进行一次自动淘汰。
	// 这个值的单位是分钟。
	GcDuration int64
}

// DefaultOptions 返回一个默认的选项设置对象。
func DefaultOptions() Options {
	return Options{
		MaxEntrySize: int64(4), // 默认是 4 GB
		MaxGcCount:   1000, // 默认是 1000 个
		GcDuration:   60, // 默认是 1 小时
	}
}

