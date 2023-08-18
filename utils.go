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
