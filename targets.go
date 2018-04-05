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

func createHelpTarget() target {
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
	@echo "    build"
	@echo "        Build (compile and link) the selected packages. For"
	@echo "        the packages that have not been configured, the"
	@echo "        configuration step will be performed automatically."
	@echo
	@echo "    check"
	@echo "        Build and run unit tests for the selected packages."
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

func createBootstrapTargets(selection packageDefinitionList,
	ws *workspace) []target {
	var targets []target

	pkgRootDir := ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(ws) + " bootstrap "

	for _, pd := range selection {
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

func createConfigureTargets(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList) []target {
	var targets []target

	relativeConftabPathname := path.Join(privateDirName, conftabFilename)

	relBuildDir := ws.buildDirRelativeToWorkspace()
	pkgRootDir := ws.pkgRootDirRelativeToWorkspace()

	cmd := "\t@" + selfPathnameRelativeToWorkspace(ws) + " configure "

	for _, pd := range selection {
		dependencies := []string{relativeConftabPathname,
			path.Join(pkgRootDir, pd.PackageName, "configure")}

		for _, dep := range selectedDeps[pd] {
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

func createBuildTargets(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList,
	globalTargetDeps []string) []target {

	targets := []target{target{"build", true, globalTargetDeps, ""}}

	relBuildDir := ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[build] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make.log && \
	date >> make.log && \
	echo '--------------------------------' >> make.log && \
	$(MAKE) >> make.log
`

	for _, pd := range selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range selectedDeps[pd] {
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

func createCheckTargets(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList) []target {

	var selectedPkgNames []string

	for _, pd := range selection {
		selectedPkgNames = append(selectedPkgNames,
			"check_"+pd.PackageName)
	}

	targets := []target{target{"check", true, selectedPkgNames, ""}}

	relBuildDir := ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[check] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_check.log && \
	date >> make_check.log && \
	echo '--------------------------------' >> make_check.log && \
	$(MAKE) check >> make_check.log
`

	for _, pd := range selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range selectedDeps[pd] {
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

func createInstallTargets(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList,
	globalTargetDeps []string) []target {

	var selectedPkgNames []string

	for _, dep := range globalTargetDeps {
		selectedPkgNames = append(selectedPkgNames,
			"install_"+dep)
	}

	targets := []target{target{"install", true, selectedPkgNames, ""}}

	relBuildDir := ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[install] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_install.log && \
	date >> make_install.log && \
	echo '--------------------------------' >> make_install.log && \
	$(MAKE) install >> make_install.log
`

	for _, pd := range selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range selectedDeps[pd] {
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

func createDistTargets(selection packageDefinitionList, ws *workspace,
	selectedDeps map[*packageDefinition]packageDefinitionList) []target {
	var selectedPkgNames []string

	for _, pd := range selection {
		selectedPkgNames = append(selectedPkgNames,
			"dist_"+pd.PackageName)
	}

	targets := []target{target{"dist", true, selectedPkgNames, ""}}

	relBuildDir := ws.buildDirRelativeToWorkspace()

	scriptTemplate := `	@echo '[dist] %[1]s'
	@cd '` + relBuildDir + `/%[1]s' && \
	echo '--------------------------------' >> make_dist.log && \
	date >> make_dist.log && \
	echo '--------------------------------' >> make_dist.log && \
	$(MAKE) dist >> make_dist.log
	@mkdir -p dist
	@mv '` + relBuildDir + `/%[1]s/%[1]s-%[2]s.tar.gz' dist/
`

	for _, pd := range selection {
		dependencies := []string{
			path.Join(relBuildDir, pd.PackageName, "Makefile")}

		for _, dep := range selectedDeps[pd] {
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
