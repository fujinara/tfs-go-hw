package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)

type finance struct {
	Company    string
	Type       string
	Value      int
	Id         interface{}
	Created_at time.Time
	Valid      bool
	Omit       bool
}

type outformat struct {
	Company     string        `json:"company"`
	ValidOpsCnt int           `json:"valid_operations_count"`
	Balance     int           `json:"balance"`
	InvalidOps  []interface{} `json:"invalid_operations,omitempty"`
}

func getPath() string {
	var FilePathFlag = flag.String("file", "", "path to data file")
	flag.Parse()

	if *FilePathFlag == "" {
		var ev = "FILE"
		FilePathEnv, ok := os.LookupEnv(ev)
		if ok {
			return FilePathEnv
		}
	}
	return *FilePathFlag
}

func (d *finance) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	err := json.Unmarshal(data, &v)

	if err != nil {
		return err
	}

	d.Omit = false

	if v["company"] != nil {
		d.Company, _ = v["company"].(string)
	}

	var ops map[string]interface{}
	var kk bool

	if v["operation"] != nil {
		ops, kk = v["operation"].(map[string]interface{})
		if !kk {
			return errors.New("operation field cannot be parsed as map[string]interface{}")
		}
	}

	if v["type"] != nil {
		if typee, ok := v["type"].(string); ok {
			d.Type = typee
		}
	} else if v["operation"] != nil && ops["type"] != nil {
		if typee, ok := ops["type"].(string); ok {
			d.Type = typee
		}
	}

	if v["id"] != nil {
		d.Id = v["id"]
	} else if v["operation"] != nil && ops["id"] != nil {
		d.Id = ops["id"]
	}

	layout := "2006-01-02T15:04:05Z07:00"

	var createdErr error

	if v["created_at"] != nil {
		d.Created_at, createdErr = time.Parse(layout, v["created_at"].(string))
		if createdErr != nil {
			d.Omit = true
		}
	} else if v["operation"] != nil && ops["created_at"] != nil {
		d.Created_at, createdErr = time.Parse(layout, ops["created_at"].(string))
		if createdErr != nil {
			d.Omit = true
		}
	}

	if v["value"] != nil {
		if fval, ok := v["value"].(float64); ok {
			if fval == float64(int64(fval)) {
				d.Value = int(fval)
			}
		}
		if sval, ok := v["value"].(string); ok {
			s, err := strconv.ParseFloat(sval, 64)
			if err == nil {
				if s == float64(int64(s)) {
					d.Value = int(s)
				}
			}
		}

	} else if v["operation"] != nil && ops["value"] != nil {
		fval, ok := ops["value"].(float64)
		if ok {
			d.Value = int(fval)
		}
		sval, ok := ops["value"].(string)
		if ok {
			s, err := strconv.ParseFloat(sval, 64)
			if err == nil {
				d.Value = int(s)
			}
		}
	}

	d.Valid = false

	if d.Value != 0 && (d.Type == "income" || d.Type == "outcome" || d.Type == "+" || d.Type == "-") {
		d.Valid = true
	}

	_, oks := d.Id.(string)
	flv, oki := d.Id.(float64)

	if d.Company == "" || (!oks && (!oki || flv != float64(int64(flv)))) || d.Created_at.IsZero() {
		d.Omit = true
	}

	return nil
}

func main() {
	path := getPath()
	var data []byte
	var fileErr error
	if path != "" {
		data, fileErr = ioutil.ReadFile(path)
		if fileErr != nil {
			fmt.Println(fileErr)
		}
	} else {
		data, fileErr = io.ReadAll(os.Stdin)
		if fileErr != nil {
			fmt.Println(fileErr)
		}
	}

	var finList []finance
	unmarshErr := json.Unmarshal(data, &finList)
	if unmarshErr != nil {
		fmt.Println(unmarshErr)
	}

	sort.Slice(finList, func(i, j int) bool {
		return finList[i].Created_at.Before(finList[j].Created_at)
	})

	var outmap = map[string]*outformat{}

	for _, op := range finList {
		if !op.Omit {
			outmap[op.Company] = &outformat{Company: op.Company, InvalidOps: make([]interface{}, 0)}
		}
	}

	for _, op := range finList {
		if !op.Omit {
			if op.Valid {
				outmap[op.Company].ValidOpsCnt += 1
				if op.Type == "income" || op.Type == "+" {
					outmap[op.Company].Balance += op.Value
				} else {
					outmap[op.Company].Balance -= op.Value
				}
			} else {
				outmap[op.Company].InvalidOps = append(outmap[op.Company].InvalidOps, op.Id)
			}
		}
	}

	var outlist = []outformat{}
	for comp := range outmap {
		outlist = append(outlist, *outmap[comp])
	}

	sort.Slice(outlist, func(i, j int) bool {
		return outlist[i].Company < outlist[j].Company
	})

	f, crtErr := os.Create("out.json")
	if crtErr != nil {
		fmt.Println(crtErr)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	encErr := enc.Encode(outlist)
	if encErr != nil {
		fmt.Println(encErr)
	}
	f.Close()
}
