package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

const formatUtf8SM = 4
const formatUtf8 = 516
const formatInteger = 1028

type header struct {
	Signature         int32
	Version           int32
	KeyTableOffset    uint32
	DataTableOffset   uint32
	IndexTableEntries uint32
}

// Basic validation. Expanding scope beyond Vita would probably require changes here.
func (h header) validate() error {
	if h.Signature != 1179865088 {
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

type indexTable struct {
	KeyTableOffset  uint16
	ParamFormat     uint16
	ParamLength     uint32
	ParamMaxLength  uint32
	DataTableOffset uint32
}

type data struct {
	Data   []byte
	Format uint16
}

func parseFile(filename string) (map[string]data, error) {
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

	dataMap := make(map[string]data)
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
		dataMap[string(key)] = data{Data: value, Format: index[i].ParamFormat}
	}

	return dataMap, nil
}

func main() {
	file := flag.String("file", "param.sfo", "SFO file to parse")
	jsondump := flag.Bool("jsondump", false, "Dump all data as json")
	flag.Parse()

	dataMap, err := parseFile(*file)
	if err != nil {
		panic(err)
	}

	if *jsondump {
		jsonMap := make(map[string]string)
		for k, v := range dataMap {
			if v.Format == formatInteger {
				jsonMap[k] = fmt.Sprintf("%v", v.Data)
			} else {
				jsonMap[k] = string(v.Data)
			}
		}

		json, err := json.MarshalIndent(jsonMap, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(json))
	}
}
