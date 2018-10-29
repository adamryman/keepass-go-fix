package main

import (
	"encoding/csv"
	//"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var (
	input1FileFlag *string = pflag.StringP("input1", "i", "", "first input file")
	input2FileFlag *string = pflag.StringP("input2", "j", "", "second input file")
	outFileFlag    *string = pflag.StringP("output", "o", "", "output file")
)

type KeePassEntry struct {
	Data struct {
		Group    string
		Title    string
		Username string
		Password string
		URL      string
		Notes    string
	}
	Row            []string
	ConflictKey    string
	GroupTitleUser string
}

var elements deduper = make(map[string]KeePassEntry)

func main() {
	pflag.Parse()
	input1FileName := *input1FileFlag
	input2FileName := *input2FileFlag
	outputFileName := *outFileFlag

	i1, err := os.Open(input1FileName)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot open input file 1"))
	}
	i2, err := os.Open(input2FileName)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot open input file 2"))
	}

	csv1 := csv.NewReader(i1)
	csv2 := csv.NewReader(i2)

	// Remove headers from csv
	header, err := csv1.Read()
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot read first line in csv1"))
	}

	// Remove headers from csv
	header, err = csv2.Read()
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot read first line in csv1"))
	}

	elements.dedupCSV(csv1)
	conflicts := elements.dedupCSV(csv2)

	out, err := os.Create(outputFileName)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot create output file"))
	}
	csvOut := csv.NewWriter(out)
	err = csvOut.Write(header)
	if err != nil {
		log.Fatal(errors.WithMessage(err, "cannot write headers"))
	}

	for _, val := range elements {
		err = csvOut.Write(val.Row)
		if err != nil {
			log.Fatal(errors.WithMessage(err, "cannot write value"))
		}
	}

	if len(conflicts) > 0 {
		log.Println("CONFLICTS DETECTED. CHECK END OF OUTPUT FILE TO RESOLVE")
		newline := make([]string, len(header), len(header))
		for i, _ := range newline {
			newline[i] = "CONFLICT"
		}
		for _, v := range conflicts {
			csvOut.Write(newline)
			csvOut.Write(v[0])
			csvOut.Write(v[1])
		}
	}
	csvOut.Flush()

	out.Close()
}

type deduper map[string]KeePassEntry

func (d deduper) dedupCSV(csv *csv.Reader) (conflicts [][][]string) {
	for row, err := csv.Read(); row != nil; row, err = csv.Read() {
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(errors.WithMessage(err, "cannot read line in csv1"))
		}
		// Join first [0] and second [1] elements for Group-Title
		k := KeePassEntry{
			Data: struct {
				Group    string
				Title    string
				Username string
				Password string
				URL      string
				Notes    string
			}{
				Group:    row[0],
				Title:    row[1],
				Username: row[2],
				Password: row[3],
				URL:      row[4],
				Notes:    row[5],
			},
			Row:            row,
			ConflictKey:    strings.Join(row, "k"),
			GroupTitleUser: row[0] + "/" + row[1] + "/" + row[2],
		}
		if val, ok := d[k.GroupTitleUser]; ok {
			if val.ConflictKey != k.ConflictKey {
				conflicts = append(conflicts, [][]string{k.Row, val.Row})
				k.PrintDiff(val)
			}
			continue
		}
		d[k.GroupTitleUser] = k
	}

	return conflicts
}

// PrintDiff
func (k KeePassEntry) PrintDiff(k2 KeePassEntry) {
	log.Println(k.GroupTitleUser)
	if k.Data.Password != k2.Data.Password {
		log.Println("Password")
		log.Println("--------------------")
		log.Println(k.Data.Password)
		log.Println(k2.Data.Password)
		log.Println("--------------------")
	}
	if k.Data.Notes != k2.Data.Notes {
		log.Println("Notes")
		log.Println("--------------------")
		log.Println(k.Data.Notes)
		log.Println(k2.Data.Notes)
		log.Println("--------------------")

	}
	if k.Data.URL != k2.Data.URL {
		log.Println("URL")
		log.Println("--------------------")
		log.Println(k.Data.URL)
		log.Println(k2.Data.URL)
		log.Println("--------------------")
	}
}
