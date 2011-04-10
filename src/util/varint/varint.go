package varint

import "os"

type VarInt uint64

func (value VarInt) Size() uint {
	value >>= 6

	if value == 0 {
		return 1
	}

	required := uint(1)
	for {
		value >>= 7
		required++

		if value == 0 {
			return required
		}
	}

	panic("Can't get here")
}

func (value VarInt) Write(target []byte) uint {
	// encode an int, assuming the first (current) byte has the last 7
	// bits allocated for storage, and all following bytes have the
	// full byte allocated
	target[0] |= byte(value & 0x3F)
	value >>= 6

	if value == 0 {
		target[0] |= 0x40
		return 1
	}

	idx := uint(1)
	for {
		target[idx] = byte(value & 0x7F)
		value >>= 7

		if value == 0 {
			target[idx] |= 0x80
			return idx + 1
		}

		idx++
	}

	panic("Can't get here")
}

func End(src []byte) (end uint, err os.Error) {
	for idx, byte := range src {
		if idx == 0 {
			if byte&0x40 == 0x40 {
				return 1, nil
			}
		} else {
			if byte&0x80 == 0x80 {
				return uint(idx+1), nil
			}
		}
	}

	return 0, os.NewError("Couldn't find end")
}

func Read(src []byte) (bytesRead uint, value VarInt) {
	// decode an int, assuming the first (current) byte has the last 7
	// bits allocated for storage, and all following bytes have the
	// full byte allocated
	value = VarInt(src[0] & 0x3F)

	if src[0]&0x40 == 0x40 {
		return 1, value
	}

	position := uint(1)
	shift := uint(6)
	for {
		value += (VarInt(src[position]&0x7F) << shift)

		if src[position]&0x80 == 0x80 {
			return position + 1, value
		}

		shift += 7
		position += 1
	}

	panic("Can't get here")
}
