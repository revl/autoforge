// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
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

func createHelpTarget(getTargetTypes func() []targetType) targetType {
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
		script += "\t@echo \"    " + t.name() + "\"\n"

		var buffer bytes.Buffer

		doc.ToText(&buffer, t.help(), "", "    ", 52)

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

func createBootstrapTarget(selection packageDefinitionList,
	ws *workspace) targetType {
	return &bootstrapTarget{makeTargetData{"bootstrap", selection, ws}}
}

func (*bootstrapTarget) help() string {
	return "Unconditionally regenerate the 'configure' " +
		"scripts for the selected packages."
}

func (bt *bootstrapTarget) targets() ([]target, error) {
	globalTarget := bt.globalTarget()

	bootstrapTargets := []target{globalTarget}

	pkgRootDir := bt.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(bt.ws) + " bootstrap "

	for i, pd := range bt.selection {
		script := cmd + pd.PackageName + "\n"

		configurePathname := path.Join(pkgRootDir,
			pd.PackageName, "configure")
		dependencies := []string{configurePathname + ".ac"}

		bootstrapTargets = append(bootstrapTargets,
			target{
				Target:       globalTarget.Dependencies[i],
				Phony:        true,
				Dependencies: dependencies,
				MakeScript:   script},
			target{
				Target:       configurePathname,
				Dependencies: dependencies,
				MakeScript:   script})
	}

	return bootstrapTargets, nil
}

type configureTarget struct {
	makeTargetData
	selectedDeps map[*packageDefinition]packageDefinitionList
}

func createConfigureTarget(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList) targetType {
	return &configureTarget{
		makeTargetData{"configure", selection, ws},
		selectedDeps}
}

func (*configureTarget) help() string {
	return "Configure the selected packages using the " +
		"current options specified in the 'conftab' file."
}

func (ct *configureTarget) targets() ([]target, error) {
	globalTarget := ct.globalTarget()

	configureTargets := []target{globalTarget}

	relBuildDir := ct.ws.buildDirRelativeToWorkspace()
	pkgRootDir := ct.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(ct.ws) + " configure "

	for i, pd := range ct.selection {
		dependencies := []string{path.Join(pkgRootDir,
			pd.PackageName, "configure")}

		for _, dep := range ct.selectedDeps[pd] {
			dependencies = append(dependencies, path.Join(
				relBuildDir, dep.PackageName, "Makefile"))
		}

		script := cmd + pd.PackageName + "\n"

		configureTargets = append(configureTargets,
			target{
				Target:       globalTarget.Dependencies[i],
				Phony:        true,
				Dependencies: dependencies,
				MakeScript:   script},
			target{
				Target: path.Join(relBuildDir,
					pd.PackageName, "Makefile"),
				Dependencies: dependencies,
				MakeScript:   script})
	}

	return configureTargets, nil
}

type buildTarget struct {
}

func createBuildTarget() targetType {
	return &buildTarget{}
}

func (*buildTarget) name() string {
	return "build"
}

func (*buildTarget) help() string {
	return "Build (compile and link) the selected packages. " +
		"For the packages that have not been configured, the " +
		"configuration step will be performed automatically."
}

func (*buildTarget) targets() ([]target, error) {
	return nil, nil
}

type checkTarget struct {
}

func createCheckTarget() targetType {
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
