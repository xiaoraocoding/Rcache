package servers

import (
	"Rcache/caches"
	"Rcache/vex"
	"encoding/binary"
	"encoding/json"
	"errors"
)

const (
	// getCommand 是 get 命令。
	getCommand = byte(1)

	// setCommand 是 set 命令。
	setCommand = byte(2)

	// deleteCommand 是 delete 命令。
	deleteCommand = byte(3)

	// statusCommand 是 status 命令。
	statusCommand = byte(4)
)

var (
	// commandNeedsMoreArgumentsErr 是命令需要更多参数的错误。
	commandNeedsMoreArgumentsErr = errors.New("command needs more arguments")

	// notFoundErr 是找不到的错误。
	notFoundErr = errors.New("not found")
)

// TCPServer 是 TCP 类型的服务器。
type TCPServer struct {
	// cache 是内部用于存储数据的缓存组件。
	cache *caches.Cache

	// server 是内部真正用于服务的服务器。
	server *vex.Server
}

// NewTCPServer 返回新的 TCP 服务器。
func NewTCPServer(cache *caches.Cache) *TCPServer {
	return &TCPServer{
		cache:  cache,
		server: vex.NewServer(),
	}
}

// Run 运行这个 TCP 服务器。
func (ts *TCPServer) Run(address string) error {
	// 注册几种命令的处理器
	ts.server.RegisterHandler(getCommand, ts.getHandler)
	ts.server.RegisterHandler(setCommand, ts.setHandler)
	ts.server.RegisterHandler(deleteCommand, ts.deleteHandler)
	ts.server.RegisterHandler(statusCommand, ts.statusHandler)
	return ts.server.ListenAndServe("tcp", address)
}

// Close 用于关闭服务器。
func (ts *TCPServer) Close() error {
	return ts.server.Close()
}

// =======================================================================

// getHandler 是处理 get 命令的的处理器。
func (ts *TCPServer) getHandler(args [][]byte) (body []byte, err error) {

	// 检查参数个数是否足够
	if len(args) < 1 {
		return nil, commandNeedsMoreArgumentsErr
	}

	// 调用缓存的 Get 方法，如果不存在就返回 notFoundErr 错误
	value, ok := ts.cache.Get(string(args[0]))
	if !ok {
		return value, notFoundErr
	}
	return value, nil
}

// setHandler 是处理 set 命令的处理器。
func (ts *TCPServer) setHandler(args [][]byte) (body []byte, err error) {

	// 检查参数个数是否足够
	if len(args) < 3 {
		return nil, commandNeedsMoreArgumentsErr
	}

	// 读取 ttl，注意这里使用大端的方式读取，所以要求客户端也以大端的方式进行存储
	ttl := int64(binary.BigEndian.Uint64(args[0]))
	err = ts.cache.SetWithTTL(string(args[1]), args[2], ttl)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// deleteHandler 是处理 delete 命令的处理器。
func (ts *TCPServer) deleteHandler(args [][]byte) (body []byte, err error) {

	// 检查参数个数是否足够
	if len(args) < 1 {
		return nil, commandNeedsMoreArgumentsErr
	}

	// 删除指定的数据
	err = ts.cache.Delete(string(args[0]))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// statusHandler 是返回缓存状态的处理器。
func (ts *TCPServer) statusHandler(args [][]byte) (body []byte, err error) {
	return json.Marshal(ts.cache.Status())
}
