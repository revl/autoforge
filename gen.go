// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

func generatePackageSources() error {

	pd := packageDefinition{
		"Name":        "Test",
		"Description": "Description",
		"Type":        "application",
		"Copyright":   "Copyright",
		"Requires":    []string{"liba", "libb"},
		"License":     "License",
		"Sources":     []string{"source1.cc", "source2.cc"}}

	err := generateBuildFilesFromProjectTemplate(
		"templates/asdf/..//./application",
		"output", templateParams(pd))

	if err != nil {
		return err
	}

	return nil
}
