package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// keySize 是测试的键值对数量。
	keySize = 10000
)

// testTask 是一个包装器，包装一个任务为测试任务。
func testTask(task func(no int)) string {
	beginTime := time.Now()
	for i := 0; i < keySize; i++ {
		task(i)
	}
	return time.Now().Sub(beginTime).String()
}

// go test -v -count=1 redis_test.go -run=^TestRedis$
func TestRedis(t *testing.T) {

	conn, err := redis.DialURL("redis://127.0.0.1:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	writeTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		conn.Do("set", data, data)
	})

	t.Logf("写入消耗时间为 %s。", writeTime)

	time.Sleep(3 * time.Second)

	readTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		conn.Do("get", data)
	})

	t.Logf("读取消耗时间为 %s。", readTime)
}
