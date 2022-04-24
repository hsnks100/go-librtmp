package librtmp

import (
	"bytes"
	"encoding/binary"
)

func uintToLittleEndian(u uint64, b int) []byte {
	buff := new(bytes.Buffer)
	_ = binary.Write(buff, binary.LittleEndian, u)

	// 0x11 0x22
	// 0x22 0x11 0x00 0x00 0x00
	// 0x
	return buff.Bytes()[:b]
}

func uintToBigEndian(u uint64, b int) []byte {
	buff := new(bytes.Buffer)
	_ = binary.Write(buff, binary.BigEndian, u)

	// 0x11 0x22
	// 0x22 0x11 0x00 0x00 0x00
	// 0x
	return buff.Bytes()[8-b:]
}

func BigEndianToInt(b2 []byte) uint64 {
	b := make([]byte, len(b2))
	copy(b, b2)
	n := 8 - len(b)
	for i := 0; i < n; i++ {
		b = append([]byte{0x00}, b...)
	}
	data := binary.BigEndian.Uint64(b)
	return data

}

func LittleEndianToInt(b2 []byte) uint64 {
	b := make([]byte, len(b2))
	copy(b, b2)
	n := 8 - len(b)
	for i := 0; i < n; i++ {
		b = append(b, 0x00)
	}
	data := binary.LittleEndian.Uint64(b)
	return data

}
