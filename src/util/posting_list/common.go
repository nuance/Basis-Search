package posting_list

import "os"

type Payload interface {
	// Write out the payload, returns the size
	Write([]byte) uint
	// The encoded size
	Size() uint

	// Return the offset of the end of this payload. This method
	// should be as fast as possible (it's part of the innter loop of
	// any posting list operation).
	End([]byte) (uint, os.Error)
}

type Stats struct {
	DocCount int
	MaxId uint64
}

const blockTypeDoc = 0x80

func readUInt(bytes []byte) uint {
	return uint(bytes[0]<<24) + uint(bytes[1]<<16) + uint(bytes[2]<<8) + uint(bytes[3])
}

func writeUInt(bytes []byte, num uint) {
	bytes[0] = byte((num & 0xF000) >> 24)
	bytes[1] = byte((num & 0xF00) >> 16)
	bytes[2] = byte((num & 0xF0) >> 8)
	bytes[3] = byte(num & 0xF)
}

func readUInt64(bytes []byte) uint64 {
	return uint64(bytes[0]<<56) + uint64(bytes[1]<<48) + uint64(bytes[2]<<40) + uint64(bytes[3]<<32) + uint64(bytes[4]<<24) + uint64(bytes[5]<<16) + uint64(bytes[6]<<8) + uint64(bytes[7])
}

func writeUInt64(bytes []byte, num uint64) {
	bytes[0] = byte((num & 0xF0000000) >> 56)
	bytes[1] = byte((num & 0xF000000) >> 48)
	bytes[2] = byte((num & 0xF00000) >> 40)
	bytes[3] = byte((num & 0xF0000) >> 32)
	bytes[4] = byte((num & 0xF000) >> 24)
	bytes[5] = byte((num & 0xF00) >> 16)
	bytes[6] = byte((num & 0xF0) >> 8)
	bytes[7] = byte(num & 0xF)
}

