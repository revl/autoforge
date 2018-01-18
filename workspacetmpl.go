// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var filenameForSelectedPackages = "selected"

var conftabFilename = "conftab"

var workspaceTemplate = []embeddedTemplateFile{
	embeddedTemplateFile{privateDirName + "/" +
		filenameForSelectedPackages, 0644,
		[]byte(`{{range .selection}}{{.PackageName}}
{{end}}`)},
	embeddedTemplateFile{privateDirName + "/" +
		conftabFilename, 0644,
		[]byte(`{{.conftab.GlobalSection.Definition -}}
{{range .conftab.PackageSections}}[{{.PkgName}}]
{{.Definition -}}{{end}}`)},
	embeddedTemplateFile{"{makefile}", 0644,
		[]byte(`.PHONY: default all

default: {{.default_target}}

help:
	@echo "Usage:"
	@echo "    make [target...]"
	@echo
	@echo "Global targets:"
	@echo "    help"
	@echo "        Display this help message. Unless overridden by the"
	@echo "        '--` + maketargetOption + `' option, this ` +
			`is the default target."
	@echo
	@echo "    bootstrap"
	@echo "        Unconditionally regenerate the 'configure' scripts"
	@echo "        for the selected packages."
	@echo
	@echo "    configure"
	@echo "        Configure the selected packages using the current"
	@echo "        options specified in the 'conftab' file."
	@echo
	@echo "    build"
	@echo "        Build (compile and link) the selected packages."
	@echo "        For the packages that have not been configured, the"
	@echo "        configuration step will be performed automatically."
	@echo
	@echo "    check"
	@echo "        Build and run unit tests for the selected packages."
	@echo
	@echo "Individual package targets:"

all: build
{{range .targets}}{{if .IsPhony}}.PHONY: {{.Name}}

{{end}}{{.Name}}:
{{.Script}}
{{end}}`)},
}

func generateWorkspaceFiles(workspaceDir string,
	selection packageDefinitionList, conftab *Conftab) error {
	var globalTargets []target

	globalTargets = []target{
		createHelpTarget(func() []target { return globalTargets }),
	}

	params := templateParams{
		"makefile":       flags.makefile,
		"default_target": flags.defaultMakeTarget,
		"selection":      selection,
		"conftab":        conftab,
		"targets":        globalTargets,
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
