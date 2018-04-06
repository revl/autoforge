// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path"
)

type target struct {
	Target       string
	Phony        bool
	Dependencies []string
	MakeScript   string
}

type makefileTargetCreator struct {
	ws               *workspace
	selection        packageDefinitionList
	selectedDeps     map[*packageDefinition]packageDefinitionList
	globalTargetDeps []string
}

func newMakefileTargetCreator(ws *workspace,
	selection packageDefinitionList,
	pi *packageIndex) *makefileTargetCreator {

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

	return &makefileTargetCreator{ws, selection,
		selectedDeps, globalTargetDeps}
}

func (mtc *makefileTargetCreator) createHelpTarget() target {
	return target{
		Target: "help",
		Phony:  true,
		MakeScript: `	@echo "Usage:"
	@echo "    make [target...]"
	@echo
	@echo "Global targets:"
	@echo "    help"
	@echo "        Display this help message. Unless overridden by the"
	@echo "        '--` + maketargetOption +
			`' option, this is the default target."
	@echo
	@echo "    bootstrap"
	@echo "        Create (or update) the 'configure' scripts for"
	@echo "        all selected packages."
	@echo
	@echo "    configure"
	@echo "        Configure the selected packages using the current"
	@echo "        conftab and generate makefiles for building them."
	@echo "        To change configuration options, run"
	@echo
	@echo "            ` + appName + " " + conftabCmdName + `"
	@echo
	@echo "    build"
	@echo "        Build (compile and link) the selected packages. For"
	@echo "        the packages that have not been configured, the"
	@echo "        configuration step will be performed automatically."
	@echo
	@echo "    check"
	@echo "        Build and run unit tests for the selected packages."
	@echo
	@echo "    install"
	@echo "        Install package binaries and library headers into"
	@echo "        '` + mtc.ws.installDir() + `'."
	@echo
	@echo "    dist"
	@echo "        Create distribution tarballs and move them to the"
	@echo "        'dist' subdirectory of the workspace."
	@echo
`}
}

func selfPathnameRelativeToWorkspace(ws *workspace) string {
	executable, err := os.Executable()
	if err != nil {
		return appName
	}

	return ws.relativeToWorkspace(executable)
}

func (mtc *makefileTargetCreator) createBootstrapTargets() []target {
	var targets []target

	pkgRootDir := mtc.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(mtc.ws) + " bootstrap "

	for _, pd := range mtc.selection {
		configurePathname := path.Join(pkgRootDir,
			pd.PackageName, "configure")
		dependencies := []string{configurePathname + ".ac"}

		targets = append(targets, target{
			Target:       configurePathname,
			Dependencies: dependencies,
			MakeScript:   cmd + pd.PackageName + "\n"})
	}

	return targets
}

func (mtc *makefileTargetCreator) createConfigureTargets() []target {
	var targets []target

	relativeConftabPathname := path.Join(privateDirName, conftabFilename)

	relBuildDir := mtc.ws.buildDirRelativeToWorkspace()
	pkgRootDir := mtc.ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(mtc.ws) + " configure "

	for _, pd := range mtc.selection {
		dependencies := []string{relativeConftabPathname,
			path.Join(pkgRootDir, pd.PackageName, "configure")}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, path.Join(
				relBuildDir, dep.PackageName, "Makefile"))
		}

		targets = append(targets, target{
			Target: path.Join(relBuildDir,
				pd.PackageName, "Makefile"),
			Dependencies: dependencies,
			MakeScript:   cmd + pd.PackageName + "\n"})
	}

	return targets
}

func (mtc *makefileTargetCreator) createBuildTargets() []target {

	targets := []target{target{"build", true, mtc.globalTargetDeps, ""}}

	relBuildDir := mtc.ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[build] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make.log && \
	date >> make.log && \
	echo '--------------------------------' >> make.log && \
	$(MAKE) >> make.log
`

	for _, pd := range mtc.selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, dep.PackageName)
		}

		targets = append(targets, target{
			Target:       pd.PackageName,
			Phony:        true,
			Dependencies: dependencies,
			MakeScript: fmt.Sprintf(scriptTemplate,
				pd.PackageName)})
	}

	return targets
}

func (mtc *makefileTargetCreator) createCheckTargets() []target {

	var selectedPkgNames []string

	for _, pd := range mtc.selection {
		selectedPkgNames = append(selectedPkgNames,
			"check_"+pd.PackageName)
	}

	targets := []target{target{"check", true, selectedPkgNames, ""}}

	relBuildDir := mtc.ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[check] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_check.log && \
	date >> make_check.log && \
	echo '--------------------------------' >> make_check.log && \
	$(MAKE) check >> make_check.log
`

	for _, pd := range mtc.selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, dep.PackageName)
		}

		targets = append(targets, target{
			Target:       "check_" + pd.PackageName,
			Phony:        true,
			Dependencies: dependencies,
			MakeScript: fmt.Sprintf(scriptTemplate,
				pd.PackageName)})
	}

	return targets
}

func (mtc *makefileTargetCreator) createInstallTargets() []target {

	var selectedPkgNames []string

	for _, dep := range mtc.globalTargetDeps {
		selectedPkgNames = append(selectedPkgNames,
			"install_"+dep)
	}

	targets := []target{target{"install", true, selectedPkgNames, ""}}

	relBuildDir := mtc.ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[install] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_install.log && \
	date >> make_install.log && \
	echo '--------------------------------' >> make_install.log && \
	$(MAKE) install >> make_install.log
`

	for _, pd := range mtc.selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies,
				"install_"+dep.PackageName)
		}

		targets = append(targets, target{
			Target:       "install_" + pd.PackageName,
			Phony:        true,
			Dependencies: dependencies,
			MakeScript: fmt.Sprintf(scriptTemplate,
				pd.PackageName)})
	}

	return targets
}

func (mtc *makefileTargetCreator) createDistTargets() []target {
	var selectedPkgNames []string

	for _, pd := range mtc.selection {
		selectedPkgNames = append(selectedPkgNames,
			"dist_"+pd.PackageName)
	}

	targets := []target{target{"dist", true, selectedPkgNames, ""}}

	relBuildDir := mtc.ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[dist] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_dist.log && \
	date >> make_dist.log && \
	echo '--------------------------------' >> make_dist.log && \
	$(MAKE) dist >> make_dist.log
	@mkdir -p dist
	@mv '` + relBuildDir + `/%[1]s/%[1]s-%[2]s.tar.gz' dist/
`

	for _, pd := range mtc.selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, path.Join(
				relBuildDir, dep.PackageName, "Makefile"))
		}

		targets = append(targets, target{
			Target:       "dist_" + pd.PackageName,
			Phony:        true,
			Dependencies: dependencies,
			MakeScript: fmt.Sprintf(scriptTemplate,
				pd.PackageName,
				pd.params["version"])})
	}

	return targets
}
