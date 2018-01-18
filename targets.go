// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import "strings"

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

		for _, l := range strings.Split(t.help(), "\n") {
			if l != "" {
				script += "\t@echo \"        " + l + "\"\n"
			} else {
				script += "\t@echo\n"
			}
		}
	}

	return script
}

func (*helpTarget) help() string {
	return `Display this help message. Unless overridden by the
'--` + maketargetOption + `' option, this is the default target.
`
}

func createHelpTarget(getGlobalTargets func() []target) target {
	return &helpTarget{getGlobalTargets}
}
