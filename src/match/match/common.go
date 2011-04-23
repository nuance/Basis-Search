package match

import "os"
import heap "container/heap"

type DocId uint64

type MatchList interface {
	Add(DocId) os.Error
}

type MatchIterator interface {
	Current() DocId
	Finished() bool

	Next() (DocId, bool)
	Seek(target DocId) (DocId, bool)
}

type docIdHeap struct {
	iters []MatchIterator
}

func (d *docIdHeap) Len() int {
	return len(d.iters)
}

func (d *docIdHeap) Less(i, j int) bool {
	return (d.iters[i].Current() < d.iters[j].Current())
}

func (d *docIdHeap) Swap(i, j int) {
	d.iters[i], d.iters[j] = d.iters[j], d.iters[i]
}

func (d *docIdHeap) Push(x interface{}) {
	d.iters = append(d.iters, x.(MatchIterator))
}

func (d *docIdHeap) Pop() interface{} {
	lastPos := len(d.iters) - 1
	last := d.iters[lastPos]

	d.iters = d.iters[:lastPos]

	return last
}

func Merge(iters []MatchIterator, result MatchList) {
	if len(iters) == 0 {
		return
	}

	h := &docIdHeap{iters}
	heap.Init(h)

	last := DocId(0)
	for {
		i := heap.Pop(h).(MatchIterator)
		next := i.Current()

		if ! i.Finished() {
			heap.Push(h, i)
		}

		if next == last {
			continue
		} else {
			result.Add(next)
			last = next
		}
	}
}

func Intersection(iters []MatchIterator, result MatchList) {
	if len(iters) == 0 {
		return
	}

	next := DocId(0)
	for {
		changed := false

		for _, it := range iters {
			if it.Finished() {
				return
			}

			docId := it.Current()
			if next != docId {
				changed = true

				if next < docId {
					next = docId
				} else if next > docId {
					docId, done := it.Seek(next)

					// Prevent an extra loop
					if next < docId {
						next = docId
					}

					if done {
						return
					}
				}
			}
		}

		if !changed {
			result.Add(iters[0].Current())
			next += 1
		}
	}
}
