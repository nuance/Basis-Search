package bufferpool

import "testing"

func TestBuffer(t *testing.T) {
	b := newBuffer(100, 100)

	a := b.alloc()

	if a == nil {
		t.Errorf("First allocation is nil")
		return
	}

	if a.buf != b {
		t.Errorf("a.buf is incorrect")
	}

	if a.chunk != 0 {
		t.Errorf("a.chunk is incorrect")
	}

	if cap(a.Raw) != 100 {
		t.Errorf("cap(a.Raw) = %d, want %d.", cap(a.Raw), 100)
	}

	if len(a.Raw) != 0 {
		t.Errorf("len(a.Raw) = %d, want %d.", len(a.Raw), 0)
	}

	a.Raw = append(a.Raw, 0x1)

	a.Free()

	re := b.alloc()

	if a != re {
		t.Error("Should have returned previous allocation")
	}

	if cap(a.Raw) != 100 {
		t.Errorf("cap(a.Raw) = %d, want %d.", cap(a.Raw), 100)
	}

	if len(a.Raw) != 0 {
		t.Errorf("len(a.Raw) = %d, want %d.", len(a.Raw), 0)
	}

	s := b.alloc()

	if s != nil {
		t.Errorf("Allocation should have failed (buffer is maxed out)")
	}
}


