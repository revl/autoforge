// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type optTypeType int

const (
	optFeat  optTypeType = iota // --enable-FEATURE type of options
	optPkg                      // --with-PACKAGE type of options
	optOther                    // all other options
)

type optionKey struct {
	optType optTypeType
	optName string
}

type conftabSection struct {
	pkgName    string               // "package" or "" if global section
	options    map[optionKey]string // "--opt=value" or "" if commented
	definition string               // verbatim text including newlines
}

func newSection(pkgName, definition string) *conftabSection {
	return &conftabSection{pkgName,
		make(map[optionKey]string), definition}
}

type conftabReader struct {
	filename      string
	scanner       *bufio.Scanner
	lineNumber    int
	optRegexp     *regexp.Regexp
	optClassifier optClassifier
}

func (reader *conftabReader) Err(message string) error {
	return fmt.Errorf("%s:%d: %s", reader.filename,
		reader.lineNumber, message)
}

type optClassifier struct {
	optTypeRegexp *regexp.Regexp
}

func createOptClassifier() optClassifier {
	return optClassifier{regexp.MustCompile(
		`^((enable|disable)|(with|without))-(.+)$`)}
}

func (classifier *optClassifier) classify(option string) (key optionKey) {
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

func (reader *conftabReader) readSection(pkgName string) (*conftabSection,
	string, error) {
	section := newSection(pkgName, "")

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
					"invalid section title format")
			}
			line = line[1 : len(line)-1]
			return section, strings.TrimSpace(line), nil
		}

		section.definition += line + "\n"

		var optDefinition string

		if line[0] == '#' {
			line = strings.TrimLeft(line, "#")
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
		} else if line[0] != '-' {
			return nil, "", reader.Err("invalid option format " +
				"(must start with a dash)")
		} else {
			optDefinition = line
		}

		matches := reader.optRegexp.FindStringSubmatch(line)
		if len(matches) > 1 {
			option := matches[1]
			key := reader.optClassifier.classify(option)
			section.options[key] = optDefinition
		}
	}

	return section, "", nil
}

type conftabStruct struct {
	globalSection        *conftabSection
	packageSections      []*conftabSection
	sectionByPackageName map[string]*conftabSection
}

func (conftab *conftabStruct) print(writer *bufio.Writer) (err error) {
	_, err = writer.WriteString(conftab.globalSection.definition)
	if err != nil {
		return
	}

	for _, section := range conftab.packageSections {
		_, err = writer.WriteString("[" + section.pkgName + "]\n")
		if err != nil {
			return
		}

		_, err = writer.WriteString(section.definition)
		if err != nil {
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
		createOptClassifier()}

	section, nextPkgName, err := reader.readSection("")

	if err != nil {
		return
	}

	conftab = &conftabStruct{section, nil, make(map[string]*conftabSection)}

	for nextPkgName != "" {
		pkgName := nextPkgName
		section, nextPkgName, err = reader.readSection(pkgName)
		if err != nil {
			return
		}
		conftab.packageSections = append(conftab.packageSections,
			section)
		conftab.sectionByPackageName[pkgName] = section
	}

	return
}

func newConftab() *conftabStruct {
	globalSection := newSection("", "\n")

	globalSection.addOption(&optDescription{optionKey{optFeat, "shared"},
		"Global defaults go here.", "--disable-shared"})

	return &conftabStruct{globalSection,
		nil, make(map[string]*conftabSection)}
}

type optDescription struct {
	key         optionKey
	description string
	definition  string
}

func (section *conftabSection) addOption(opt *optDescription) {
	// Novel options are commented out.
	section.options[opt.key] = ""

	section.definition = "# " + opt.description + "\n#" +
		opt.definition + "\n\n" + section.definition
}

func (conftab *conftabStruct) addOption(pkgName string,
	opt *optDescription) bool {
	section, found := conftab.sectionByPackageName[pkgName]
	if found {
		if _, found = section.options[opt.key]; found {
			return false
		}
	} else {
		section = newSection(pkgName, "\n")

		conftab.packageSections = append(conftab.packageSections,
			section)
		conftab.sectionByPackageName[pkgName] = section
	}

	section.addOption(opt)

	return true
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

func (conftab *conftabStruct) getConfigureArgs(pkgName string) []string {
	var args []string

	section, found := conftab.sectionByPackageName[pkgName]

	if !found {
		return args
	}

	for key, val := range section.options {
		if val != "" {
			args = append(args, val)
		} else if val = conftab.globalSection.options[key]; val != "" {
			args = append(args, val)
		}
	}

	return args
}

type sectionChange struct {
	deleted, added string
}

func (conftab *conftabStruct) diff(otherConftab *conftabStruct) (
	deletedSections []string,
	changedSections map[string][]sectionChange,
	addedSections []string) {

	for pkgName, origSection := range conftab.sectionByPackageName {
		section := otherConftab.sectionByPackageName[pkgName]
		if section == nil {
			deletedSections = append(deletedSections,
				origSection.pkgName)
		}
	}

	changedSections = make(map[string][]sectionChange)

	for pkgName, section := range otherConftab.sectionByPackageName {
		origSection := conftab.sectionByPackageName[pkgName]

		if origSection == nil {
			addedSections = append(addedSections, section.pkgName)
			continue
		}

		changes := changedSections[section.pkgName]

		for key, val := range origSection.options {
			if val == "" {
				val = conftab.globalSection.options[key]
				if val == "" {
					continue
				}
			}

			if section.options[key] == "" && otherConftab.
				globalSection.options[key] == "" {
				changes = append(changes,
					sectionChange{deleted: val})
			}
		}

		for key, val := range section.options {
			if val == "" {
				val = otherConftab.globalSection.options[key]
				// Deletions are discovered
				// in the previous loop.
				if val == "" {
					continue
				}
			}

			origVal := origSection.options[key]
			if origVal == "" {
				origVal = conftab.globalSection.options[key]
			}

			if val != origVal {
				changes = append(changes,
					sectionChange{origVal, val})
			}
		}

		if len(changes) > 0 {
			changedSections[section.pkgName] = changes
		}
	}

	return
}
