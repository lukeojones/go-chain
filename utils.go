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

//func Contains(toFind int, toSearch []int) bool {
//	if toSearch == nil {
//		return false
//	}
//	for _, index := range toSearch {
//		if index == toFind {
//			return true
//		}
//	}
//	return false
//}
