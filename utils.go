package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func Int64ToBytes(i int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		fmt.Println("Failed to write int64 to buffer:", err)
	}
	return buf.Bytes()
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
