package vex

import "errors"

// Request:
// version    command    argsLength    {argLength    arg}
//  1byte      1byte       4byte          4byte    unknown

// Response:
// version    reply    bodyLength    {body}
//  1byte     1byte      4byte      unknown

const (
	ProtocolVersion        = byte(1) // 协议版本号
	headerLengthInProtocol = 6       // 协议中头部占用的字节数
	argsLengthInProtocol   = 4       // 协议中参数个数占用的字节数
	argLengthInProtocol    = 4       // 协议中参数长度占用的字节数
	bodyLengthInProtocol   = 4       // 协议体长度占用的字节数
)

var (
	// 协议版本不匹配错误，如果客户端和服务端的版本不一样就会返回这个错误
	ProtocolVersionMismatchErr = errors.New("protocol version between client and server doesn't match")
)
