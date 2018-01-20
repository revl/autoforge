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
		[]byte(`.PHONY: default all{{range .globalTargets -}}
{{if .IsPhony}} {{.Name}}{{end}}{{end}}

default: {{.default_target}}

all: build

{{range .globalTargets}}{{.Name}}:
{{.Script}}
{{end}}`)},
}

func generateWorkspaceFiles(workspaceDir string,
	selection packageDefinitionList, conftab *Conftab) error {
	var globalTargets []target

	globalTargets = []target{
		createHelpTarget(func() []target { return globalTargets }),
		createBootstrapTarget(),
		createConfigureTarget(),
		createBuildTarget(),
		createCheckTarget(),
	}

	params := templateParams{
		"makefile":       flags.makefile,
		"default_target": flags.defaultMakeTarget,
		"selection":      selection,
		"conftab":        conftab,
		"globalTargets":  globalTargets,
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
