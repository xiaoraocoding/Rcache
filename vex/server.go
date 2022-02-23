package vex

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"sync"
)

var (
	// 找不到对应的命令处理器错误
	commandHandlerNotFoundErr = errors.New("failed to find a handler of command")
)

// 服务端结构。
type Server struct {

	// 监听器，这个应该大家都很熟悉了吧。
	listener net.Listener

	// 命令处理器，通过命令可以找到对应的处理器。
	handlers map[byte]func(args [][]byte) (body []byte, err error)
}

// 创建新的服务端。
func NewServer() *Server {
	return &Server{
		handlers: map[byte]func(args [][]byte) (body []byte, err error){},
	}
}

// 注册命令处理器。
func (s *Server) RegisterHandler(command byte, handler func(args [][]byte) (body []byte, err error)) {
	s.handlers[command] = handler
}

// 监听并服务于 network 和 address。
func (s *Server) ListenAndServe(network string, address string) (err error) {

	// 监听指定地址
	s.listener, err = net.Listen(network, address)
	if err != nil {
		return err
	}

	// 使用 WaitGroup 记录连接数，并等待所有连接处理完毕
	wg := &sync.WaitGroup{}
	for {
		// 等待客户端连接
		conn, err := s.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			continue
		}

		// 记录连接
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(conn)
		}()
	}

	// 等待所有连接处理完毕
	wg.Wait()
	return nil
}

// 处理连接。
func (s *Server) handleConn(conn net.Conn) {

	// 将连接包装成缓冲读取器，提高读取的性能
	reader := bufio.NewReader(conn)
	defer conn.Close()

	for {
		// 读取并解析请求请求
		command, args, err := readRequestFrom(reader)
		if err != nil {
			if err == ProtocolVersionMismatchErr {
				continue
			}
			return
		}

		// 处理请求
		reply, body, err := s.handleRequest(command, args)
		if err != nil {
			writeErrorResponseTo(conn, err.Error())
			continue
		}

		// 发送处理结果的响应
		_, err = writeResponseTo(conn, reply, body)
		if err != nil {
			continue
		}
	}
}

// 处理请求。
func (s *Server) handleRequest(command byte, args [][]byte) (reply byte, body []byte, err error) {

	// 从命令处理器集合中选出对应的处理器
	handle, ok := s.handlers[command]
	if !ok {
		return ErrorReply, nil, commandHandlerNotFoundErr
	}

	// 将处理结果返回
	body, err = handle(args)
	if err != nil {
		return ErrorReply, body, err
	}
	return SuccessReply, body, err
}

// 关闭服务端的方法。
func (s *Server) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}
