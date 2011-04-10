
type Stats struct {
	DocCount int
	MaxId int
}

const blockTypeDoc = 0x80

const SKIP_PAYLOAD = 8
const SKIP_UNINITIALIZED = 0
const SKIP_INITIALIZED = 1

func readInt(bytes []byte) int {
	return int(bytes[0]<<24) + int(bytes[1]<<16) + int(bytes[2]<<8) + int(bytes[3])
}

func writeInt(bytes []byte, num int) {
	bytes[0] = byte((num & 0xF000) >> 24)
	bytes[1] = byte((num & 0xF00) >> 16)
	bytes[2] = byte((num & 0xF0) >> 8)
	bytes[3] = byte(num & 0xF)
}

const (
	SkipLayoutRandom = iota
	SkipLayoutNext
)

