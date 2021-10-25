package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)

type OpType struct {
	Type       interface{} `json:"type,omitempty"`
	Value      interface{} `json:"value,omitempty"`
	Id         interface{} `json:"id,omitempty"`
	Created_at interface{} `json:"created_at,omitempty"`
}

type inpformat struct {
	Company   string `json:"company,omitempty"`
	Operation OpType `json:"operation,omitempty"`
	OpType
}

type outformat struct {
	Company     string        `json:"company"`
	ValidOpsCnt int           `json:"valid_operations_count"`
	Balance     int           `json:"balance"`
	InvalidOps  []interface{} `json:"invalid_operations,omitempty"`
}

const layout = "2006-01-02T15:04:05Z07:00"

func (tr *inpformat) IsCorrect() bool {
	_, oks := tr.Id.(string)
	floatVal, oki := tr.Id.(float64)
	timeStr, oktime := tr.Created_at.(string)
	_, err := time.Parse(layout, timeStr)
	if tr.Company == "" || tr.Id == nil || (!oks && (!oki || floatVal != float64(int64(floatVal)))) || err != nil || !oktime {
		return false
	}
	return true
}

func (tr *inpformat) IsValid() bool {
	verdict := false
	if tr.Value != nil {
		if fval, ok := tr.Value.(float64); ok {
			if fval == float64(int64(fval)) {
				verdict = true
			}
		}
		if sval, ok := tr.Value.(string); ok {
			s, err := strconv.ParseFloat(sval, 64)
			if err == nil {
				if s == float64(int64(s)) {
					verdict = true
				}
			}
		}
	}
	if tr.Type != nil {
		typeStr, ok := tr.Type.(string)
		if verdict && ok && (typeStr == "income" || typeStr == "outcome" || typeStr == "+" || typeStr == "-") {
			verdict = true
		}
	}
	return verdict
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

	var finList []inpformat
	unmarshErr := json.Unmarshal(data, &finList)
	if unmarshErr != nil {
		fmt.Println(unmarshErr)
	}

	for i := range finList {
		if finList[i].Operation != (OpType{}) {
			if finList[i].Operation.Type != nil {
				finList[i].Type = finList[i].Operation.Type
			}
			if finList[i].Operation.Value != nil {
				finList[i].Value = finList[i].Operation.Value
			}
			if finList[i].Operation.Id != nil {
				finList[i].Id = finList[i].Operation.Id
			}
			if finList[i].Operation.Created_at != nil {
				finList[i].Created_at = finList[i].Operation.Created_at
			}
		}
	}

	var outmap = map[string]*outformat{}
	var CorrectFinList = []inpformat{}

	for _, op := range finList {
		if op.IsCorrect() {
			outmap[op.Company] = &outformat{Company: op.Company, InvalidOps: make([]interface{}, 0)}
			CorrectFinList = append(CorrectFinList, op)
		}
	}

	sort.Slice(CorrectFinList, func(i, j int) bool {
		creatTimeI, _ := time.Parse(layout, CorrectFinList[i].Created_at.(string))
		creatTimeJ, _ := time.Parse(layout, CorrectFinList[j].Created_at.(string))
		return creatTimeI.Before(creatTimeJ)
	})

	for _, op := range CorrectFinList {
		if op.IsCorrect() {
			if op.IsValid() {
				outmap[op.Company].ValidOpsCnt += 1
				typeStr := fmt.Sprintf("%v", op.Type)
				var valueInt int
				if fval, ok := op.Value.(float64); ok {
					if fval == float64(int64(fval)) {
						valueInt = int(fval)
					}
				}
				if sval, ok := op.Value.(string); ok {
					s, err := strconv.ParseFloat(sval, 64)
					if err == nil {
						if s == float64(int64(s)) {
							valueInt = int(s)
						}
					}
				}
				if typeStr == "income" || typeStr == "+" {
					outmap[op.Company].Balance += valueInt
				} else {
					outmap[op.Company].Balance -= valueInt
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
