
package test

import (
	"Rcache/servers"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	// keySize 是测试的键值对数量。
	keySize = 10000
)

// testTask 将任务包装成测试任务。
func testTask(task func(no int)) string {
	beginTime := time.Now()
	for i := 0; i < keySize; i++ {
		task(i)
	}
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

// go test -v -count=1 performance_test.go -run=^TestTcpServer$
func TestTcpServer(t *testing.T) {

	client, err := servers.NewTCPClient("127.0.0.1:5837")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	writeTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		err := client.Set(data, []byte(data), 0)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Logf("写入消耗时间为 %s。", writeTime)

	time.Sleep(3 * time.Second)

	readTime := testTask(func(no int) {
		data := strconv.Itoa(no)
		_, err := client.Get(data)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Logf("读取消耗时间为 %s。", readTime)
}