package postinglist

import "fmt"
import "os"
import varint "basis/util/varint"
import match "basis/match"

type PostingList struct {
	Raw        []byte
	MaxId      match.DocId
}

func FromBytes(raw []byte) *PostingList {
	maxId := readUInt64(raw)
	raw = raw[8:]

	n, rawLen := varint.Read(raw)
	raw = raw[n:]
	raw = raw[:rawLen]

	return &PostingList{raw, match.DocId(maxId)}
}

func (pl *PostingList) Size() int {
	return len(pl.Raw) + int(varint.VarInt(len(pl.Raw)).Size()) + 8
}

func (pl *PostingList) ToBytes(dst []byte) {
	if cap(dst) < pl.Size() {
		panic("dst is too small")
	}

	writeUInt64(dst, uint64(pl.MaxId))
	dst = dst[8:]

	written := varint.VarInt(len(pl.Raw)).Write(dst[8:])
	dst = dst[written:]

	copy(dst, pl.Raw)
}

// Encoding scheme:
// 1st byte: high bit is 1 iff offset follows, 0 iff skip follows
// following bits: 7 bit varint, followed by overflow (stored
// least-significant bits first)

func (pl *PostingList) Add(doc match.DocId) (err os.Error) {
	numBlocks := uint(len(pl.Raw))
	if numBlocks > 0 && doc < pl.MaxId {
		return os.NewError("doc isn't larger than current max doc")
	}

	diff := varint.VarInt(doc - pl.MaxId)

	size := diff.Size()

	if size + numBlocks >= uint(cap(pl.Raw)) {
		return os.NewError("Out of space")
	}

	pl.Raw = pl.Raw[0:numBlocks+size]
	diff.Write(pl.Raw[numBlocks:])

	// Set the high bit
	pl.Raw[numBlocks] = blockTypeDoc | pl.Raw[numBlocks]
	pl.MaxId = doc

	return nil
}

type Block struct {
	// position in the underlying array
	start uint

	isSkip bool

	// always 1 for non-skips
	nextBlockOffset uint
	// the doc id for this block (same as the last for skip blocks)
	doc match.DocId

	// for skip blocks
	nextDoc match.DocId
}

func (pl PostingList) readBlock(idx uint, lastDoc match.DocId) (uint, Block) {
	bytes := pl.Raw[idx:]

	if bytes[0]&blockTypeDoc == blockTypeDoc {
		docSize, docOffset := varint.Read(bytes)

		doc := match.DocId(docOffset) + lastDoc
		data := Block{idx, false, 1, doc, 0}

		return docSize, data
	}

	nextBlockOffset := readUInt(bytes[1:])
	nextDocOffset := readUInt64(bytes[5:])

	data := Block{idx, true, nextBlockOffset, lastDoc, match.DocId(uint64(lastDoc) + nextDocOffset)}
	return 1 + SKIP_PAYLOAD, data
}

func (pl PostingList) blocks(visit func(Block)) {
	i := uint(0)
	lastDoc := match.DocId(0)

	// walk through the blocks
	numBlocks := uint(len(pl.Raw))
	for i < numBlocks {
		r, block := pl.readBlock(i, lastDoc)

		lastDoc = block.doc

		visit(block)

		i += r
	}
}

func (pl PostingList) skips() []Block {
	skips := []Block{}

	pl.blocks(func(b Block) {
		if b.isSkip {
			skips = append(skips, b)
		}
	})

	return skips
}

func (pl PostingList) Docs(visit func(match.DocId)) {
	pl.blocks(func(b Block) {
		if !b.isSkip {
			visit(b.doc)
		}
	})
}

func (pl PostingList) Stats() Stats {
	docCount := 0

	pl.Docs(func (match.DocId) { docCount += 1 })

	return Stats{docCount, uint64(pl.MaxId)}
}

func (b Block) initialized() bool {
	return b.nextBlockOffset != 0
}

func (pl *PostingList) String() string {
	numBlocks := len(pl.Raw)
	s := fmt.Sprintf("PostingList: %d of %d blocks used, max doc %d\n", numBlocks, cap(pl.Raw), pl.MaxId)

	pl.blocks(func(b Block) {
		s += b.String() + "\n"
	})

	return s
}

func (b Block) String() string {
	if b.isSkip && b.initialized() {
		return fmt.Sprintf("Skip - doc %d @ idx %d", b.nextDoc, b.start+b.nextBlockOffset)
	} else if b.isSkip {
		return fmt.Sprintf("Skip - uninitialized")
	} else {
		return fmt.Sprintf("Data - doc %d", b.doc)
	}

	panic("can't get here")
}
