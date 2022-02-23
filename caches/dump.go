package caches

import (
	"encoding/gob"
	"os"
	"sync"
	"time"
)

type dump struct {
	// SegmentSize 是 segment 的数量。
	SegmentSize int

	// Segments 存储着所有的 segment 实例。
	Segments []*segment

	// Options 是缓存的选项配置。
	Options *Options
}

// newEmptyDump 返回一个空的持久化实例。
func newEmptyDump() *dump {
	return &dump{}
}

// newDump 返回一个从缓存实例初始化过来的持久化实例。
func newDump(c *Cache) *dump {
	return &dump{
		SegmentSize: c.segmentSize,
		Segments:    c.segments,
		Options:     c.options,
	}
}

// nowSuffix 返回当前时间，类似于 20060102150405。
func nowSuffix() string {
	return "." + time.Now().Format("20060102150405")
}

// to 会将 dump 实例持久化到文件中。
func (d *dump) to(dumpFile string) error {

	newDumpFile := dumpFile + nowSuffix()
	file, err := os.OpenFile(newDumpFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(d)
	if err != nil {
		file.Close()
		os.Remove(newDumpFile)
		return err
	}

	os.Remove(dumpFile)
	file.Close()
	return os.Rename(newDumpFile, dumpFile)
}

// from 返回一个从持久化文件中恢复的缓存实例。
func (d *dump) from(dumpFile string) (*Cache, error) {

	file, err := os.Open(dumpFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = gob.NewDecoder(file).Decode(d); err != nil {
		return nil, err
	}

	// 恢复出 segment 之后需要为每一个 segment 的未导出字段进行初始化
	for _, segment := range d.Segments {
		segment.options = d.Options
		segment.lock = &sync.RWMutex{}
	}

	return &Cache{
		segmentSize: d.SegmentSize,
		segments:    d.Segments,
		options:     d.Options,
		dumping:     0,
	}, nil
}
