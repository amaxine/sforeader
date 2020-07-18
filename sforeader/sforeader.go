package sforeader

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Constants defining data types
const FormatUtf8SM = 4
const FormatUtf8 = 516
const FormatInteger = 1028

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
	ParamFormat     uint16 // see consts
	ParamLength     uint32
	ParamMaxLength  uint32
	DataTableOffset uint32
}

// Data structure of individual entries
type Data struct {
	Data   []byte
	Format uint16
}

// Basic validation. Expanding scope beyond Vita would probably require changes here.
// This only verifies the header is valid and doesn't go beyond it.
func (h *header) validate() error {
	if h.Magic != 1179865088 {
		return fmt.Errorf("PSF header invalid")
	}
	if h.Version != 257 {
		return fmt.Errorf("version invalid")
	}
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, h.KeyTableOffset)
	if bytes.Compare(tmp[2:4], []byte{0, 0}) != 0 {
		return fmt.Errorf("key table offset invalid")
	}
	binary.LittleEndian.PutUint32(tmp, h.DataTableOffset)
	if bytes.Compare(tmp[2:4], []byte{0, 0}) != 0 {
		return fmt.Errorf("data table offset invalid")
	}
	binary.LittleEndian.PutUint32(tmp, h.IndexTableEntries)
	if bytes.Compare(tmp[1:4], []byte{0, 0, 0}) != 0 {
		return fmt.Errorf("index table entries invalid")
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

		value := make([]byte, int(index[i].ParamMaxLength))
		err = binary.Read(f, binary.LittleEndian, value)
		if err != nil {
			return nil, err
		}
		// We read until ParamMaxLength, but we still want to trim the emptiness.
		// Would reading until ParamLength make more sense? Maybe, but just being safe.
		value = bytes.Trim(value, "\x00")
		dataMap[string(key)] = Data{Data: value, Format: index[i].ParamFormat}
	}

	return dataMap, nil
}
