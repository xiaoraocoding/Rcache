package client

import (
	"Rcache/vex"
	"encoding/binary"
)

const (
	// getCommand 是 get 的命令。
	getCommand = byte(1)

	// setCommand 是 set 的命令。
	setCommand = byte(2)

	// deleteCommand 是 delete 的命令。
	deleteCommand = byte(3)

	// statusCommand 是 status 的命令。
	statusCommand = byte(4)
)

// AsyncClient 是异步客户端。
type AsyncClient struct {

	// client 用于内部执行命令。
	client *vex.Client

	// requestChan 用于接收请求。
	requestChan chan *request
}

// NewAsyncClient 会创建一个异步客户端并返回。
func NewAsyncClient(address string) (*AsyncClient, error) {

	client, err := vex.NewClient("tcp", address)
	if err != nil {
		return nil, err
	}

	c := &AsyncClient{
		client:      client,
		requestChan: make(chan *request, 163840),
	}
	c.handleRequests()
	return c, nil
}

// handleRequests 会开启一个 goroutine 去处理请求。
func (ac *AsyncClient) handleRequests() {

	go func() {
		for request := range ac.requestChan {
			body, err := ac.client.Do(request.command, request.args)
			request.resultChan <- &Response{
				Body: body,
				Err:  err,
			}
		}
	}()
}

// do 使用异步的方式执行命令。
func (ac *AsyncClient) do(command byte, args [][]byte) <-chan *Response {

	// 设置一个缓冲位置放响应
	resultChan := make(chan *Response, 1)
	ac.requestChan <- &request{
		command:    command,
		args:       args,
		resultChan: resultChan,
	}
	return resultChan
}

// Get 用于执行 get 命令。
func (ac *AsyncClient) Get(key string) <-chan *Response {
	return ac.do(getCommand, [][]byte{[]byte(key)})
}

// Set 用于执行 set 命令。
func (ac *AsyncClient) Set(key string, value []byte, ttl int64) <-chan *Response {
	ttlBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(ttlBytes, uint64(ttl))
	return ac.do(setCommand, [][]byte{
		ttlBytes, []byte(key), value,
	})
}

// Delete 用于执行 delete 命令。
func (ac *AsyncClient) Delete(key string) <-chan *Response {
	return ac.do(deleteCommand, [][]byte{[]byte(key)})
}

// Status 用于执行 status 命令。
func (ac *AsyncClient) Status() <-chan *Response {
	return ac.do(statusCommand, nil)
}

// Close 关闭客户端并释放资源。
func (ac *AsyncClient) Close() error {
	close(ac.requestChan)
	return ac.client.Close()
}
