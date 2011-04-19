package posting_list

import "os"

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

func (i *PostingListIterator) Next() (Block, os.Error) {
	if i.finished {
		panic("Called Next on a finished iterator")
	}

	i.advance()
	for i.b.isSkip && !i.finished {
		i.advance()
	}

	if i.finished {
		return i.b, os.NewError("Finished")
	}

	return i.b, nil
}

func (i *PostingListIterator) Seek(target uint64) (Block, os.Error) {
	if i.finished {
		panic("Called Seek on a finished iterator")
	} else if i.b.doc > target {
		panic("Can't seek backwards")
	} else if i.b.doc == target {
		return i.b, nil
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

	if i.finished {
		return i.b, os.NewError("Finished")
	}

	return i.b, nil
}

type Docs struct {
	Doc uint64
	Payloads [][]byte
}

func Intersection(pls []*PostingList, out chan<- Docs) {
	defer close(out)

	if len(pls) == 0 {
		return
	}

	iters := []*PostingListIterator{}

	for _, pl := range pls {
		iters = append(iters, NewIter(pl))
	}

	next := uint64(0)
	blocks := make([]Block, len(iters))

	for {
		changed := false

		for idx, it := range iters {
			if it.finished {
				return
			}

			if next != it.b.doc {
				changed = true

				if next < it.b.doc {
					next = it.b.doc
				} else if next > it.b.doc {
					_, err := it.Seek(next)

					// Prevent an extra loop
					if next < it.b.doc {
						next = it.b.doc
					}

					if err != nil {
						return
					}
				}
			}

			blocks[idx] = it.b
		}

		if !changed {
			payloads := make([][]byte, len(blocks))
			for idx, doc := range blocks {
				payloads[idx] = doc.payload
			}

			out <- Docs{blocks[0].doc, payloads}

			next += 1
		}
	}
}
