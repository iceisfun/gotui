package hexedit

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Buffer holds the file data for the hex editor.
type Buffer struct {
	data     []byte
	path     string
	readonly bool
}

// OpenFile loads a file into memory.
func OpenFile(path string) (*Buffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("hexedit: open: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("hexedit: read: %w", err)
	}

	return &Buffer{data: data, path: path}, nil
}

// NewBuffer creates a buffer from raw bytes.
func NewBuffer(data []byte, name string) *Buffer {
	return &Buffer{data: data, path: name}
}

// Len returns the number of bytes.
func (b *Buffer) Len() int { return len(b.data) }

// Path returns the file path (or name).
func (b *Buffer) Path() string { return b.path }

// Byte returns the byte at offset, or 0 if out of range.
func (b *Buffer) Byte(offset int) byte {
	if offset < 0 || offset >= len(b.data) {
		return 0
	}
	return b.data[offset]
}

// Bytes returns a slice of bytes starting at offset, up to n bytes.
func (b *Buffer) Bytes(offset, n int) []byte {
	if offset < 0 || offset >= len(b.data) {
		return nil
	}
	end := offset + n
	if end > len(b.data) {
		end = len(b.data)
	}
	return b.data[offset:end]
}

// SetByte modifies a single byte.
func (b *Buffer) SetByte(offset int, val byte) {
	if offset >= 0 && offset < len(b.data) && !b.readonly {
		b.data[offset] = val
	}
}

// InterpretAt returns numeric interpretations of bytes at the given offset.
func (b *Buffer) InterpretAt(offset int) Interpretation {
	var interp Interpretation
	remaining := len(b.data) - offset
	if offset < 0 || remaining <= 0 {
		return interp
	}

	interp.U8 = b.data[offset]
	interp.I8 = int8(interp.U8)
	interp.HasU8 = true

	if remaining >= 2 {
		interp.U16LE = binary.LittleEndian.Uint16(b.data[offset:])
		interp.U16BE = binary.BigEndian.Uint16(b.data[offset:])
		interp.I16LE = int16(interp.U16LE)
		interp.I16BE = int16(interp.U16BE)
		interp.HasU16 = true
	}

	if remaining >= 4 {
		interp.U32LE = binary.LittleEndian.Uint32(b.data[offset:])
		interp.U32BE = binary.BigEndian.Uint32(b.data[offset:])
		interp.I32LE = int32(interp.U32LE)
		interp.I32BE = int32(interp.U32BE)
		interp.HasU32 = true
	}

	if remaining >= 8 {
		interp.U64LE = binary.LittleEndian.Uint64(b.data[offset:])
		interp.U64BE = binary.BigEndian.Uint64(b.data[offset:])
		interp.I64LE = int64(interp.U64LE)
		interp.I64BE = int64(interp.U64BE)
		interp.HasU64 = true
	}

	return interp
}

// Search searches for a byte pattern starting at fromOffset.
// Returns the offset or -1 if not found.
func (b *Buffer) Search(pattern []byte, fromOffset int) int {
	return b.SearchAligned(pattern, fromOffset, 0)
}

// SearchAligned searches for a byte pattern with alignment constraints.
// align=0 or 1 means no alignment, align=4 means 32-bit boundaries, etc.
func (b *Buffer) SearchAligned(pattern []byte, fromOffset int, align int) int {
	if len(pattern) == 0 || fromOffset < 0 {
		return -1
	}
	if align <= 1 {
		align = 1
	}

	// Snap fromOffset up to the next aligned boundary.
	start := fromOffset
	if start%align != 0 {
		start += align - start%align
	}

	for i := start; i <= len(b.data)-len(pattern); i += align {
		match := true
		for j := range pattern {
			if b.data[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// Interpretation holds numeric interpretations at a position.
type Interpretation struct {
	U8    uint8
	I8    int8
	HasU8 bool

	U16LE, U16BE uint16
	I16LE, I16BE int16
	HasU16       bool

	U32LE, U32BE uint32
	I32LE, I32BE int32
	HasU32       bool

	U64LE, U64BE uint64
	I64LE, I64BE int64
	HasU64       bool
}
