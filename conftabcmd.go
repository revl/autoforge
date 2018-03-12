// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

func editConftab() error {
	workspaceDir, err := getWorkspaceDir()
	if err != nil {
		return err
	}

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return errors.New("neither $VISUAL nor $EDITOR is set")
	}

	privateDir := getPrivateDir(workspaceDir)

	conftabPathname := path.Join(privateDir, conftabFilename)

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

	if len(deletedSections) == 0 &&
		len(changedSections) == 0 &&
		len(addedSections) == 0 {
		fmt.Println("No effective changes detected")
		return nil
	}

	for _, pkgName := range deletedSections {
		fmt.Println("Deleted section: [" + pkgName + "]")
	}

	for _, pkgName := range addedSections {
		fmt.Println("New section: [" + pkgName + "]")
	}

	for pkgName, changes := range changedSections {
		fmt.Println("Changes in [" + pkgName + "]:")
		for _, chg := range changes {
			if chg.added != "" {
				if chg.deleted != "" {
					fmt.Println(chg.added +
						" (changed from " +
						chg.deleted + ")")
				} else {
					fmt.Println(chg.added + " (added)")
				}
			} else if chg.deleted != "" {
				fmt.Println("# " + chg.deleted + " (deleted)")
			}
		}
	}

	return nil
}

var conftabCmdName = "conftab"

var conftabCmd = &cobra.Command{
	Use:   conftabCmdName,
	Short: "Edit the conftab file",
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		if err := editConftab(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(conftabCmd)

	conftabCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(conftabCmd)
}
