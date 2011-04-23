package bufferpool

import list "container/list"
import "sync"

type Reference struct {
	Buffer, Chunk int
}

type Allocation struct {
	Raw []byte
	Ref Reference

	buf *buffer
}

type buffer struct {
	chunks [][]byte
	freeList *list.List
	freeLock *sync.Mutex

	chunkSize uint64
	maxSize uint64
	chunkNum int
}

type BufferPool struct {
	buffers []*buffer

	MaxBufSize uint64
}

func newBuffer(chunkNum int, chunkSize, maxSize uint64) *buffer {
	return &buffer{[][]byte{}, list.New(), new(sync.Mutex), chunkSize, maxSize, chunkNum}
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

		chunk := make([]byte, 0, b.chunkSize)
		b.chunks = append(b.chunks, chunk)

		ref := Reference{b.chunkNum, len(b.chunks)-1}
		return &Allocation{chunk, ref, b}
	} 

	// Pop the first item off the free list
	return b.freeList.Remove(b.freeList.Front()).(*Allocation)
}

func (b *buffer) free(a *Allocation) {
	// reset the slice
	a.Raw = a.Raw[:0]
	b.freeList.PushBack(a)
}

func New(MaxBufSize uint64) *BufferPool {
	return &BufferPool{[]*buffer{}, MaxBufSize}
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
	buf := newBuffer(len(p.buffers), size, p.MaxBufSize)
	p.buffers = append(p.buffers, buf)

	return buf.alloc()
}

func (p *BufferPool) Find(ref Reference) *Allocation {
	buffer := p.buffers[ref.Buffer]
	raw := buffer.chunks[ref.Chunk]

	return &Allocation{raw, ref, buffer}
}

func (a *Allocation) Free() {
	a.buf.free(a)
}
