// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var errInvalidThanksFile = errors.New("invalid rngit thanks file")

// ParseThanksCount decodes rngit's repository .thanks msgpack payload: {"count": N}.
func ParseThanksCount(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errInvalidThanksFile
	}

	offset := 0
	mapLen, n, err := readMsgpackMapHeader(data, offset)
	if err != nil {
		return 0, err
	}
	offset += n

	var count *int
	for i := 0; i < mapLen; i++ {
		key, n, err := readMsgpackString(data, offset)
		if err != nil {
			return 0, err
		}
		offset += n

		if key == "count" {
			value, n, err := readMsgpackUint(data, offset)
			if err != nil {
				return 0, err
			}
			offset += n
			v := int(value)
			count = &v
			continue
		}

		skip, err := skipMsgpackValue(data, offset)
		if err != nil {
			return 0, err
		}
		offset += skip
	}

	if count == nil {
		return 0, errInvalidThanksFile
	}
	return *count, nil
}

func readMsgpackMapHeader(data []byte, offset int) (int, int, error) {
	if offset >= len(data) {
		return 0, 0, errInvalidThanksFile
	}
	b := data[offset]
	switch {
	case b >= 0x80 && b <= 0x8f:
		return int(b & 0x0f), 1, nil
	case b == 0xde:
		if offset+3 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return int(binary.BigEndian.Uint16(data[offset+1 : offset+3])), 3, nil
	case b == 0xdf:
		if offset+5 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return int(binary.BigEndian.Uint32(data[offset+1 : offset+5])), 5, nil
	default:
		return 0, 0, fmt.Errorf("%w: unexpected map header %#x", errInvalidThanksFile, b)
	}
}

func readMsgpackString(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", 0, errInvalidThanksFile
	}
	b := data[offset]
	var size int
	var header int
	switch {
	case b >= 0xa0 && b <= 0xbf:
		size = int(b & 0x1f)
		header = 1
	case b == 0xd9:
		if offset+2 > len(data) {
			return "", 0, errInvalidThanksFile
		}
		size = int(data[offset+1])
		header = 2
	case b == 0xda:
		if offset+3 > len(data) {
			return "", 0, errInvalidThanksFile
		}
		size = int(binary.BigEndian.Uint16(data[offset+1 : offset+3]))
		header = 3
	case b == 0xdb:
		if offset+5 > len(data) {
			return "", 0, errInvalidThanksFile
		}
		size = int(binary.BigEndian.Uint32(data[offset+1 : offset+5]))
		header = 5
	default:
		return "", 0, fmt.Errorf("%w: unexpected string header %#x", errInvalidThanksFile, b)
	}
	start := offset + header
	end := start + size
	if end > len(data) {
		return "", 0, errInvalidThanksFile
	}
	return string(data[start:end]), end - offset, nil
}

func readMsgpackUint(data []byte, offset int) (uint64, int, error) {
	if offset >= len(data) {
		return 0, 0, errInvalidThanksFile
	}
	b := data[offset]
	switch {
	case b <= 0x7f:
		return uint64(b), 1, nil
	case b == 0xcc:
		if offset+2 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return uint64(data[offset+1]), 2, nil
	case b == 0xcd:
		if offset+3 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return uint64(binary.BigEndian.Uint16(data[offset+1 : offset+3])), 3, nil
	case b == 0xce:
		if offset+5 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return uint64(binary.BigEndian.Uint32(data[offset+1 : offset+5])), 5, nil
	case b == 0xcf:
		if offset+9 > len(data) {
			return 0, 0, errInvalidThanksFile
		}
		return binary.BigEndian.Uint64(data[offset+1 : offset+9]), 9, nil
	default:
		return 0, 0, fmt.Errorf("%w: unexpected integer header %#x", errInvalidThanksFile, b)
	}
}

func skipMsgpackValue(data []byte, offset int) (int, error) {
	consumed, err := skipMsgpackValueAt(data, offset)
	return consumed, err
}

func skipMsgpackValueAt(data []byte, offset int) (int, error) {
	if offset >= len(data) {
		return 0, errInvalidThanksFile
	}
	b := data[offset]
	switch {
	case b <= 0x7f, b >= 0xe0:
		return 1, nil
	case b >= 0x80 && b <= 0x8f:
		return skipMsgpackMap(data, offset, int(b&0x0f))
	case b >= 0x90 && b <= 0x9f:
		return skipMsgpackArray(data, offset, int(b&0x0f))
	case b >= 0xa0 && b <= 0xbf:
		return 1 + int(b&0x1f), nil
	case b == 0xc0:
		return 1, nil
	case b == 0xc2, b == 0xc3:
		return 1, nil
	case b == 0xc4, b == 0xd9:
		if offset+2 > len(data) {
			return 0, errInvalidThanksFile
		}
		return 2 + int(data[offset+1]), nil
	case b == 0xc5, b == 0xda:
		if offset+3 > len(data) {
			return 0, errInvalidThanksFile
		}
		return 3 + int(binary.BigEndian.Uint16(data[offset+1:offset+3])), nil
	case b == 0xc6, b == 0xdb:
		if offset+5 > len(data) {
			return 0, errInvalidThanksFile
		}
		return 5 + int(binary.BigEndian.Uint32(data[offset+1:offset+5])), nil
	case b == 0xcc:
		return 2, nil
	case b == 0xcd:
		return 3, nil
	case b == 0xce:
		return 5, nil
	case b == 0xcf:
		return 9, nil
	case b == 0xde:
		if offset+3 > len(data) {
			return 0, errInvalidThanksFile
		}
		return skipMsgpackMap(data, offset, int(binary.BigEndian.Uint16(data[offset+1:offset+3])))
	case b == 0xdf:
		if offset+5 > len(data) {
			return 0, errInvalidThanksFile
		}
		return skipMsgpackMap(data, offset, int(binary.BigEndian.Uint32(data[offset+1:offset+5])))
	case b == 0xdc:
		if offset+3 > len(data) {
			return 0, errInvalidThanksFile
		}
		return skipMsgpackArray(data, offset, int(binary.BigEndian.Uint16(data[offset+1:offset+3])))
	case b == 0xdd:
		if offset+5 > len(data) {
			return 0, errInvalidThanksFile
		}
		return skipMsgpackArray(data, offset, int(binary.BigEndian.Uint32(data[offset+1:offset+5])))
	default:
		return 0, fmt.Errorf("%w: unsupported msgpack value %#x", errInvalidThanksFile, b)
	}
}

func skipMsgpackMap(data []byte, offset, size int) (int, error) {
	start := offset
	_, header, err := readMsgpackMapHeader(data, offset)
	if err != nil {
		return 0, err
	}
	offset += header
	for i := 0; i < size; i++ {
		n, err := skipMsgpackValueAt(data, offset)
		if err != nil {
			return 0, err
		}
		offset += n
		n, err = skipMsgpackValueAt(data, offset)
		if err != nil {
			return 0, err
		}
		offset += n
	}
	return offset - start, nil
}

func skipMsgpackArray(data []byte, offset, size int) (int, error) {
	if offset >= len(data) {
		return 0, errInvalidThanksFile
	}
	start := offset
	b := data[offset]
	header := 1
	switch {
	case b >= 0x90 && b <= 0x9f:
	case b == 0xdc:
		header = 3
	case b == 0xdd:
		header = 5
	default:
		return 0, errInvalidThanksFile
	}
	offset += header
	for i := 0; i < size; i++ {
		n, err := skipMsgpackValueAt(data, offset)
		if err != nil {
			return 0, err
		}
		offset += n
	}
	return offset - start, nil
}
