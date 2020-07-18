package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/maxeaubrey/sforeader/sforeader"
)

func main() {
	file := flag.String("file", "param.sfo", "SFO file to parse")
	jsondump := flag.Bool("jsondump", false, "Dump all data as json")
	flag.Parse()

	dataMap, err := sforeader.ParseFile(*file)
	if err != nil {
		panic(err)
	}

	if *jsondump {
		jsonMap := make(map[string]string)
		for k, v := range dataMap {
			if v.Format == sforeader.FormatInteger {
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
