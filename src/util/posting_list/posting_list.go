package posting_list

import "fmt"
import "os"
import varint "basis/util/varint"

type PostingList struct {
	Raw        []byte
	MaxId      uint64
	PayloadEnd func([]byte) (uint, os.Error)
}

// Encoding scheme:
// 1st byte: high bit is 1 iff offset follows, 0 iff skip follows
// following bits: 7 bit varint, followed by overflow (stored
// least-significant bits first)

func (pl *PostingList) Add(doc uint64, payload Payload) (err os.Error) {
	numBlocks := uint(len(pl.Raw))
	if numBlocks > 0 && doc < pl.MaxId {
		return os.NewError("doc isn't larger than current max doc")
	}

	diff := varint.VarInt(doc - pl.MaxId)

	size := diff.Size()
	size += payload.Size()

	if size + numBlocks >= uint(cap(pl.Raw)) {
		return os.NewError("Out of space")
	}

	pl.Raw = pl.Raw[0:numBlocks+size]

	used := diff.Write(pl.Raw[numBlocks:])
	used += payload.Write(pl.Raw[numBlocks+used:])

	// Set the high bit
	pl.Raw[numBlocks] = blockTypeDoc | pl.Raw[numBlocks]
	pl.MaxId = doc

	return nil
}

func (pl *PostingList) addSkip(doc int) (err os.Error) {
	numBlocks := len(pl.Raw)
	if numBlocks+1+SKIP_PAYLOAD >= cap(pl.Raw) {
		return os.NewError("Out of space")
	}

	pl.Raw = pl.Raw[0:numBlocks+1+SKIP_PAYLOAD]

	// Mark the block as uninitialized
	pl.Raw[numBlocks] = SKIP_UNINITIALIZED

	return nil
}

type Block struct {
	// position in the underlying array
	start uint

	isSkip bool

	// always 1 for non-skips
	nextBlockOffset uint
	// the doc id for this block (same as the last for skip blocks)
	doc uint64

	// for skip blocks
	nextDoc uint64

	// for data blocks
	payload []byte
}

func readBlock(bytes []byte, idx uint, lastDoc uint64, payloadEnd func([]byte) (uint, os.Error)) (uint, Block) {
	if bytes[0]&blockTypeDoc == blockTypeDoc {
		docSize, docOffset := varint.Read(bytes)
		payloadSize, _ := payloadEnd(bytes[docSize:])
		payload := bytes[docSize:docSize+payloadSize]

		doc := uint64(docOffset) + lastDoc
		data := Block{idx, false, 1, doc, 0, payload}

		return docSize + payloadSize, data
	}

	nextBlockOffset := readUInt(bytes[1:])
	nextDocOffset := readUInt64(bytes[5:])

	data := Block{idx, true, nextBlockOffset, lastDoc, lastDoc + nextDocOffset, nil}
	return 1 + SKIP_PAYLOAD, data
}

func (pl PostingList) blocks(visit func(Block)) {
	i := uint(0)
	lastDoc := uint64(0)

	// walk through the blocks
	numBlocks := uint(len(pl.Raw))
	for i < numBlocks {
		r, block := readBlock(pl.Raw[i:], i, lastDoc, pl.PayloadEnd)

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

type Doc struct {
	Doc uint64
	Payload []byte
}

func (pl PostingList) Docs(visit func(Doc)) {
	pl.blocks(func(b Block) {
		if !b.isSkip {
			visit(Doc{b.doc, b.payload})
		}
	})
}

func (pl PostingList) Stats() Stats {
	docCount := 0

	pl.Docs(func (Doc) { docCount += 1 })

	return Stats{docCount, pl.MaxId}
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
		return fmt.Sprintf("Skip - doc %d @ idx %d", b.payload, b.start+b.nextBlockOffset)
	} else if b.isSkip {
		return fmt.Sprintf("Skip - uninitialized")
	} else {
		return fmt.Sprintf("Data - doc %d, payload %d", b.doc, b.payload)
	}

	panic("can't get here")
}
