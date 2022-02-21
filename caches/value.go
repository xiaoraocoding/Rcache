package caches

import (
	"Rcache/helpers"
	"sync/atomic"
	"time"
)

const (
	// NeverDie 是一个常量，我们设计的时候规定如果 ttl 为 0，那就是永不过期，相当于灵丹妙药。
	NeverDie = 0
)

// value 是一个包装了数据的结构体。
type value struct {

	// data 存储着真正的数据。
	Data []byte

	// ttl 代表这个数据的寿命。
	// 这个值的单位是秒。
	Ttl int64

	// ctime 代表这个数据的创建时间。
	Ctime int64
}

func newValue(data []byte, ttl int64) *value {
	return &value{
		// 注意修改字段为大写开头
		Data:  helpers.Copy(data),
		Ttl:   ttl,
		Ctime: time.Now().Unix(),
	}
}

func (v *value) alive() bool {
	return v.Ttl == NeverDie || time.Now().Unix()-v.Ctime < v.Ttl
}

func (v *value) visit() []byte {
	// 注意修改字段为大写开头
	atomic.SwapInt64(&v.Ctime, time.Now().Unix())
	return v.Data
}