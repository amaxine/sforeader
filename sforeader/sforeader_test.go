package sforeader

import (
	"encoding/binary"
	"testing"
)

func TestValidate(t *testing.T) {
	headers := []struct {
		h    *header
		name string
		e    error
	}{
		{&header{
			Magic:             1179865088,
			Version:           257,
			KeyTableOffset:    binary.LittleEndian.Uint32([]byte{5, 0, 0, 0}),
			DataTableOffset:   binary.LittleEndian.Uint32([]byte{15, 1, 0, 0}),
			IndexTableEntries: binary.LittleEndian.Uint32([]byte{12, 0, 0, 0}),
		}, "valid_header", nil},
		{&header{
			Magic:             1179865087,
			Version:           257,
			KeyTableOffset:    0,
			DataTableOffset:   0,
			IndexTableEntries: 0,
		}, "invalid_header_magic_word", errInvalidMagic},
		{&header{
			Magic:             1179865088,
			Version:           256,
			KeyTableOffset:    0,
			DataTableOffset:   0,
			IndexTableEntries: 0,
		}, "invalid_header_version", errInvalidVersion},
		{&header{
			Magic:             1179865088,
			Version:           257,
			KeyTableOffset:    binary.LittleEndian.Uint32([]byte{0, 0, 1, 0}),
			DataTableOffset:   0,
			IndexTableEntries: 0,
		}, "invalid_header_key_offset", errInvalidKeyOffset},
		{&header{
			Magic:             1179865088,
			Version:           257,
			KeyTableOffset:    0,
			DataTableOffset:   binary.LittleEndian.Uint32([]byte{0, 0, 1, 0}),
			IndexTableEntries: 0,
		}, "invalid_header_data_offset", errInvalidDataOffset},
		{&header{
			Magic:             1179865088,
			Version:           257,
			KeyTableOffset:    0,
			DataTableOffset:   0,
			IndexTableEntries: binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
		}, "invalid_header_index_entries", errInvalidIndexEntries},
	}

	for _, header := range headers {
		t.Run(header.name, func(t *testing.T) {
			err := header.h.validate()
			if err != header.e {
				t.Errorf("Validation failed, expected error \"%v\" but got \"%v\"", header.e, err)
			}
		})
	}
}
