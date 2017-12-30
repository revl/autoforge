// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// BootstrapInDir bootstraps the package if 'configure' does not exist.
func bootstrapInDir(packageName, packageDir string) error {
	_, err := os.Lstat(filepath.Join(packageDir, "configure"))
	if os.IsNotExist(err) {
		fmt.Println("Bootstrapping " + packageName + "...")
		bootstrapCmd := exec.Command("./autogen.sh")
		bootstrapCmd.Dir = packageDir
		if err = bootstrapCmd.Run(); err != nil {
			return errors.New(filepath.Join(packageDir,
				"autogen.sh") + ": " + err.Error())
		}
	}

	return nil
}

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
			"FEATURE":             struct{}{},
			"PACKAGE":             struct{}{},
			"aix-soname":          struct{}{},
			"dependency-tracking": struct{}{},
			"fast-install":        struct{}{},
			"gnu-ld":              struct{}{},
			"libtool-lock":        struct{}{},
			"option-checking":     struct{}{},
			"pic":                 struct{}{},
			"pkgconfigdir":        struct{}{},
			"shared":              struct{}{},
			"silent-rules":        struct{}{},
			"static":              struct{}{},
			"sysroot":             struct{}{},
		}}
}

func (helpParser *configureHelpParser) printOptions(packageDir string) error {
	configureHelpCmd := exec.Command("./configure", "--help")
	configureHelpCmd.Dir = packageDir
	configureHelpStdout, err := configureHelpCmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = configureHelpCmd.Start(); err != nil {
		return err
	}
	helpScanner := bufio.NewScanner(configureHelpStdout)
	type optDescription struct {
		option           string
		arg              string
		description      string
		visibleInConftab bool
	}
	var options []optDescription
	var currentOption *optDescription

	for helpScanner.Scan() {
		helpLine := strings.TrimRight(helpScanner.Text(), " ")

		if helpLine == "" ||
			!strings.HasPrefix(helpLine, " ") {
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

		visible := false

		if key.optType != optOther {
			_, present := helpParser.ignoredFeatOrPkg[key.optName]
			if present {
				continue
			}
			visible = true
		}

		currentOption = &optDescription{
			opt, arg, descr, visible}
	}
	if err := helpScanner.Err(); err != nil {
		return err
	}
	if err = configureHelpCmd.Wait(); err != nil {
		return err
	}
	for _, opt := range options {
		if opt.visibleInConftab {
			fmt.Println(opt.option)
			fmt.Println(opt.arg)
			fmt.Println(opt.description)
		}
	}

	return nil
}

func generateAndBootstrapPackage(workspaceDir string,
	pkgSelection []string) error {
	packageIndex, err := readPackageDefinitions(workspaceDir)
	if err != nil {
		return err
	}

	privateDir := getPrivateDir(workspaceDir)

	pkgRootDir := filepath.Join(privateDir, "packages")

	type packageAndGenerator struct {
		pd         *packageDefinition
		packageDir string
		generator  func() error
	}

	var packagesAndGenerators []packageAndGenerator

	for _, packageName := range pkgSelection {
		pd, ok := packageIndex.packageByName[packageName]
		if !ok {
			return errors.New("no such package: " + packageName)
		}

		packageDir := filepath.Join(pkgRootDir, pd.packageName)

		generator, err := pd.getPackageGeneratorFunc(packageDir)
		if err != nil {
			return err
		}

		packagesAndGenerators = append(packagesAndGenerators,
			packageAndGenerator{pd, packageDir, generator})
	}

	params := templateParams{
		"makefile":       flags.makefile,
		"default_target": flags.defaultMakeTarget,
	}

	if err = generateWorkspaceFiles(workspaceDir, params); err != nil {
		return err
	}

	helpParser := createConfigureHelpParser()

	// Generate autoconf and automake sources for the selected packages.
	for _, pg := range packagesAndGenerators {
		if err = pg.generator(); err != nil {
			return err
		}
	}

	// Bootstrap the selected packages.
	for _, pg := range packagesAndGenerators {
		if err = bootstrapInDir(pg.pd.packageName,
			pg.packageDir); err != nil {
			return err
		}
	}

	for _, pg := range packagesAndGenerators {
		if err = helpParser.printOptions(pg.packageDir); err != nil {
			return err
		}
	}

	return nil
}

// SelectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select package_range...",
	Short: "Choose one or more packages to work on",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := generateAndBootstrapPackage(getWorkspaceDir(),
			args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)

	selectCmd.Flags().SortFlags = false
	addQuietFlag(selectCmd)
	addPkgPathFlag(selectCmd)
	addWorkspaceDirFlag(selectCmd)
	addMakefileFlag(selectCmd)
	addDefaultMakeTargetFlag(selectCmd)
}
