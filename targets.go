// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/doc"
	"os"
	"path"
	"strings"
)

type target struct {
	Target       string
	Phony        bool
	Dependencies []string
	MakeScript   string
}

type targetType interface {
	name() string
	help() string
	targets() ([]target, error)
}

type helpTarget struct {
	getTargetTypes func() []targetType
}

func createHelpTargetType(getTargetTypes func() []targetType) targetType {
	return &helpTarget{getTargetTypes}
}

func (*helpTarget) name() string {
	return "help"
}

func (*helpTarget) help() string {
	return "Display this help message. Unless overridden by the '--" +
		maketargetOption + "' option, this is the default target."
}

func (ht *helpTarget) targets() ([]target, error) {
	script := `	@echo "Usage:"
	@echo "    make [target...]"
	@echo
	@echo "Global targets:"
`

	for _, t := range ht.getTargetTypes() {
		helpText := t.help()

		if helpText == "" {
			continue
		}

		script += "\t@echo \"    " + t.name() + "\"\n"

		var buffer bytes.Buffer

		doc.ToText(&buffer, helpText, "", "    ", 52)

		help := buffer.String()

		for _, l := range strings.Split(help, "\n") {
			if l != "" {
				script += "\t@echo \"        " + l + "\"\n"
			} else {
				script += "\t@echo\n"
			}
		}
	}

	return []target{{
		Target:     "help",
		Phony:      true,
		MakeScript: script}}, nil
}

type makeTargetData struct {
	targetName string
	selection  packageDefinitionList
	ws         *workspace
}

func (mtd *makeTargetData) name() string {
	return mtd.targetName
}

func (mtd *makeTargetData) globalTarget() target {
	prefix := mtd.targetName + "_"

	var dependencies []string

	for _, pd := range mtd.selection {
		dependencies = append(dependencies, prefix+pd.PackageName)
	}

	return target{
		Target:       mtd.targetName,
		Phony:        true,
		Dependencies: dependencies}
}

func selfPathnameRelativeToWorkspace(ws *workspace) string {
	executable, err := os.Executable()
	if err != nil {
		return appName
	}

	return ws.relativeToWorkspace(executable)
}

type bootstrapTarget struct {
	makeTargetData
}

func createBootstrapTargetType(selection packageDefinitionList,
	ws *workspace) targetType {
	return &bootstrapTarget{makeTargetData{"bootstrap", selection, ws}}
}

func (*bootstrapTarget) help() string {
	return ""
}

func (bt *bootstrapTarget) targets() ([]target, error) {
	var bootstrapTargets []target

	pkgRootDir := bt.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(bt.ws) + " bootstrap "

	for _, pd := range bt.selection {
		configurePathname := path.Join(pkgRootDir,
			pd.PackageName, "configure")
		dependencies := []string{configurePathname + ".ac"}

		bootstrapTargets = append(bootstrapTargets, target{
			Target:       configurePathname,
			Dependencies: dependencies,
			MakeScript:   cmd + pd.PackageName + "\n"})
	}

	return bootstrapTargets, nil
}

type configureTarget struct {
	makeTargetData
	selectedDeps map[*packageDefinition]packageDefinitionList
}

func createConfigureTargetType(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList) targetType {
	return &configureTarget{
		makeTargetData{"configure", selection, ws},
		selectedDeps}
}

func (*configureTarget) help() string {
	return ""
}

func (ct *configureTarget) targets() ([]target, error) {
	var configureTargets []target

	relativeConftabPathname := path.Join(privateDirName, conftabFilename)

	relBuildDir := ct.ws.buildDirRelativeToWorkspace()
	pkgRootDir := ct.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(ct.ws) + " configure "

	for _, pd := range ct.selection {
		dependencies := []string{relativeConftabPathname,
			path.Join(pkgRootDir, pd.PackageName, "configure")}

		for _, dep := range ct.selectedDeps[pd] {
			dependencies = append(dependencies, path.Join(
				relBuildDir, dep.PackageName, "Makefile"))
		}

		configureTargets = append(configureTargets, target{
			Target: path.Join(relBuildDir,
				pd.PackageName, "Makefile"),
			Dependencies: dependencies,
			MakeScript:   cmd + pd.PackageName + "\n"})
	}

	return configureTargets, nil
}

type buildTarget struct {
	makeTargetData
	selectedDeps, dependentOnSelected map[*packageDefinition]packageDefinitionList
}

func createBuildTargetType(selection packageDefinitionList, ws *workspace,
	selectedDeps,
	dependentOnSelected map[*packageDefinition]packageDefinitionList) targetType {
	return &buildTarget{makeTargetData{"build", selection, ws},
		selectedDeps, dependentOnSelected}
}

func (*buildTarget) name() string {
	return "build"
}

func (*buildTarget) help() string {
	return "Build (compile and link) the selected packages. " +
		"For the packages that have not been configured, the " +
		"configuration step will be performed automatically."
}

func (bt *buildTarget) targets() ([]target, error) {
	var dependencies []string

	for _, pd := range bt.selection {
		if len(bt.dependentOnSelected[pd]) == 0 {
			dependencies = append(dependencies, pd.PackageName)
		}
	}

	globalTarget := target{
		Target:       bt.targetName,
		Phony:        true,
		Dependencies: dependencies}

	buildTargets := []target{globalTarget}

	relBuildDir := bt.ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[build] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make.log && \
	date >> make.log && \
	echo '--------------------------------' >> make.log && \
	$(MAKE) >> make.log
`
	for _, pd := range bt.selection {
		dependencies = []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range bt.selectedDeps[pd] {
			dependencies = append(dependencies, dep.PackageName)
		}

		buildTargets = append(buildTargets, target{
			Target:       pd.PackageName,
			Phony:        true,
			Dependencies: dependencies,
			MakeScript: fmt.Sprintf(scriptTemplate,
				pd.PackageName)})
	}

	return buildTargets, nil
}

type checkTarget struct {
}

func createCheckTargetType() targetType {
	return &checkTarget{}
}

func (*checkTarget) name() string {
	return "check"
}

func (*checkTarget) help() string {
	return "Build and run unit tests for the selected packages."
}

func (*checkTarget) targets() ([]target, error) {
	return nil, nil
}
