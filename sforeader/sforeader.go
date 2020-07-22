package sforeader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// Constants defining data types

// FormatUtf8SM UTF-8 Special Mode format (not null terminated)
const FormatUtf8SM = 4

// FormatUtf8 UTF-8 format
const FormatUtf8 = 516

// FormatInteger Int format
const FormatInteger = 1028

// Error types
var errInvalidMagic = errors.New("PSF header invalid")
var errInvalidVersion = errors.New("version invalid")
var errInvalidKeyOffset = errors.New("key table offset invalid")
var errInvalidDataOffset = errors.New("data table offset invalid")
var errInvalidIndexEntries = errors.New("index table entries invalid")

var hdrMagic = int32(binary.LittleEndian.Uint32([]byte{0, 0x50, 0x53, 0x46}))
var hdrVersion = int32(binary.LittleEndian.Uint32([]byte{1, 1, 0, 0}))

// Expected SFO header format
type header struct {
	Magic             int32
	Version           int32
	KeyTableOffset    uint32
	DataTableOffset   uint32
	IndexTableEntries uint32
}

// Index table listing placement and format of entries
type indexTable struct {
	KeyTableOffset  uint16
	ParamFormat     uint16 // One of FormatUtf8/FormatUtf8SM/FormatInteger
	ParamLength     uint32
	ParamMaxLength  uint32
	DataTableOffset uint32
}

// Data structure of individual entries
type Data struct {
	Data   []byte
	Format uint16 // One of FormatUtf8/FormatUtf8SM/FormatInteger
}

// Basic validation. Expanding scope beyond Vita would probably require changes here.
// This only verifies the header is valid and doesn't go beyond it.
func (h *header) validate() error {
	if h.Magic != hdrMagic {
		return errInvalidMagic
	}
	if h.Version != hdrVersion {
		return errInvalidVersion
	}
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, h.KeyTableOffset)
	if bytes.Compare(tmp[2:4], []byte{0, 0}) != 0 {
		return errInvalidKeyOffset
	}
	binary.LittleEndian.PutUint32(tmp, h.DataTableOffset)
	if bytes.Compare(tmp[2:4], []byte{0, 0}) != 0 {
		return errInvalidDataOffset
	}
	binary.LittleEndian.PutUint32(tmp, h.IndexTableEntries)
	if bytes.Compare(tmp[1:4], []byte{0, 0, 0}) != 0 {
		return errInvalidIndexEntries
	}
	return nil
}

// ParseFile is a function for parsing sfo files and returning their set properties
func ParseFile(filename string) (map[string]Data, error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	hdr := header{}
	err = binary.Read(f, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}
	err = hdr.validate()
	if err != nil {
		return nil, err
	}

	index := make([]indexTable, hdr.IndexTableEntries)
	for i := 0; i < int(hdr.IndexTableEntries); i++ {
		li := indexTable{}
		err = binary.Read(f, binary.LittleEndian, &li)
		if err != nil {
			return nil, err
		}
		index[i] = li
	}

	dataMap := make(map[string]Data)
	for i := 0; i < int(hdr.IndexTableEntries); i++ {
		_, err = f.Seek(int64(hdr.KeyTableOffset)+int64(index[i].KeyTableOffset), 0)
		if err != nil {
			return nil, err
		}

		// Awful but simple way of reading the key until \x00, perhaps a bit unsafe
		var key string
		for {
			var c byte
			err = binary.Read(f, binary.LittleEndian, &c)
			if err != nil {
				return nil, err
			}

			if c == 0 {
				break
			}
			key = key + string(c)
		}

		_, err = f.Seek(int64(hdr.DataTableOffset)+int64(index[i].DataTableOffset), 0)
		if err != nil {
			return nil, err
		}

		value := make([]byte, int(index[i].ParamLength))
		err = binary.Read(f, binary.LittleEndian, value)
		if err != nil {
			return nil, err
		}

		dataMap[string(key)] = Data{Data: value, Len: index[i].ParamLength, MaxLen: index[i].ParamMaxLength, Format: index[i].ParamFormat}
	}

	return dataMap, nil
}
