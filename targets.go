// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"go/doc"
	"strings"
)

type target interface {
	Name() string
	IsPhony() bool
	Dependencies() []string
	Script() string

	help() string
}

type helpTarget struct {
	getGlobalTargets func() []target
}

func (*helpTarget) Name() string {
	return "help"
}

func (*helpTarget) IsPhony() bool {
	return true
}

func (*helpTarget) Dependencies() []string {
	return nil
}

func (ht *helpTarget) Script() string {
	script := `	@echo "Usage:"
	@echo "    make [target...]"
	@echo
	@echo "Global targets:"
`

	for _, t := range ht.getGlobalTargets() {
		script += "\t@echo \"    " + t.Name() + "\"\n"

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

	script += `	@echo "Individual package targets:"
`
	return script
}

func (*helpTarget) help() string {
	return "Display this help message. Unless overridden by the '--" +
		maketargetOption + "' option, this is the default target."
}

func createHelpTarget(getGlobalTargets func() []target) target {
	return &helpTarget{getGlobalTargets}
}

type bootstrapTarget struct {
}

func (*bootstrapTarget) Name() string {
	return "bootstrap"
}

func (*bootstrapTarget) IsPhony() bool {
	return true
}

func (*bootstrapTarget) Dependencies() []string {
	return nil
}

func (*bootstrapTarget) Script() string {
	return ""
}

func (*bootstrapTarget) help() string {
	return "Unconditionally regenerate the 'configure' " +
		"scripts for the selected packages."
}

func createBootstrapTarget() target {
	return &bootstrapTarget{}
}

type configureTarget struct {
}

func (*configureTarget) Name() string {
	return "configure"
}

func (*configureTarget) IsPhony() bool {
	return true
}

func (*configureTarget) Dependencies() []string {
	return nil
}

func (*configureTarget) Script() string {
	return ""
}

func (*configureTarget) help() string {
	return "Configure the selected packages using the " +
		"current options specified in the 'conftab' file."
}

func createConfigureTarget() target {
	return &configureTarget{}
}

type buildTarget struct {
}

func (*buildTarget) Name() string {
	return "build"
}

func (*buildTarget) IsPhony() bool {
	return true
}

func (*buildTarget) Dependencies() []string {
	return nil
}

func (*buildTarget) Script() string {
	return ""
}

func (*buildTarget) help() string {
	return "Build (compile and link) the selected packages. " +
		"For the packages that have not been configured, the " +
		"configuration step will be performed automatically."
}

func createBuildTarget() target {
	return &buildTarget{}
}

type checkTarget struct {
}

func (*checkTarget) Name() string {
	return "check"
}

func (*checkTarget) IsPhony() bool {
	return true
}

func (*checkTarget) Dependencies() []string {
	return nil
}

func (*checkTarget) Script() string {
	return ""
}

func (*checkTarget) help() string {
	return "Build and run unit tests for the selected packages."
}

func createCheckTarget() target {
	return &checkTarget{}
}
