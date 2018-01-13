// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"os"
)

type packageSelection []string

var packageSelectionFilename = "selected"

func readPackageSelection(pathname string) (packageSelection, error) {
	file, err := os.Open(pathname)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var selection packageSelection

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		selection = append(selection, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return selection, nil
}