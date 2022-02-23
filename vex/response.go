package vex

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	SuccessReply = 0 // 成功的答复码
	ErrorReply   = 1 // 发生错误的答复码
)

//解析从服务端发送过来的响应
func readResponseFrom(reader io.Reader) (reply byte, body []byte, err error) {
	header := make([]byte, headerLengthInProtocol)
	_, err = io.ReadFull(reader, header)
	if err != nil {
		return ErrorReply, nil, err
	}
	version := header[0]
	if version != ProtocolVersion {
		return ErrorReply, nil, errors.New("response " + ProtocolVersionMismatchErr.Error())
	}
	reply = header[1]
	header = header[2:]
	body = make([]byte, binary.BigEndian.Uint32(header))
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return ErrorReply, nil, err
	}
	return reply, body, nil
}

//服务端将响应写入到writer
func writeResponseTo(writer io.Writer, reply byte, body []byte) (int, error) {
	// 将响应体相关数据写入响应缓存区，并发送
	bodyLengthBytes := make([]byte, bodyLengthInProtocol)
	binary.BigEndian.PutUint32(bodyLengthBytes, uint32(len(body)))

	response := make([]byte, 2, headerLengthInProtocol+len(body))
	response[0] = ProtocolVersion
	response[1] = reply
	response = append(response, bodyLengthBytes...)
	response = append(response, body...)
	return writer.Write(response)
}

func writeErrorResponseTo(writer io.Writer, msg string) (int, error) {
	return writeResponseTo(writer, ErrorReply, []byte(msg))
}
