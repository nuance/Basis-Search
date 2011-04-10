package posting_list

import "fmt"
import "io"
import "rand"
import "os"
import varint "basis/util/varint"

type SortedVarIntList struct {
	Raw       []byte
	NumBlocks int

	MaxId int
}

type Stats struct {
	DocCount int
	MaxId int
}

// Encoding scheme:
// 1st byte: high bit is 1 iff offset follows, 0 iff skip follows
// following bits: 7 bit varint, followed by overflow (stored
// least-significant bits first)

func New(blocks []byte) *SortedVarIntList {
	return &SortedVarIntList{blocks, 0, 0}
}

func (pl SortedVarIntList) Write(w io.Writer) (n int, error os.Error) {
	return fmt.Fprintf(w, "%x%x%x%s", pl.NumBlocks, pl.MaxId, len(pl.Raw), pl.Raw)
}

func Read(r io.Reader, pl *SortedVarIntList) (n int, error os.Error) {
	var NumBlocks, MaxId, blockCount int

	n, error = fmt.Fscanf(r, "%x%x%x", NumBlocks, MaxId, blockCount)

	if error != nil {
		return n, error
	}

	blocks := make([]byte, blockCount)

	br, error := r.Read(blocks)

	if error != nil {
		return n + br, error
	}

	if br < blockCount {
		return n + br, os.NewError("Read too few bytes")
	}

	pl.MaxId = MaxId
	pl.NumBlocks = NumBlocks
	pl.Raw = blocks

	return n + br, nil
}

const blockTypeDoc = 0x80

func (pl *SortedVarIntList) Add(doc, payload int) (err os.Error) {
	if pl.NumBlocks > 0 && doc < pl.MaxId {
		return os.NewError("doc isn't larger than current max doc")
	}

	diff := doc - pl.MaxId
	size := varint.Size(diff)

	size += varint.Size(payload)

	if pl.NumBlocks+size >= len(pl.Raw) {
		return os.NewError("Out of space")
	}

	used := varint.Encode(diff, pl.Raw[pl.NumBlocks:])
	varint.Encode(payload, pl.Raw[pl.NumBlocks+used:])

	// Set the high bit
	pl.Raw[pl.NumBlocks] = blockTypeDoc | pl.Raw[pl.NumBlocks]

	pl.NumBlocks += size
	pl.MaxId = doc

	return nil
}

const SKIP_PAYLOAD = 8
const SKIP_UNINITIALIZED = 0
const SKIP_INITIALIZED = 1

func (pl *SortedVarIntList) addSkip(doc int) (err os.Error) {
	if pl.NumBlocks+1+SKIP_PAYLOAD >= len(pl.Raw) {
		return os.NewError("Out of space")
	}

	// Mark the block as uninitialized
	pl.Raw[pl.NumBlocks] = SKIP_UNINITIALIZED
	pl.NumBlocks += 1 + SKIP_PAYLOAD

	return nil
}

func readInt(bytes []byte) int {
	return int(bytes[0]<<24) + int(bytes[1]<<16) + int(bytes[2]<<8) + int(bytes[3])
}

func writeInt(bytes []byte, num int) {
	bytes[0] = byte((num & 0xF000) >> 24)
	bytes[1] = byte((num & 0xF00) >> 16)
	bytes[2] = byte((num & 0xF0) >> 8)
	bytes[3] = byte(num & 0xF)
}

type block struct {
	// position in the underlying array
	start int

	isSkip bool

	// always 1 for non-skips
	nextBlockOffset int
	// the doc id for this block (same as the last for skip blocks)
	doc int
	// for skip blocks, the next doc id, for data blocks, the payload
	payload int
}

func readBlock(bytes []byte, idx int, lastDoc int) (int, block) {
	if bytes[0]&blockTypeDoc == blockTypeDoc {
		docSize, docOffset := varint.Decode(bytes)
		payloadSize, payload := varint.Decode(bytes[docSize:])

		doc := docOffset + lastDoc
		data := block{idx, false, 1, doc, payload}

		return docSize + payloadSize, data
	}

	nextBlockOffset := readInt(bytes[1:])
	nextDocOffset := readInt(bytes[5:])

	data := block{idx, true, nextBlockOffset, lastDoc, lastDoc + nextDocOffset}
	return 1 + SKIP_PAYLOAD, data
}

func (pl SortedVarIntList) blocks(visit func(block)) {
	i := 0
	lastDoc := 0

	// walk through the blocks
	for i < pl.NumBlocks {
		r, block := readBlock(pl.Raw[i:], i, lastDoc)

		lastDoc = block.doc

		visit(block)

		i += r
	}
}

func (pl SortedVarIntList) skips() []block {
	skips := []block{}

	pl.blocks(func(b block) {
		if b.isSkip {
			skips = append(skips, b)
		}
	})

	return skips
}

type Doc struct {
	Doc, Payload int
}

func (pl SortedVarIntList) Docs(visit func(Doc)) {
	pl.blocks(func(b block) {
		if !b.isSkip {
			visit(Doc{b.doc, b.payload})
		}
	})
}

func (pl SortedVarIntList) Stats() Stats {
	docCount := 0

	pl.Docs(func (Doc) { docCount += 1 })

	return Stats{docCount, pl.MaxId}
}

func (b block) initialized() bool {
	return b.nextBlockOffset != 0
}

func (pl *SortedVarIntList) updateSkip(src block, target block) {
	pl.Raw[src.start] = SKIP_INITIALIZED

	writeInt(pl.Raw[1:], target.nextBlockOffset)
	writeInt(pl.Raw[5:], target.payload)
}

func (pl *SortedVarIntList) setupSkipsRandom() {
	// XXX: This is biased towards later skips. Need to dRaw with a
	// better distribution. Or maybe this is actually ok

	lastSkip := block{}
	skips := pl.skips()

	for idx, skip := range skips {
		if idx == 0 {
			lastSkip = skip
			continue
		}

		// Pick a random upcoming skip
		goal := rand.Intn(len(skips)-idx) + idx

		pl.updateSkip(lastSkip, skips[goal])
		lastSkip = skip
	}
}

func (pl *SortedVarIntList) setupSkipsNext() {
	lastSkip := block{}

	for idx, skip := range pl.skips() {
		if idx == 0 {
			lastSkip = skip
			continue
		}

		pl.updateSkip(lastSkip, skip)
		lastSkip = skip
	}
}

const (
	SkipLayoutRandom = iota
	SkipLayoutNext
)

func (pl *SortedVarIntList) BuildSkips(layoutOption int) (err os.Error) {
	switch layoutOption {
	case SkipLayoutRandom:
		pl.setupSkipsRandom()
	case SkipLayoutNext:
		pl.setupSkipsNext()
	default:
		return os.NewError("Invalid layout option")
	}

	return nil
}

func (pl *SortedVarIntList) String() string {
	s := fmt.Sprintf("SortedVarIntList: %d of %d blocks used, max doc %d\n", pl.NumBlocks, cap(pl.Raw), pl.MaxId)

	pl.blocks(func(b block) {
		s += b.String() + "\n"
	})

	return s
}

func (b block) String() string {
	if b.isSkip && b.initialized() {
		return fmt.Sprintf("Skip - doc %d @ idx %d", b.payload, b.start+b.nextBlockOffset)
	} else if b.isSkip {
		return fmt.Sprintf("Skip - uninitialized")
	} else {
		return fmt.Sprintf("Data - doc %d, payload %d", b.doc, b.payload)
	}

	panic("can't get here")
}
