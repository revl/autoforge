// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type selectedPackages []string

var filenameForSelectedPackages = "selected"

func readPackageSelection(privateDir string) (selectedPackages, error) {
	file, err := os.Open(filepath.Join(privateDir,
		filenameForSelectedPackages))
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var selected selectedPackages

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		selected = append(selected, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return selected, nil
}

func getRequired(pd *packageDefinition) packageDefinitionList {
	return pd.required
}

func getDependent(pd *packageDefinition) packageDefinitionList {
	return pd.dependent
}

func applyToSubtree(action func(*packageDefinition),
	root *packageDefinition,
	direction func(*packageDefinition) packageDefinitionList) {

	queue := packageDefinitionList{root}

	for {
		pd := queue[0]
		queue = queue[1:]

		action(pd)

		queue = append(queue, direction(pd)...)

		if len(queue) == 0 {
			break
		}
	}
}

func selectPackages(pi *packageIndex, args []string) (packageDefinitionList,
	error) {
	selected := make(map[string]bool)
	marked := make(map[string]int)

	inclusion := true

	selectPackage := func(pd *packageDefinition) {
		selected[pd.PackageName] = inclusion
	}

	mark := 0

	markPackage := func(pd *packageDefinition) {
		marked[pd.PackageName] = mark
	}

	selectIfMarked := func(pd *packageDefinition) {
		if marked[pd.PackageName] == mark {
			selected[pd.PackageName] = inclusion
		}
	}

	for _, arg := range args {
		if arg == "+" {
			inclusion = true
			continue
		}

		if arg == "-" {
			inclusion = false
			continue
		}

		var pkgRange packageDefinitionList

		emptyRange := true

		for _, pkgName := range strings.SplitN(arg, ":", 2) {
			var pd *packageDefinition
			if pkgName != "" {
				pd = pi.packageByName[pkgName]
				if pd == nil {
					return nil, errors.New(
						"no such package: " + pkgName)
				}
				emptyRange = false
			}
			pkgRange = append(pkgRange, pd)
		}

		if emptyRange {
			continue
		}

		if len(pkgRange) == 1 {
			selected[arg] = inclusion
			continue
		}

		from, to := pkgRange[0], pkgRange[1]

		if from == nil {
			applyToSubtree(selectPackage, to, getRequired)
		} else if to == nil {
			applyToSubtree(selectPackage, from, getDependent)
		} else {
			mark++

			applyToSubtree(markPackage, to, getRequired)

			applyToSubtree(selectIfMarked, from, getDependent)
		}
	}

	var pkgSelection packageDefinitionList

	for _, pd := range pi.orderedPackages {
		if selected[pd.PackageName] {
			pkgSelection = append(pkgSelection, pd)
		}
	}

	return pkgSelection, nil
}
