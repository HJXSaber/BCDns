package utils

import (
	"bytes"
	"encoding/binary"
	"os"
)

func DBExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// merge map1 into map2
func CoverMap(map1, map2 map[string]string) map[string]string {
	for k, v := range map1 {
		if _, ok := map2[k]; !ok {
			map2[k] = v
		}
	}
	return map2
}

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}