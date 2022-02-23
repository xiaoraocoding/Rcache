package caches

// Options 是选项配置结构体。
type Options struct {

	// MaxEntrySize 指键值对最大容量。
	MaxEntrySize int

	// MaxGcCount 指每个 segment 要清理的过期数据个数。
	MaxGcCount int

	// GcDuration 指多久执行一次 Gc 工作。
	// 单位是分钟。
	GcDuration int

	// DumpFile 指持久化文件的路径。
	DumpFile string

	// DumpDuration 指多久执行一次持久化。
	// 单位是分钟。
	DumpDuration int

	// MapSizeOfSegment 指 segment 中 map 的初始化大小。
	MapSizeOfSegment int

	// SegmentSize 指缓存中有多少个 segment。
	SegmentSize int

	// CasSleepTime 指每一次 CAS 自旋需要等待的时间。
	// 单位是微秒。
	CasSleepTime int
}

// DefaultOptions 返回默认的选项配置。
func DefaultOptions() Options {
	return Options{
		MaxEntrySize:     4, // 4 GB
		MaxGcCount:       10,
		GcDuration:       60, // 1 hour
		DumpFile:         "kafo.dump",
		DumpDuration:     30, // 30 minutes
		MapSizeOfSegment: 256,
		SegmentSize:      1024,
		CasSleepTime:     1000, // 1 ms
	}
}
