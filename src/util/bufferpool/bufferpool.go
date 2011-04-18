package bufferpool

import "os"
import list "container/list"
import "sync"

type Error struct {
	os.ErrorString
}

type Allocation struct {
	Raw []byte

	buf *buffer
	chunk int
}

type buffer struct {
	chunks [][]byte
	freeList *list.List
	freeLock *sync.Mutex

	chunkSize uint64
	maxSize uint64
}

type BufferPool struct {
	buffers []*buffer

	MaxBufSize uint64
}

func newBuffer(chunkSize, maxSize uint64) *buffer {
	return &buffer{[][]byte{}, list.New(), new(sync.Mutex), chunkSize, maxSize}
}

func (b *buffer) alloc() (*Allocation) {
	b.freeLock.Lock()
	defer b.freeLock.Unlock()

	// first, check if we need to re-size
	if b.freeList.Len() == 0 {
		if uint64(len(b.chunks)) * b.chunkSize >= b.maxSize {
			// No more space to allocate
			return nil
		}

		chunk := make([]byte, b.chunkSize)
		b.chunks = append(b.chunks, chunk)

		return &Allocation{chunk, b, len(b.chunks)-1}
	} 

	// Pop the first item off the free list
	return b.freeList.Remove(b.freeList.Front()).(*Allocation)
}

func (b *buffer) free(a *Allocation) {
	b.freeList.PushBack(a)
}

func (p *BufferPool) Alloc(size uint64) *Allocation {
	for _, buffer := range p.buffers {
		if buffer.chunkSize == size {
			// alloc returns nil if buffer is full
			if a := buffer.alloc(); a != nil {
				return a
			}
		}
	}

	// Create a new buffer
	buf := newBuffer(size, p.MaxBufSize)
	p.buffers = append(p.buffers, buf)

	return buf.alloc()
}

func (a *Allocation) Free() {
	a.buf.free(a)
}
