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

// ConftabSection contains a multiline plain text definition
// of the conftab section for the given package.
type ConftabSection struct {
	PkgName    string               // "package" or "" if global section
	Definition string               // verbatim text including newlines
	options    map[optionKey]string // "--opt=value" or "" if commented
}

// Conftab contains definitions as well as an index of all conftab sections.
type Conftab struct {
	GlobalSection        *ConftabSection
	PackageSections      []*ConftabSection
	sectionByPackageName map[string]*ConftabSection
}

func newSection(pkgName, definition string) *ConftabSection {
	return &ConftabSection{pkgName, definition,
		make(map[optionKey]string)}
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

func (reader *conftabReader) readSection(pkgName string) (*ConftabSection,
	string, error) {
	section := newSection(pkgName, "")

	for reader.scanner.Scan() {
		reader.lineNumber++
		line := strings.TrimSpace(reader.scanner.Text())

		if line == "" {
			section.Definition += "\n"
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

		section.Definition += line + "\n"

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

func readConftab(pathname string) (conftab *Conftab, err error) {
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

	conftab = &Conftab{section, nil, make(map[string]*ConftabSection)}

	for nextPkgName != "" {
		pkgName := nextPkgName
		section, nextPkgName, err = reader.readSection(pkgName)
		if err != nil {
			return
		}
		conftab.PackageSections = append(conftab.PackageSections,
			section)
		conftab.sectionByPackageName[pkgName] = section
	}

	return
}

func newConftab() *Conftab {
	globalSection := newSection("", "\n")

	globalSection.addOption(&optDescription{optionKey{optFeat, "shared"},
		"Global defaults go here.", "--disable-shared"})

	return &Conftab{globalSection,
		nil, make(map[string]*ConftabSection)}
}

type optDescription struct {
	key         optionKey
	description string
	definition  string
}

func (section *ConftabSection) addOption(opt *optDescription) {
	// Novel options are commented out.
	section.options[opt.key] = ""

	section.Definition = "# " + opt.description + "\n#" +
		opt.definition + "\n\n" + section.Definition
}

func (conftab *Conftab) addOption(pkgName string,
	opt *optDescription) bool {
	section, found := conftab.sectionByPackageName[pkgName]
	if found {
		if _, found = section.options[opt.key]; found {
			return false
		}
	} else {
		section = newSection(pkgName, "\n")

		conftab.PackageSections = append(conftab.PackageSections,
			section)
		conftab.sectionByPackageName[pkgName] = section
	}

	section.addOption(opt)

	return true
}

func (conftab *Conftab) getConfigureArgs(pkgName string) []string {
	var args []string

	section, found := conftab.sectionByPackageName[pkgName]

	if !found {
		return args
	}

	for key, val := range section.options {
		if val != "" {
			args = append(args, val)
		} else if val = conftab.GlobalSection.options[key]; val != "" {
			args = append(args, val)
		}
	}

	return args
}

type sectionChange struct {
	deleted, added string
}

func (conftab *Conftab) diff(otherConftab *Conftab) (
	deletedSections []string,
	changedSections map[string][]sectionChange,
	addedSections []string) {

	for pkgName, origSection := range conftab.sectionByPackageName {
		section := otherConftab.sectionByPackageName[pkgName]
		if section == nil {
			deletedSections = append(deletedSections,
				origSection.PkgName)
		}
	}

	changedSections = make(map[string][]sectionChange)

	for pkgName, section := range otherConftab.sectionByPackageName {
		origSection := conftab.sectionByPackageName[pkgName]

		if origSection == nil {
			addedSections = append(addedSections, section.PkgName)
			continue
		}

		changes := changedSections[section.PkgName]

		for key, val := range origSection.options {
			if val == "" {
				val = conftab.GlobalSection.options[key]
				if val == "" {
					continue
				}
			}

			if section.options[key] == "" && otherConftab.
				GlobalSection.options[key] == "" {
				changes = append(changes,
					sectionChange{deleted: val})
			}
		}

		for key, val := range section.options {
			if val == "" {
				val = otherConftab.GlobalSection.options[key]
				// Deletions are discovered
				// in the previous loop.
				if val == "" {
					continue
				}
			}

			origVal := origSection.options[key]
			if origVal == "" {
				origVal = conftab.GlobalSection.options[key]
			}

			if val != origVal {
				changes = append(changes,
					sectionChange{origVal, val})
			}
		}

		if len(changes) > 0 {
			changedSections[section.PkgName] = changes
		}
	}

	return
}
