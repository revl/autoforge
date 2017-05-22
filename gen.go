// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
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

	if info.IsDir() {
		return nil
	}

	t, err := template.ParseFiles(path)

	if err != nil {
		return err
	}

	type tmp struct {
		PackageName, PackageDescription, Copyright, License string
	}

	data := tmp{"Test", "Description", "Copyright", "License"}

	if err = t.Execute(os.Stdout, data); err != nil {
		return err
	}

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
