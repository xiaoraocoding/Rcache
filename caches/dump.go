package caches

import (
	"encoding/gob"
	"os"
	"sync"
	"time"
)

type dump struct {
	Data map[string]*value
	Options Options
	Status *Status
}


func newEmptyDump() *dump {
	return &dump{}
}

func newDump(c *Cache) *dump {
	return &dump{
		Data:    c.data,
		Options: c.options,
		Status:  c.status,
	}
}

// nowSuffix 返回一个类似于 20060102150405 的文件后缀名.
func nowSuffix() string {
	return "." + time.Now().Format("20060102150405")
}

// to 会将 dump 持久化到 dumpFile 中
func (d *dump) to(dumpFile string) error {

	newDumpFile := dumpFile + nowSuffix()
	file, err := os.OpenFile(newDumpFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(d)
	if err != nil {
		// 注意这里需要先把文件关闭了，不然 os.Remove 是没有权限删除这个文件的
		file.Close()
		os.Remove(newDumpFile)
		return err
	}

	// 将旧的持久化文件删除
	os.Remove(dumpFile)


	file.Close()
	return os.Rename(newDumpFile, dumpFile)
}

// from 会从 dumpFile 中恢复数据到一个 Cache 结构对象并返回。
func (d *dump) from(dumpFile string) (*Cache, error) {
	// 读取 dumpFile 文件并使用反序列化器进行反序列化
	file, err := os.Open(dumpFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = gob.NewDecoder(file).Decode(d); err != nil {
		return nil, err
	}

	// 然后初始化一个缓存对象并返回
	return &Cache{
		data:    d.Data,
		options: d.Options,
		status:  d.Status,
		lock:    &sync.RWMutex{},
	}, nil
}




