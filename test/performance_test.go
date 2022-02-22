
package test

import (
"net/http"
"strconv"
"strings"
"sync"
"testing"
"time"
)

const (
	// concurrency 是测试的并发度。
	concurrency = 1000
)

// testTask 是一个包装器，把 task 包装成 testTask.
func testTask(task func(no int)) string {

	beginTime := time.Now()
	wg := &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(no int) {
			defer wg.Done()
			task(no)
		}(i)
	}
	wg.Wait()
	return time.Now().Sub(beginTime).String()
}

// go test -v -count=1 performance_test.go -run=^TestHttpServer$
func TestHttpServer(t *testing.T) {

	writeTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		request, err := http.NewRequest("PUT", "http://localhost:5837/v1/cache/"+data, strings.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	})

	t.Logf("写入消耗时间为 %s。", writeTime)

	time.Sleep(3 * time.Second)

	readTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		request, err := http.NewRequest("GET", "http://localhost:5837/v1/cache/"+data, nil)
		if err != nil {
			t.Fatal(err)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	})

	t.Logf("读取消耗时间为 %s。", readTime)
}