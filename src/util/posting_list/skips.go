
func (pl *PostingList) updateSkip(src, target Block) {
	pl.Raw[src.start] = SKIP_INITIALIZED

	writeInt(pl.Raw[1:], target.nextBlockOffset)
	writeInt(pl.Raw[5:], target.payload)
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
	case SkiplayoutRandom:
		pl.setupSkipsRandom()
	case SkiplayoutNext:
		pl.setupSkipsNext()
	default:
		return os.NewError("Invalid layout option")
	}

	return nil
}
