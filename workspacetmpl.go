// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var filenameForSelectedPackages = "selected"

var conftabFilename = "conftab"

var workspaceTemplate = []embeddedTemplateFile{
	{privateDirName + "/" + filenameForSelectedPackages, 0644,
		[]byte(`{{range .selection}}{{.PackageName}}
{{end}}`)},
	{privateDirName + "/" + conftabFilename, 0644,
		[]byte(`{{.conftab.GlobalSection.Definition -}}
{{range .conftab.PackageSections}}[{{.PkgName}}]
{{.Definition -}}{{end}}`)},
	{"{makefile}", 0644,
		[]byte(`.PHONY: default all

default: {{.default_target}}

all: build

{{range .targets}}{{if .Phony}}.PHONY: {{.Target}}

{{end}}{{.Target}}:{{range .Dependencies}} \
	{{.}}{{end}}
{{.MakeScript}}
{{end}}`)},
}

func generateWorkspaceFiles(workspaceDir string,
	selection packageDefinitionList, conftab *Conftab) error {
	var targetTypes []targetType

	targetTypes = []targetType{
		createHelpTarget(func() []targetType { return targetTypes }),
		createBootstrapTarget(selection),
		createConfigureTarget(selection),
		createBuildTarget(),
		createCheckTarget(),
	}

	var targets []target

	for _, gt := range targetTypes {
		targets = append(targets, gt.targets()...)
	}

	params := templateParams{
		"makefile":       flags.makefile,
		"default_target": flags.defaultMakeTarget,
		"selection":      selection,
		"conftab":        conftab,
		"targets":        targets,
	}

	for _, templateFile := range workspaceTemplate {
		outputFiles, err := parseAndExecuteTemplate(
			templateFile.pathname, templateFile.contents,
			nil, nil, params)
		if err != nil {
			return err
		}
		_, err = writeGeneratedFiles(workspaceDir, outputFiles,
			templateFile.mode)
		if err != nil {
			return err
		}
	}
	return nil
}
