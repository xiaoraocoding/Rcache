package client

import "encoding/json"

// Status 是缓存状态结构体。
type Status struct {

	// Count 是键值对的数量。
	Count int `json:"count"`

	// KeySize 是 key 占用的大小
	KeySize int64 `json:"keySize"`

	// ValueSize 是 value 占用的大小。
	ValueSize int64 `json:"valueSize"`
}

// request 是请求结构体。
type request struct {

	// command 是执行的命令。
	command byte

	// args 是执行的参数。
	args [][]byte

	// resultChan 是用于接收结果的管道。
	resultChan chan *Response
}

// Response 是响应结构体。
type Response struct {

	// Body 是响应体。
	Body []byte

	// Err 是响应的错误。
	Err error
}

// ToStatus 会返回一个状态实例和错误。
func (r *Response) ToStatus() (*Status, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	status := &Status{}
	return status, json.Unmarshal(r.Body, status)
}
