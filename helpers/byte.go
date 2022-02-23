package helpers

import "strconv"

func Copy(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func JoinAddressAndPort(address string, port int) string {
	return address + ":" + strconv.Itoa(port)
}
