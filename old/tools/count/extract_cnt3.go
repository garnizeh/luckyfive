package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type SimulationResult struct {
	Summary map[string]float64 `json:"summary"`
}

func main() {
	dir := flag.String("dir", "data/simulations", "Diretório com arquivos de simulação (.json)")
	flag.Parse()

	var items []struct {
		path string
		mod  int64
	}
	_ = filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			items = append(items, struct {
				path string
				mod  int64
			}{path: path, mod: info.ModTime().UnixNano()})
		}
		return nil
	})

	if len(items) == 0 {
		fmt.Println(0)
		return
	}

	sort.Slice(items, func(i, j int) bool { return items[i].mod > items[j].mod })
	newest := items[0].path

	b, err := os.ReadFile(newest)
	if err != nil {
		fmt.Fprintln(os.Stderr, 0)
		return
	}
	var res SimulationResult
	if err := json.Unmarshal(b, &res); err != nil {
		fmt.Fprintln(os.Stderr, 0)
		return
	}
	if res.Summary == nil {
		fmt.Println(0)
		return
	}
	v := int(res.Summary["cnt_3_plus"])
	fmt.Printf("%d", v)
}
