// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

const (
	optFeat = iota
	optPkg
	optOther
)

type optionKey struct {
	optType int
	optName string
}

type conftabSection struct {
	title      string
	options    map[optionKey]struct{}
	definition string
}

type conftabReader struct {
	filename      string
	scanner       *bufio.Scanner
	lineNumber    int
	optRegexp     *regexp.Regexp
	optClassifier featOrPkgClassifier
}

func (reader *conftabReader) Err(message string) error {
	return fmt.Errorf("%s:%d: %s", reader.filename,
		reader.lineNumber, message)
}

type featOrPkgClassifier struct {
	optTypeRegexp *regexp.Regexp
}

func createFeatOrPkgClassifier() featOrPkgClassifier {
	return featOrPkgClassifier{regexp.MustCompile(
		`^((enable|disable)|(with|without))-(.+)$`)}
}

func (classifier *featOrPkgClassifier) classify(option string) (key optionKey) {
	matches := classifier.optTypeRegexp.FindStringSubmatch(option)
	if len(matches) < 5 {
		key.optType = optOther
		key.optName = option
	} else {
		if matches[2] != "" {
			key.optType = optFeat
		} else {
			key.optType = optPkg
		}
		key.optName = matches[4]
	}
	return
}

func (reader *conftabReader) readSection(sectionName string) (*conftabSection,
	string, error) {
	section := conftabSection{sectionName, make(map[optionKey]struct{}), ""}

	for reader.scanner.Scan() {
		reader.lineNumber++
		line := strings.TrimSpace(reader.scanner.Text())

		if line == "" {
			section.definition += "\n"
			continue
		}

		if line[0] == '[' {
			if line[len(line)-1] != ']' {
				return nil, "", reader.Err(
					"invalid section caption format")
			}
			return &section, line + "\n", nil
		}

		section.definition += line + "\n"

		if line[0] == '#' {
			line = strings.TrimLeftFunc(
				strings.TrimLeft(line, "#"),
				unicode.IsSpace)
		} else if line[0] != '-' {
			return nil, "", reader.Err("invalid option format " +
				"(must start with a dash)")
		}

		matches := reader.optRegexp.FindStringSubmatch(line)
		if len(matches) > 1 {
			option := matches[1]
			key := reader.optClassifier.classify(option)
			section.options[key] = struct{}{}
		}
	}

	return &section, "", nil
}

type conftabStruct struct {
	globalSection        *conftabSection
	packageSections      []*conftabSection
	sectionByPackageName map[string]*conftabSection
}

func (conftab *conftabStruct) print(writer *bufio.Writer) (err error) {

	if _, err = writer.WriteString(
		conftab.globalSection.definition); err != nil {
		return
	}

	for _, section := range conftab.packageSections {
		if _, err = writer.WriteString(section.title); err != nil {
			return
		}

		if _, err = writer.WriteString(section.definition); err != nil {
			return
		}
	}

	return
}

func readConftab(pathname string) (conftab *conftabStruct, err error) {
	conftabFile, err := os.Open(pathname)

	if err != nil {
		return
	}

	defer func() {
		closeErr := conftabFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	conftabScanner := bufio.NewScanner(conftabFile)

	reader := conftabReader{pathname, conftabScanner, 0,
		regexp.MustCompile(`^--([^\s\[=]+)`),
		createFeatOrPkgClassifier()}

	section, nextSectionCaption, err := reader.readSection("")

	if err != nil {
		return
	}

	conftab = &conftabStruct{section, nil, make(map[string]*conftabSection)}

	for nextSectionCaption != "" {
		packageName := strings.TrimSpace(
			nextSectionCaption[1 : len(nextSectionCaption)-2])
		section, nextSectionCaption, err =
			reader.readSection(nextSectionCaption)
		if err != nil {
			return
		}
		conftab.packageSections = append(conftab.packageSections,
			section)
		conftab.sectionByPackageName[packageName] = section
	}

	return
}

func (conftab *conftabStruct) writeTo(pathname string) (err error) {
	conftabFile, err := os.Create(pathname)
	if err != nil {
		return
	}

	defer func() {
		closeErr := conftabFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	writer := bufio.NewWriter(conftabFile)

	if err = conftab.print(writer); err != nil {
		return
	}

	if err = writer.Flush(); err != nil {
		return
	}

	return
}

var dumpconftabCmd = &cobra.Command{
	Use:   "dumpconftab",
	Short: "Dump the conftab.ini file",
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		conftab, err := readConftab("conftab")
		if err != nil {
			log.Fatal(err)
		}

		if err = conftab.writeTo("conftab"); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(dumpconftabCmd)

	dumpconftabCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(dumpconftabCmd)
}
