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
	data []byte

	// ttl 代表这个数据的寿命。
	// 这个值的单位是秒。
	ttl int64

	// ctime 代表这个数据的创建时间。
	ctime int64
}

// newValue 返回一个包装之后的数据。
func newValue(data []byte, ttl int64) *value {
	return &value{
		data:  helpers.Copy(data),
		ttl:   ttl,
		ctime: time.Now().Unix(),
	}
}

// alive 返回这个数据是否存活。
func (v *value) alive() bool {
	// 首先判断是否有过期时间，然后判断当前时间是否超过了这个数据的死期
	return v.ttl == NeverDie || time.Now().Unix()-v.ctime < v.ttl
}

// visit 返回这个数据的实际存储数据。
func (v *value) visit() []byte {
	atomic.SwapInt64(&v.ctime, time.Now().Unix())
	return v.data
}