// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/doc"
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
	targets() []target
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

func (ht *helpTarget) targets() []target {
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
		MakeScript: script}}
}

type bootstrapTarget struct {
	selection packageDefinitionList
}

func createBootstrapTarget(selection packageDefinitionList) targetType {
	return &bootstrapTarget{selection}
}

func (*bootstrapTarget) name() string {
	return "bootstrap"
}

func (*bootstrapTarget) help() string {
	return "Unconditionally regenerate the 'configure' " +
		"scripts for the selected packages."
}

func (bt *bootstrapTarget) targets() []target {
	prefix := "bootstrap_"

	var dependencies []string

	for _, pd := range bt.selection {
		dependencies = append(dependencies, prefix+pd.PackageName)
	}

	bootstrapTargets := []target{{
		Target:       "bootstrap",
		Phony:        true,
		Dependencies: dependencies,
	}}

	scriptTemplate := `	@echo "[bootstrap] %[1]s"
	@cd ` + privateDirName + "/" + pkgDirName + `/%[1]s && ./autogen.sh
`

	for i, pd := range bt.selection {
		bootstrapTargets = append(bootstrapTargets,
			target{
				Target: dependencies[i],
				Phony:  true,
				MakeScript: fmt.Sprintf(scriptTemplate,
					pd.PackageName),
			},
			target{
				Target: privateDirName + "/" + pkgDirName +
					"/" + pd.PackageName + "/configure",
				MakeScript: "	@$(MAKE) -s " +
					dependencies[i] + "\n",
			})
	}

	return bootstrapTargets
}

type configureTarget struct {
}

func createConfigureTarget() targetType {
	return &configureTarget{}
}

func (*configureTarget) name() string {
	return "configure"
}

func (*configureTarget) help() string {
	return "Configure the selected packages using the " +
		"current options specified in the 'conftab' file."
}

func (*configureTarget) targets() []target {
	return nil
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

func (*buildTarget) targets() []target {
	return nil
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

func (*checkTarget) targets() []target {
	return nil
}
