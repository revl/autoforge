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

func (reader *conftabReader) readSection() (*conftabSection,
	string, error) {
	section := conftabSection{make(map[string]bool), ""}

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
			return &section, line[1 : len(line)-1], nil
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

func (section *conftabSection) printSection(stream *os.File) {
	writer := bufio.NewWriter(stream)

	writer.WriteString(section.definition)

	writer.Flush()
}

type conftab struct {
	globalSection   *conftabSection
	packageSections []conftabSection
}

func readConftab(filename string) (*conftab, error) {
	conftabFile, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	conftabScanner := bufio.NewScanner(conftabFile)

	reader := conftabReader{filename, conftabScanner, 0,
		regexp.MustCompile(`^--([^\s\[=]+)`)}

	globalSection, nextSectionCaption, err := reader.readSection()

	if err != nil {
		return nil, err
	}

	if nextSectionCaption == "" {
		if err = conftabFile.Close(); err != nil {
			return nil, err
		}
	}

	return &conftab{globalSection, nil}, err
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

		conftab.globalSection.printSection(os.Stdout)
	},
}

func init() {
	RootCmd.AddCommand(dumpconftabCmd)

	dumpconftabCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(dumpconftabCmd)
}
