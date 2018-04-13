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

type makefileTargetCollector struct {
	ws               *workspace
	relBuildDir      string
	pkgRootDir       string
	selection        packageDefinitionList
	selectedDeps     map[*packageDefinition]packageDefinitionList
	globalTargetDeps []string
	targets          []target
}

func createMakefileTargets(ws *workspace, selection packageDefinitionList,
	pi *packageIndex) []target {

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

	mtc := &makefileTargetCollector{ws,
		ws.buildDirRelativeToWorkspace(),
		ws.pkgRootDirRelativeToWorkspace(),
		selection, selectedDeps, globalTargetDeps, nil}

	mtc.addHelpTarget()
	mtc.addBootstrapTargets()
	mtc.addConfigureTargets()
	mtc.addBuildTargets()
	mtc.addCheckTargets()
	mtc.addInstallTargets()
	mtc.addDistTargets()

	return mtc.targets
}

func (mtc *makefileTargetCollector) makefileFor(pd *packageDefinition) string {
	return path.Join(mtc.relBuildDir, pd.PackageName, "Makefile")
}

func (mtc *makefileTargetCollector) configureFor(pd *packageDefinition) string {
	return path.Join(mtc.pkgRootDir, pd.PackageName, "configure")
}

func (mtc *makefileTargetCollector) addTarget(name string, phony bool,
	dependencies []string, makeScript string) {
	mtc.targets = append(mtc.targets,
		target{name, phony, dependencies, makeScript})
}

func (mtc *makefileTargetCollector) addHelpTarget() {
	mtc.addTarget("help", true, nil,
		`	@echo "Usage:"
	@echo "    make [target...]"
	@echo
	@echo "Global targets:"
	@echo "    help"
	@echo "        Display this help message. Unless overridden by the"
	@echo "        '--`+maketargetOption+
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
	@echo "            `+appName+" "+conftabCmdName+`"
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
	@echo "        '`+mtc.ws.installDir()+`'."
	@echo
	@echo "    dist"
	@echo "        Create distribution tarballs and move them to the"
	@echo "        'dist' subdirectory of the workspace."
	@echo
`)
}

func selfPathnameRelativeToWorkspace(ws *workspace) string {
	executable, err := os.Executable()
	if err != nil {
		return appName
	}

	return ws.relativeToWorkspace(executable)
}

func (mtc *makefileTargetCollector) addBootstrapTargets() {
	cmd := "\t@" + selfPathnameRelativeToWorkspace(mtc.ws) + " bootstrap "

	for _, pd := range mtc.selection {
		configurePathname := mtc.configureFor(pd)

		mtc.addTarget(configurePathname, false,
			[]string{configurePathname + ".ac"},
			cmd+pd.PackageName+"\n")
	}
}

func (mtc *makefileTargetCollector) addConfigureTargets() {
	relativeConftabPathname := path.Join(privateDirName, conftabFilename)

	cmd := "\t@" + selfPathnameRelativeToWorkspace(mtc.ws) + " configure "

	for _, pd := range mtc.selection {
		dependencies := []string{relativeConftabPathname,
			mtc.configureFor(pd)}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies,
				mtc.makefileFor(dep))
		}

		mtc.addTarget(mtc.makefileFor(pd), false,
			dependencies, cmd+pd.PackageName+"\n")
	}
}

func (mtc *makefileTargetCollector) scriptTemplate(targetName,
	projectTarget string) string {
	var logFileSuffix string
	if projectTarget != "" {
		logFileSuffix = "_" + projectTarget
		projectTarget = " " + projectTarget
	}
	return fmt.Sprintf(`	@echo '[%[1]s] %%[1]s'
	@cd '`+mtc.relBuildDir+`/%%[1]s' && \
	echo '--------------------------------' >> make%[2]s.log && \
	date >> make%[2]s.log && \
	echo '--------------------------------' >> make%[2]s.log && \
	$(MAKE)%[3]s >> make%[2]s.log
`, targetName, logFileSuffix, projectTarget)
}

func (mtc *makefileTargetCollector) addBuildTargets() {
	mtc.addTarget("build", true, mtc.globalTargetDeps, "")

	scriptTemplate := mtc.scriptTemplate("build", "")

	for _, pd := range mtc.selection {
		dependencies := []string{mtc.makefileFor(pd)}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, dep.PackageName)
		}

		mtc.addTarget(pd.PackageName, true, dependencies,
			fmt.Sprintf(scriptTemplate, pd.PackageName))
	}
}

func (mtc *makefileTargetCollector) addCheckTargets() {
	var selectedPkgNames []string

	for _, pd := range mtc.selection {
		selectedPkgNames = append(selectedPkgNames,
			"check_"+pd.PackageName)
	}

	mtc.addTarget("check", true, selectedPkgNames, "")

	scriptTemplate := mtc.scriptTemplate("check", "check")

	for _, pd := range mtc.selection {
		dependencies := []string{mtc.makefileFor(pd)}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies, dep.PackageName)
		}

		mtc.addTarget("check_"+pd.PackageName, true, dependencies,
			fmt.Sprintf(scriptTemplate, pd.PackageName))
	}
}

func (mtc *makefileTargetCollector) addInstallTargets() {
	var selectedPkgNames []string

	for _, dep := range mtc.globalTargetDeps {
		selectedPkgNames = append(selectedPkgNames,
			"install_"+dep)
	}

	mtc.addTarget("install", true, selectedPkgNames, "")

	scriptTemplate := mtc.scriptTemplate("install", "install")

	for _, pd := range mtc.selection {
		dependencies := []string{mtc.makefileFor(pd)}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies,
				"install_"+dep.PackageName)
		}

		mtc.addTarget("install_"+pd.PackageName, true, dependencies,
			fmt.Sprintf(scriptTemplate, pd.PackageName))
	}
}

func (mtc *makefileTargetCollector) addDistTargets() {
	var selectedPkgNames []string

	for _, pd := range mtc.selection {
		selectedPkgNames = append(selectedPkgNames,
			"dist_"+pd.PackageName)
	}

	mtc.addTarget("dist", true, selectedPkgNames, "")

	scriptTemplate := mtc.scriptTemplate("dist", "dist") +
		`	@mkdir -p dist
	@mv '` + mtc.relBuildDir + `/%[1]s/%[1]s-%[2]s.tar.gz' dist/
`

	for _, pd := range mtc.selection {
		dependencies := []string{mtc.makefileFor(pd)}

		for _, dep := range mtc.selectedDeps[pd] {
			dependencies = append(dependencies,
				mtc.makefileFor(dep))
		}

		mtc.addTarget("dist_"+pd.PackageName, true, dependencies,
			fmt.Sprintf(scriptTemplate, pd.PackageName,
				pd.params["version"]))
	}
}
