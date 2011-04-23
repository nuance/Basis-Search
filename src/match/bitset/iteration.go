package bitset

import match "basis/match"

type BitSetIterator struct {
	b *BitSet

	doc match.DocId
	finished bool
}

func NewIter(b *BitSet) *BitSetIterator {
	firstBlock, finished := b.firstBlock(0)

	if finished {
		return &BitSetIterator{b, 0, true}
	}

	firstBit, _ := b.firstBit(firstBlock, 0)
	doc := docId(firstBlock, firstBit)
	return &BitSetIterator{b, doc, false}
}

func (b *BitSetIterator) Current() match.DocId {
	return b.doc
}

func (b *BitSetIterator) Finished() bool {
	return b.finished
}

func (b *BitSetIterator) Next() (match.DocId, bool) {
	if b.finished {
		panic("Next called on finished iterator")
	}

	block, bit := position(b.doc)
	// start one past the current position
	nextBlock, nextBit, _ := b.b.nextBit(block, bit+1)

	b.doc = docId(nextBlock, nextBit)
	b.finished = (b.doc == b.b.MaxId)

	return b.doc, b.finished
}

// TODO: what happens when you seek past then end?
func (b *BitSetIterator) Seek(target match.DocId) (match.DocId, bool) {
	if b.finished {
		panic("Seek called on finished iterator")
	}

	block, bit := position(target)
	nextBlock, nextBit, _ := b.b.nextBit(block, bit)

	b.doc = docId(nextBlock, nextBit)
	b.finished = (b.doc == b.b.MaxId)

	return b.doc, b.finished
}
