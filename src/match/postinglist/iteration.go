package posting_list

import match "basis/match"

type PostingListIterator struct {
	pl       *PostingList
	b        Block
	finished bool

	last uint
	size uint
}

func NewIter(pl *PostingList) *PostingListIterator {
	read, first := pl.readBlock(0, 0)
	size := uint(len(pl.Raw))
	finished := read > size

	return &PostingListIterator{pl, first, finished, read, size}
}

func (i *PostingListIterator) advance() {
	read, b := i.pl.readBlock(i.last, i.b.doc)
	i.b = b
	i.last += read
	i.finished = (i.last > i.size)
}

func (i *PostingListIterator) Current() match.DocId {
	return i.b.doc
}

func (i *PostingListIterator) Finished() bool {
	return i.finished
}

func (i *PostingListIterator) Next() (match.DocId, bool) {
	if i.finished {
		panic("Called Next on a finished iterator")
	}

	i.advance()
	for i.b.isSkip && !i.finished {
		i.advance()
	}

	return i.b.doc, i.finished
}

func (i *PostingListIterator) Seek(target match.DocId) (match.DocId, bool) {
	if i.finished {
		panic("Called Seek on a finished iterator")
	} else if i.b.doc > target {
		panic("Can't seek backwards")
	} else if i.b.doc == target {
		return i.b.doc, false
	}

	for !i.finished && i.b.doc < target {
		if i.b.isSkip {
			if i.b.nextDoc <= target {
				// take the skip
				start := i.last + i.b.nextBlockOffset
				read, b := i.pl.readBlock(start, i.b.nextDoc)
				i.b = b

				i.last = i.b.start + read
				i.finished = (i.last > i.size)
			} else {
				// ignore it
				i.advance()
			}
		} else {
			i.advance()
		}
	}

	// We might have ended on a skip. Advance past that.
	for i.b.isSkip && !i.finished {
		i.advance()
	}

	return i.b.doc, i.finished
}
