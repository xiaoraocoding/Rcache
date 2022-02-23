package vex

import (
	"bufio"
	"errors"
	"io"
	"net"
)

// 客户端结构。
type Client struct {

	// 和服务端建立的连接。
	conn net.Conn

	// 通往服务端的读取器。
	reader io.Reader
}

// 创建新的客户端。
func NewClient(network string, address string) (*Client, error) {

	// 和服务端建立连接
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Do(command byte, args [][]byte) (body []byte, err error) {

	// 包装请求然后发送给服务端
	_, err = writeRequestTo(c.conn, command, args)
	if err != nil {
		return nil, err
	}

	// 读取服务端返回的响应
	reply, body, err := readResponseFrom(c.reader)
	if err != nil {
		return body, err
	}

	// 如果是错误答复码，将内容包装成 error 并返回
	if reply == ErrorReply {
		return body, errors.New(string(body))
	}
	return body, nil
}

// 关闭客户端。
func (c *Client) Close() error {
	return c.conn.Close()
}
