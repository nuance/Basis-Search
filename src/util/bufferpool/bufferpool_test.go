package bufferpool

import "testing"

func verify(t *testing.T, b *buffer, a *Allocation, chunk int, expCap int, expLen int) {
	if a == nil {
		t.Errorf("Allocation is nil")
		return
	}

	if a.buf != b {
		t.Errorf("buf is incorrect")
	}

	if a.chunk != chunk {
		t.Errorf("chunk is incorrect (expected %d, got %d)", chunk, a.chunk)
	}

	if cap(a.Raw) != expCap {
		t.Errorf("cap(a.Raw) = %d, want %d.", cap(a.Raw), expCap)
	}

	if len(a.Raw) != expLen {
		t.Errorf("len(a.Raw) = %d, want %d.", len(a.Raw), expLen)
	}
}

func TestBuffer(t *testing.T) {
	b := newBuffer(100, 100)

	a := b.alloc()

	verify(t, b, a, 0, 100, 0)

	a.Raw = append(a.Raw, 0x1)

	a.Free()

	re := b.alloc()

	if a != re {
		t.Error("Should have returned previous allocation")
	}

	verify(t, b, a, 0, 100, 0)

	s := b.alloc()

	if s != nil {
		t.Errorf("Allocation should have failed (buffer is maxed out)")
	}
}

func TestBufferPool(t *testing.T) {
	p := New(100)

	a1 := p.Alloc(50)
	a2 := p.Alloc(50)
	a3 := p.Alloc(50)

	if a1.buf != a2.buf {
		t.Errorf("original buffer is not yet full")
	}

	if a2.buf == a3.buf {
		t.Errorf("original buffer should be full")
	}

	a2.Free()

	a4 := p.Alloc(50)

	if a4.buf != a1.buf {
		t.Errorf("new allocation should have been from first buffer")
	}
}
