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
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type conftabSection struct {
	caption    string
	options    map[string]bool
	definition string
}

type conftabReader struct {
	filename   string
	scanner    *bufio.Scanner
	lineNumber int
	optRegexp  *regexp.Regexp
}

func (reader *conftabReader) Err(message string) error {
	return errors.New(fmt.Sprintf("%s:%d: %s", reader.filename,
		reader.lineNumber, message))
}

func (reader *conftabReader) readSection(sectionName string) (*conftabSection,
	string, error) {
	section := conftabSection{sectionName, make(map[string]bool), ""}

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
			section.options[matches[1]] = true
		}
	}

	return &section, "", nil
}

type conftabStruct struct {
	globalSection        *conftabSection
	packageSections      []*conftabSection
	sectionByPackageName map[string]*conftabSection
}

func (conftab *conftabStruct) print(writer *bufio.Writer) {
	writer.WriteString(conftab.globalSection.definition)

	for _, section := range conftab.packageSections {
		writer.WriteString(section.caption)
		writer.WriteString(section.definition)
	}
}

func readConftab(filename string) (*conftabStruct, error) {
	conftabFile, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	conftabScanner := bufio.NewScanner(conftabFile)

	reader := conftabReader{filename, conftabScanner, 0,
		regexp.MustCompile(`^--([^\s\[=]+)`)}

	section, nextSectionCaption, err := reader.readSection("")

	if err != nil {
		return nil, err
	}

	conftab := conftabStruct{section, nil,
		make(map[string]*conftabSection)}

	for nextSectionCaption != "" {
		packageName := strings.TrimSpace(
			nextSectionCaption[1 : len(nextSectionCaption)-2])
		section, nextSectionCaption, err =
			reader.readSection(nextSectionCaption)
		if err != nil {
			return nil, err
		}
		conftab.packageSections = append(conftab.packageSections,
			section)
		conftab.sectionByPackageName[packageName] = section
	}

	if err = conftabFile.Close(); err != nil {
		return nil, err
	}

	return &conftab, nil
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

		writer := bufio.NewWriter(os.Stdout)

		conftab.print(writer)

		writer.Flush()
	},
}

func init() {
	RootCmd.AddCommand(dumpconftabCmd)

	dumpconftabCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(dumpconftabCmd)
}
