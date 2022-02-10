package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/inputs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	repro_lang "github.com/sourcegraph/sourcegraph/lib/codeintel/repro_lang/src"
)

var update = flag.Bool("update", false, "update .golden files, removing unused if running all tests")

func TestLsifTyped(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	inputDirectory := filepath.Join(cwd, "snapshots-input")
	outputDirectory := filepath.Join(cwd, "snapshots-output")
	if *update {
		err = os.RemoveAll(outputDirectory)
		if err != nil {
			t.Fatal(err)
		}
	}
	testCases, err := os.ReadDir(inputDirectory)
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			t.Fatalf("not a directory: %v", testCase.Name())
		}
		t.Run(testCase.Name(), func(t *testing.T) {
			dir := filepath.Join(inputDirectory, testCase.Name())
			outputFile := filepath.Join(outputDirectory, testCase.Name(), "dump.lsif")
			var documents []*inputs.Document
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				absolutePath := filepath.Join(dir, entry.Name())
				relativePath := filepath.Join(testCase.Name(), entry.Name())
				data, err := os.ReadFile(absolutePath)
				if err != nil {
					t.Fatal(err)
				}
				documents = append(documents, inputs.NewDocument(relativePath, string(data)))
			}
			index, err := repro_lang.Index("file:/root", testCase.Name(), documents)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(index)
			lsif, err := reader.ConvertTypedIndexToGraphIndex(index)
			if err != nil {
				t.Fatal(err)
			}
			var obtained bytes.Buffer
			err = reader.WriteNDJSON(reader.ElementsToEmptyInterfaces(lsif), &obtained)
			if err != nil {
				t.Fatal(err)
			}
			if *update {
				err = os.MkdirAll(filepath.Dir(outputFile), 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(outputFile, obtained.Bytes(), 0755)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				expected, err := os.ReadFile(outputFile)
				if err != nil {
					expected = []byte{}
				}
				edits := myers.ComputeEdits(span.URIFromPath(outputFile), string(expected), obtained.String())
				if len(edits) > 0 {
					diff := fmt.Sprint(gotextdiff.ToUnified(
						outputFile+" (obtained)",
						outputFile+" (expected)",
						string(expected),
						edits,
					))
					t.Fatalf("\n" + diff)
				}
			}
		})
	}
}
