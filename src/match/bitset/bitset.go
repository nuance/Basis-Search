package bitset

import "fmt"
import match "basis/match"
import "os"

type BitSet struct {
	backing []uint

	MaxId match.DocId
}

func New(capacity uint) *BitSet {
	return &BitSet{make([]uint, capacity / 32), 0}
}

func position(doc match.DocId) (block, bit uint) {
	block = uint(doc) / 32
	bit = uint(doc) % 32
	return
}

func docId(block, bit uint) (doc match.DocId) {
	return match.DocId(block * 32 + bit)
}

func (b *BitSet) Add(doc match.DocId) os.Error {
	block, bit := position(doc)

	if uint(cap(b.backing)) < block {
		return os.NewError(fmt.Sprintf("DocId is too large for this BitSet (required capacity %d, have %d)", block, cap(b.backing)))
	}

	b.backing[block] |= (1 << bit)

	if b.MaxId < doc {
		b.MaxId = doc
	}

	return nil
}

// Find the first non-empty block starting with backing[start]
func (b *BitSet) firstBlock(start uint) (uint, bool) {
	pos := start

	for pos := start; pos < uint(len(b.backing)) && b.backing[pos] == 0; pos++ {
		// Scan until a non-empty block or we reach the end
	}

	return pos, (pos == uint(len(b.backing)))
}

// table for zeroing out leading bits. Position 5 => zeroes bits 0-4
var wipe []uint = []uint{
	0xFFFFFFFF - 0x00000000,
	0xFFFFFFFF - 0x00000001,
	0xFFFFFFFF - 0x00000003,
	0xFFFFFFFF - 0x00000007,
	0xFFFFFFFF - 0x0000000F,
	0xFFFFFFFF - 0x0000001F,
	0xFFFFFFFF - 0x0000003F,
	0xFFFFFFFF - 0x0000007F,
	0xFFFFFFFF - 0x000000FF,
	0xFFFFFFFF - 0x000001FF,
	0xFFFFFFFF - 0x000003FF,
	0xFFFFFFFF - 0x000007FF,
	0xFFFFFFFF - 0x00000FFF,
	0xFFFFFFFF - 0x00001FFF,
	0xFFFFFFFF - 0x00003FFF,
	0xFFFFFFFF - 0x00007FFF,
	0xFFFFFFFF - 0x0000FFFF,
	0xFFFFFFFF - 0x0001FFFF,
	0xFFFFFFFF - 0x0003FFFF,
	0xFFFFFFFF - 0x0007FFFF,
	0xFFFFFFFF - 0x000FFFFF,
	0xFFFFFFFF - 0x001FFFFF,
	0xFFFFFFFF - 0x003FFFFF,
	0xFFFFFFFF - 0x007FFFFF,
	0xFFFFFFFF - 0x00FFFFFF,
	0xFFFFFFFF - 0x01FFFFFF,
	0xFFFFFFFF - 0x03FFFFFF,
	0xFFFFFFFF - 0x07FFFFFF,
	0xFFFFFFFF - 0x0FFFFFFF,
	0xFFFFFFFF - 0x1FFFFFFF,
	0xFFFFFFFF - 0x3FFFFFFF,
	0xFFFFFFFF - 0x7FFFFFFF,
	0xFFFFFFFF - 0xFFFFFFFF,
}

var deBruijn []uint = []uint{0, 1, 28, 2, 29, 14, 24, 3, 30, 22, 20, 15, 25, 17, 4, 8, 31, 27, 13, 23, 21, 19, 16, 7, 26, 12, 18, 6, 11, 5, 10, 9}
const deBruijnMask uint = 0x077CB531

func (b *BitSet) firstBit(block, start uint) (uint, bool) {
	val := b.backing[block] & wipe[start]

	if val == 0 {
		return 0, true
	}

	return deBruijn[(((val & -val) * deBruijnMask)) >> 27], false
}

// Return the first set bit starting with block, bit
func (b *BitSet) nextBit(block, bit uint) (nextBlock, nextBit uint, finished bool) {
	// Allow the caller to increment the bit
	if bit > 32 {
		block += bit / 32
		bit = bit % 32
	}

	nBlock, finished := b.firstBlock(block)

	if finished {
		return 0, 0, finished
	}

	nBit := uint(0)
	if nBlock == block {
		nBit = bit
	}

	nBit, _ = b.firstBit(nBlock, nBit)

	return nBlock, nBit, false
}
