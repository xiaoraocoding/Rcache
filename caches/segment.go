package caches

import (
	"errors"
	"sync"
)

type segment struct {

	//  存储这个数据块的数据。
	Data map[string]*value

	//  记录着这个数据块的情况。
	Status *Status

	//  是缓存的选项设置。
	options *Options

	lock *sync.RWMutex
}

func newSegment(options *Options) *segment {
	return &segment{
		// 初始化 map 的时候给出初始大小，可以避免大量扩容带来的性能损耗
		Data:    make(map[string]*value, options.MapSizeOfSegment),
		Status:  NewStatus(),
		options: options,
		lock:    &sync.RWMutex{},
	}
}

// get 返回指定 key 的数据。
// 这个方法和原来 cache 的方法一样，只是移动到 segment 这里。
func (s *segment) get(key string) ([]byte, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok := s.Data[key]
	if !ok {
		return nil, false
	}

	if !value.alive() {
		s.lock.RUnlock()
		s.delete(key)
		s.lock.RLock()
		return nil, false
	}
	return value.visit(), true
}

// set 添加一个数据进 segment。
func (s *segment) set(key string, value []byte, ttl int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if oldValue, ok := s.Data[key]; ok {
		s.Status.subEntry(key, oldValue.Data)
	}

	if !s.checkEntrySize(key, value) {
		if oldValue, ok := s.Data[key]; ok {
			s.Status.addEntry(key, oldValue.Data)
		}
		return errors.New("the entry size will exceed if you set this entry")
	}

	s.Status.addEntry(key, value)
	s.Data[key] = newValue(value, ttl)
	return nil
}

// delete 从 segment 中删除指定 key 的数据。
func (s *segment) delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if oldValue, ok := s.Data[key]; ok {
		s.Status.subEntry(key, oldValue.Data)
		delete(s.Data, key)
	}
}

// Status 返回这个 segment 的情况。
func (s *segment) status() Status {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return *s.Status
}

// checkEntrySize 会判断数据容量是否已经达到了设定的上限。
// 因为这个配置是针对整个缓存的，而这边判断大小是针对单个 segment 的，所以需要算出单个 segment 的上限来判断。
func (s *segment) checkEntrySize(newKey string, newValue []byte) bool {
	return s.Status.entrySize()+int64(len(newKey))+int64(len(newValue)) <= int64((s.options.MaxEntrySize*1024*1024)/s.options.SegmentSize)
}

func (s *segment) gc() {
	s.lock.Lock()
	defer s.lock.Unlock()
	count := 0
	for key, value := range s.Data {
		if !value.alive() {
			s.Status.subEntry(key, value.Data)
			delete(s.Data, key)
			count++
			if count >= s.options.MaxGcCount {
				break
			}
		}
	}
}
