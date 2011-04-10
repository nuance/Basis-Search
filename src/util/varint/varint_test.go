package varint

import "testing"

type sizeTest struct {
	in, out VarInt
}

var sizeTests = []sizeTest{
	sizeTest{0, 1},
	sizeTest{1, 1},
	sizeTest{(1 << 6) - 1, 1},
	sizeTest{(1 << 6), 2},
	sizeTest{(1 << 13) - 1, 2},
	sizeTest{(1 << 13), 3},
	sizeTest{0x3F, 1},
}

func TestSize(t *testing.T) {
	for _, dt := range sizeTests {
		v := dt.in.Size()
		if VarInt(v) != dt.out {
			t.Errorf("Size(%d) = %d, want %d.", dt.in, v, dt.out)
		}
	}
}

type encodeTest struct {
	in  VarInt
	out []byte
}

var encodeTests = []encodeTest{
	{0, []byte{0x40}},
	{1, []byte{0x41}},
	{(1 << 6) - 1, []byte{0x7F}},
	{(1 << 6), []byte{0x0, 0x81}},
	{(1 << 13) - 1, []byte{0x3F, 0xFF}},
	{(1 << 13), []byte{0x0, 0x0, 0x81}},
	{1000, []byte{0x28, 0x8F}},
	{9000, []byte{0x28, 0xC, 0x81}},
}

func TestEncode(t *testing.T) {
	for _, dt := range encodeTests {
		buffer := make([]byte, len(dt.out))

		v := dt.in.Write(buffer)

		if v != dt.in.Size() {
			t.Errorf("Encode(%d, buffer) = %d, want size %d.", dt.in, v, dt.in.Size())
		}

		for idx, v := range buffer {
			if v != dt.out[idx] {
				t.Errorf("Encode(%d, buffer): buffer[%d] = %d, want %d.", dt.in, idx, v, dt.out[idx])
			}
		}

		read, decoded := Read(buffer)

		if read != v {
			t.Errorf("Decode(Encode(%d)) bytes = %d, want %d", dt.in, read, v)
		}

		if decoded != dt.in {
			t.Errorf("Decode(Encode(%d)) = %d", dt.in, decoded)
		}
	}
}
