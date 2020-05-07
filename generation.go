// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type configureHelpParser struct {
	optRegexp        *regexp.Regexp
	classifier       optClassifier
	ignoredFeatOrPkg map[string]struct{}
}

func createConfigureHelpParser() configureHelpParser {
	return configureHelpParser{
		regexp.MustCompile(`^--([^\s\[=]+)([^\s]*)\s*(.*)$`),
		createOptClassifier(),
		map[string]struct{}{
			"FEATURE":             {},
			"PACKAGE":             {},
			"aix-soname":          {},
			"dependency-tracking": {},
			"fast-install":        {},
			"gnu-ld":              {},
			"libtool-lock":        {},
			"option-checking":     {},
			"pkgconfigdir":        {},
			"silent-rules":        {},
			"sysroot":             {},
		}}
}

func (helpParser *configureHelpParser) parseOptions(packageDir string) (
	[]optDescription, error) {
	configureHelpCmd := exec.Command("./configure", "--help")
	configureHelpCmd.Dir = packageDir
	configureHelpStdout, err := configureHelpCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = configureHelpCmd.Start(); err != nil {
		return nil, err
	}
	helpScanner := bufio.NewScanner(configureHelpStdout)

	var options []optDescription
	var currentOption *optDescription

	for helpScanner.Scan() {
		helpLine := strings.TrimRight(helpScanner.Text(), " ")

		if helpLine == "" || !strings.HasPrefix(helpLine, " ") {
			if currentOption != nil {
				options = append(options, *currentOption)
				currentOption = nil
			}
			continue
		}

		helpLine = strings.TrimLeft(helpLine, " ")

		if strings.HasPrefix(helpLine, "-") {
			if currentOption != nil {
				options = append(options, *currentOption)
				currentOption = nil
			}
		} else {
			if currentOption != nil {
				if currentOption.description != "" {
					currentOption.description += " "
				}
				currentOption.description += helpLine
			}
			continue
		}

		parts := helpParser.optRegexp.FindStringSubmatch(helpLine)

		if len(parts) < 4 {
			continue
		}
		opt, arg, descr := parts[1], parts[2], parts[3]

		key := helpParser.classifier.classify(opt)

		if key.optType != optOther {
			_, present := helpParser.ignoredFeatOrPkg[key.optName]
			if present {
				continue
			}
		}

		currentOption = &optDescription{key, descr, "--" + opt + arg}
	}
	if err := helpScanner.Err(); err != nil {
		return nil, err
	}
	if err = configureHelpCmd.Wait(); err != nil {
		return nil, err
	}

	return options, nil
}

func generateAndBootstrapPackages(ws *workspace, pi *packageIndex,
	selection packageDefinitionList, conftab *Conftab) error {

	type packageAndGenerator struct {
		pd        *packageDefinition
		generator func() (bool, error)
	}

	var packagesAndGenerators []packageAndGenerator

	for _, pd := range selection {
		generator, err := pd.getPackageGeneratorFunc()
		if err != nil {
			return err
		}

		packagesAndGenerators = append(packagesAndGenerators,
			packageAndGenerator{pd, generator})
	}

	var packagesToBootstrap []packageAndGenerator

	// Generate autoconf and automake sources for the selected packages.
	for _, pg := range packagesAndGenerators {
		changed, err := pg.generator()
		if err != nil {
			return err
		}

		_, err = os.Stat(path.Join(pg.pd.packageDir(), "configure"))

		if changed || os.IsNotExist(err) {
			packagesToBootstrap = append(packagesToBootstrap, pg)
		}
	}

	if !flags.noBootstrap {
		// Bootstrap the selected packages.
		for _, pg := range packagesToBootstrap {
			err := bootstrapPackage(pg.pd)
			if err != nil {
				return err
			}
		}

		helpParser := createConfigureHelpParser()

		for _, pg := range packagesAndGenerators {
			options, err := helpParser.parseOptions(
				pg.pd.packageDir())
			if err != nil {
				return err
			}

			for _, opt := range options {
				if opt.key.optType != optOther {
					conftab.addOption(
						pg.pd.PackageName, &opt)
				}
			}
		}
	}

	return generateWorkspaceFiles(ws, pi, selection, conftab)
}
