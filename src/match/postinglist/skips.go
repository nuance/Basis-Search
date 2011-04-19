package posting_list

import "os"
import "rand"

const SKIP_PAYLOAD = 12
const SKIP_UNINITIALIZED = 0
const SKIP_INITIALIZED = 1

const (
	SkipLayoutRandom = iota
	SkipLayoutNext
)

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

func (pl *PostingList) updateSkip(src, target Block) {
	pl.Raw[src.start] = SKIP_INITIALIZED

	writeUInt(pl.Raw[1:], target.nextBlockOffset)
	writeUInt64(pl.Raw[5:], target.nextDoc)
}

func (pl *PostingList) setupSkipsRandom() {
	// XXX: This is biased towards later skips. Need to dRaw with a
	// better distribution. Or maybe this is actually ok

	lastSkip := Block{}
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

func (pl *PostingList) setupSkipsNext() {
	lastSkip := Block{}

	for idx, skip := range pl.skips() {
		if idx == 0 {
			lastSkip = skip
			continue
		}

		pl.updateSkip(lastSkip, skip)
		lastSkip = skip
	}
}

func (pl *PostingList) BuildSkips(layoutOption int) (err os.Error) {
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
