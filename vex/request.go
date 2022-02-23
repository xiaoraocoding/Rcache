package vex

import (
	"encoding/binary"
	"io"
)

//读取请求，解析出命令
func readRequestFrom(reader io.Reader) (command byte, args [][]byte, err error) {
	header := make([]byte, headerLengthInProtocol)
	_, err = io.ReadFull(reader, header)
	if err != nil {
		return 0, nil, err
	}
	version := header[0]
	if version != ProtocolVersion {
		return 0, nil, ProtocolVersionMismatchErr
	}
	command = header[1]
	header = header[2:]
	//将头部的信息转化为一个数字
	argsLength := binary.BigEndian.Uint32(header) //此时的argsLength就是本次请求的个数

	args = make([][]byte, argsLength)
	if argsLength > 0 {
		argLength := make([]byte, argLengthInProtocol)
		for i := uint32(0); i < argsLength; i++ {
			_, err = io.ReadFull(reader, argLength)
			if err != nil {
				return 0, nil, err
			}

			arg := make([]byte, binary.BigEndian.Uint32(argLength))
			_, err = io.ReadFull(reader, arg)
			if err != nil {
				return 0, nil, err
			}
			args[i] = arg
		}
	}
	return command, args, nil
}

//将请求的具体内容写入到writer
func writeRequestTo(writer io.Writer, command byte, args [][]byte) (int, error) {
	// 创建一个缓存区，并将协议版本号、命令和参数个数等写入缓存区
	request := make([]byte, headerLengthInProtocol)
	request[0] = ProtocolVersion
	request[1] = command
	binary.BigEndian.PutUint32(request[2:], uint32(len(args)))

	if len(args) > 0 {
		// 将参数都添加到缓存区
		argLength := make([]byte, argLengthInProtocol)
		for _, arg := range args {
			binary.BigEndian.PutUint32(argLength, uint32(len(arg)))
			request = append(request, argLength...)
			request = append(request, arg...)
		}
	}
	return writer.Write(request)
}
