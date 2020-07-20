package sforeader

import (
	"encoding/binary"
	"testing"
)

func TestValidate(t *testing.T) {
	headers := []struct {
		name    string
		hdr     *header
		wantErr error
	}{
		{
			"valid_header",
			&header{
				Magic:             hdrMagic,
				Version:           hdrVersion,
				KeyTableOffset:    binary.LittleEndian.Uint32([]byte{5, 0, 0, 0}),
				DataTableOffset:   binary.LittleEndian.Uint32([]byte{15, 1, 0, 0}),
				IndexTableEntries: binary.LittleEndian.Uint32([]byte{12, 0, 0, 0}),
			},
			nil,
		},
		{
			"invalid_header_magic_word",
			&header{
				Magic:             hdrMagic + 1,
				Version:           hdrVersion,
				KeyTableOffset:    0,
				DataTableOffset:   0,
				IndexTableEntries: 0,
			},
			errInvalidMagic,
		},
		{
			"invalid_header_version",
			&header{
				Magic:             hdrMagic,
				Version:           0,
				KeyTableOffset:    0,
				DataTableOffset:   0,
				IndexTableEntries: 0,
			},
			errInvalidVersion,
		},
		{
			"invalid_header_key_offset",
			&header{
				Magic:             hdrMagic,
				Version:           hdrVersion,
				KeyTableOffset:    binary.LittleEndian.Uint32([]byte{0, 0, 1, 0}),
				DataTableOffset:   0,
				IndexTableEntries: 0,
			},
			errInvalidKeyOffset,
		},
		{
			"invalid_header_data_offset",
			&header{
				Magic:             hdrMagic,
				Version:           hdrVersion,
				KeyTableOffset:    0,
				DataTableOffset:   binary.LittleEndian.Uint32([]byte{0, 0, 1, 0}),
				IndexTableEntries: 0,
			},
			errInvalidDataOffset,
		},
		{
			"invalid_header_index_entries",
			&header{
				Magic:             hdrMagic,
				Version:           hdrVersion,
				KeyTableOffset:    0,
				DataTableOffset:   0,
				IndexTableEntries: binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
			},
			errInvalidIndexEntries,
		},
	}

	for _, header := range headers {
		t.Run(header.name, func(t *testing.T) {
			err := header.hdr.validate()
			if err != header.wantErr {
				t.Errorf("Validation failed, expected error \"%v\" but got \"%v\"", header.wantErr, err)
			}
		})
	}
}
