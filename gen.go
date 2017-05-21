// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const tmpl = `Hello {{.Name}}!
`

func printFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func generatePackageSources() {
	type PackageInfo struct {
		Name string
	}
	var packages = []PackageInfo{{"World"}}

	t := template.Must(template.New("package").Parse(tmpl))

	// Execute the template for each package.
	for _, p := range packages {
		err := t.Execute(os.Stdout, p)
		if err != nil {
			panic(err)
		}
	}

	err := filepath.Walk("templates/application", printFile)

	if err != nil {
		panic(err)
	}
}
