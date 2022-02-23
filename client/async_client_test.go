package client

import (
	"strconv"
	"testing"
	"time"
)

const (
	// keySize is the key size of test.
	keySize = 10000
)

// testTask is a wrapper wraps task to testTask.
func testTask(task func(no int)) string {
	beginTime := time.Now()
	for i := 0; i < keySize; i++ {
		task(i)
	}
	return time.Now().Sub(beginTime).String()
}

// go test -v -count=1 -run=^TestAsyncClientPerformance$
func TestAsyncClientPerformance(t *testing.T) {

	client, err := NewAsyncClient(":5837")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	writeTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		client.Set(data, []byte(data), 0)
	})

	t.Logf("写入消耗时间为 %s！", writeTime)

	time.Sleep(3 * time.Second)

	readTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		client.Get(data)
	})

	t.Logf("读取消耗时间为 %s！", readTime)

	time.Sleep(time.Second)
}
