package hexedit

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// SearchMode determines how search input is interpreted.
type SearchMode int

const (
	SearchString SearchMode = iota
	SearchBytes
	SearchU8
	SearchU16
	SearchU32
	SearchU64
)

var searchModeNames = []string{"string", "bytes", "u8", "u16", "u32", "u64"}

func (m SearchMode) String() string {
	if int(m) < len(searchModeNames) {
		return searchModeNames[m]
	}
	return "?"
}

// ParseSearchPattern converts user input into a byte pattern based on the mode.
func ParseSearchPattern(input string, mode SearchMode) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty search")
	}

	switch mode {
	case SearchString:
		return []byte(input), nil

	case SearchBytes:
		// Accept hex like "DE AD BE EF" or "DEADBEEF" or "de ad be ef".
		clean := strings.ReplaceAll(input, " ", "")
		b, err := hex.DecodeString(clean)
		if err != nil {
			return nil, fmt.Errorf("invalid hex: %w", err)
		}
		return b, nil

	case SearchU8:
		v, err := strconv.ParseUint(input, 0, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid u8: %w", err)
		}
		return []byte{byte(v)}, nil

	case SearchU16:
		v, err := strconv.ParseUint(input, 0, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid u16: %w", err)
		}
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(v))
		return b, nil

	case SearchU32:
		v, err := strconv.ParseUint(input, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid u32: %w", err)
		}
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(v))
		return b, nil

	case SearchU64:
		v, err := strconv.ParseUint(input, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid u64: %w", err)
		}
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, v)
		return b, nil
	}

	return nil, fmt.Errorf("unknown mode")
}

// Alignment constrains search to aligned boundaries.
type Alignment int

const (
	AlignNone Alignment = 0
	Align32   Alignment = 4
	Align64   Alignment = 8
)

var alignmentNames = map[Alignment]string{
	AlignNone: "Unaligned",
	Align32:   "32-bit",
	Align64:   "64-bit",
}

func (a Alignment) String() string {
	if name, ok := alignmentNames[a]; ok {
		return name
	}
	return "?"
}

// NextAlignment cycles to the next alignment value.
func NextAlignment(a Alignment) Alignment {
	switch a {
	case AlignNone:
		return Align32
	case Align32:
		return Align64
	default:
		return AlignNone
	}
}

// PrevAlignment cycles to the previous alignment value.
func PrevAlignment(a Alignment) Alignment {
	switch a {
	case AlignNone:
		return Align64
	case Align64:
		return Align32
	default:
		return AlignNone
	}
}

// FormatSearchPreview shows the byte representation of the search pattern.
func FormatSearchPreview(pattern []byte) string {
	if len(pattern) == 0 {
		return ""
	}
	parts := make([]string, len(pattern))
	for i, b := range pattern {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, " ")
}

// FormatSize returns a human-readable file size.
func FormatSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	f := float64(n)
	for _, unit := range []string{"KiB", "MiB", "GiB"} {
		f /= 1024
		if f < 1024 || unit == "GiB" {
			if f == math.Trunc(f) {
				return fmt.Sprintf("%.0f %s", f, unit)
			}
			return fmt.Sprintf("%.1f %s", f, unit)
		}
	}
	return fmt.Sprintf("%d B", n)
}
