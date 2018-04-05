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

func generateWorkspaceFiles(ws *workspace, pi *packageIndex,
	selection packageDefinitionList, conftab *Conftab) error {

	selectedDeps := establishDependenciesInSelection(selection, pi)

	dependentOnSelected := map[*packageDefinition]packageDefinitionList{}
	for pd, deps := range selectedDeps {
		for _, dep := range deps {
			dependentOnSelected[dep] = append(
				dependentOnSelected[dep], pd)
		}
	}
	var globalTargetDeps []string
	for _, pd := range selection {
		if len(dependentOnSelected[pd]) == 0 {
			globalTargetDeps = append(globalTargetDeps,
				pd.PackageName)
		}
	}

	targets := []target{createHelpTarget()}

	targets = append(targets, createBootstrapTargets(selection, ws)...)
	targets = append(targets, createConfigureTargets(selection, ws,
		selectedDeps)...)
	targets = append(targets, createBuildTargets(selection, ws,
		selectedDeps, globalTargetDeps)...)
	targets = append(targets, createCheckTargets(selection, ws,
		selectedDeps)...)
	targets = append(targets, createInstallTargets(selection, ws,
		selectedDeps, globalTargetDeps)...)
	targets = append(targets, createDistTargets(selection, ws,
		selectedDeps)...)

	makefile := ws.wp.Makefile
	if flags.makefile != "" {
		makefile = flags.makefile
	} else if makefile == "" {
		makefile = "Makefile"
	}

	defaultTarget := ws.wp.DefaultMakeTarget
	if flags.defaultMakeTarget != "" {
		defaultTarget = flags.defaultMakeTarget
	} else if defaultTarget == "" {
		defaultTarget = "help"
	}

	params := templateParams{
		"makefile":       makefile,
		"default_target": defaultTarget,
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
		_, err = writeGeneratedFiles(ws.absDir, outputFiles,
			templateFile.mode)
		if err != nil {
			return err
		}
	}
	return nil
}
