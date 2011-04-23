package main

import bufferpool "basis/util/bufferpool"
import postinglist "basis/match/postinglist"
import match "basis/match"

const initialSize = 100

type Index struct {
	index map[string] *bufferpool.Reference
	postingLists *bufferpool.BufferPool
}

func NewIndex() *Index {
	return &Index{make(map[string] *bufferpool.Reference), bufferpool.New(1024*1024)}
}

func (i *Index) Lookup(word string) match.MatchIterator {
	ref := i.index[word]

	if ref == nil {
		panic("word not found")
	}

	allocation := i.postingLists.Find(*ref)
	pl := postinglist.FromBytes(allocation.Raw)

	return postinglist.NewIter(pl)
}

func (i *Index) Replace(word string, pl *postinglist.PostingList) {
	if i.index[word] != nil {
		alloc := i.postingLists.Find(*i.index[word])
		alloc.Free()
		i.index[word] = nil
	}

	size := uint64(pl.Size())
	dst := i.postingLists.Alloc(size)

	pl.ToBytes(dst.Raw)
	i.index[word] = &dst.Ref
}
