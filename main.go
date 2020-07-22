package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/maxeaubrey/sforeader/sforeader"
)

func main() {
	file := flag.String("file", "param.sfo", "SFO file to parse")
	jsondump := flag.Bool("jsondump", true, "Dump all data as json")
	flag.Parse()

	dataMap, err := sforeader.ParseFile(*file)
	if err != nil {
		panic(err)
	}

	if *jsondump {
		jsonMap := make(map[string]string)
		for k, v := range dataMap {
			if v.Format == sforeader.FormatInteger {
				jsonMap[k] = fmt.Sprintf("%v", int(binary.LittleEndian.Uint32(v.Data)))
			} else if v.Format == sforeader.FormatUtf8 {
				jsonMap[k] = string(bytes.Trim(v.Data, "\x00"))
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
