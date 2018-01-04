// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func printConftabDiff(origConftab, updatedConftab *conftabStruct) {
	conftabChanged := false

	var deletedSections []string

	for pkgName, origSection := range origConftab.sectionByPackageName {
		section := updatedConftab.sectionByPackageName[pkgName]
		if section == nil {
			deletedSections = append(deletedSections,
				origSection.title)
		}
	}

	if len(deletedSections) > 0 {
		//fmt.Println("Deleted sections:")

		for _, title := range deletedSections {
			fmt.Println("< " + title)
		}

		//fmt.Println("")
		conftabChanged = true
	}

	var addedSections []string

	for pkgName, section := range updatedConftab.sectionByPackageName {
		origSection := origConftab.sectionByPackageName[pkgName]

		if origSection == nil {
			addedSections = append(addedSections, section.title)
			continue
		}

		for key, val := range origSection.options {
			if val == "" {
				val = origConftab.globalSection.options[key]
			}
			if val != "" {
				newVal := section.options[key]
				if newVal == "" {
					newVal = updatedConftab.
						globalSection.options[key]
				}

				fmt.Println("Deleted value", val)
				conftabChanged = true
			}
		}

		for key, val := range section.options {
			if val == "" {
				val = updatedConftab.globalSection.options[key]
			}

			origVal := origSection.options[key]
			if origVal == "" {
				origVal = origConftab.globalSection.options[key]
			}

			if val != origVal {
				if origVal != "" {
					fmt.Println("< " + origVal)
				}
				if val != "" {
					fmt.Println("> " + val)
				}
				conftabChanged = true
			}
		}
	}

	if len(addedSections) > 0 {
		//fmt.Println("Added sections: " + strings.Join(addedSections, ", "))
		conftabChanged = true
	}

	for _, title := range addedSections {
		fmt.Println("> " + title)
	}

	if !conftabChanged {
		fmt.Println("No changes made")
	}
}

func editConftab(workspaceDir string) error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return errors.New("neither $VISUAL nor $EDITOR is set")
	}

	privateDir := getPrivateDir(workspaceDir)

	conftabPathname := filepath.Join(privateDir, "conftab")

	origConftab, err := readConftab(conftabPathname)
	if err != nil {
		return err
	}

	editorCmd := exec.Command(editor, conftabPathname)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	err = editorCmd.Run()
	if err != nil {
		return errors.New(editor + ": " + err.Error())
	}

	updatedConftab, err := readConftab(conftabPathname)
	if err != nil {
		return err
	}

	deletedSections, changedSections, addedSections :=
		origConftab.diff(updatedConftab)

	conftabChanged := false

	if len(deletedSections) > 0 {
		for _, title := range deletedSections {
			fmt.Println("< " + title)
		}

		conftabChanged = true
	}

	if len(changedSections) > 0 {
		for title, changes := range changedSections {
			fmt.Print("  " + title)
			for _, chg := range changes {
				if chg.deleted != "" {
					fmt.Println("< " + chg.deleted)
				}
				if chg.added != "" {
					fmt.Println("> " + chg.added)
				}
			}
			fmt.Println()
		}

		conftabChanged = true
	}

	if len(addedSections) > 0 {
		for _, title := range addedSections {
			fmt.Println("> " + title)
		}

		conftabChanged = true
	}

	if !conftabChanged {
		fmt.Println("No changes made")
	}

	return nil
}

var conftabCmdName = "conftab"

var conftabCmd = &cobra.Command{
	Use:   conftabCmdName,
	Short: "Edit the conftab file",
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		if err := editConftab(getWorkspaceDir()); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(conftabCmd)

	conftabCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(conftabCmd)
}
