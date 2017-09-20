// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type initParams struct {
	OutputDir string `yaml:"outputdir,omitempty"`
}

func loadInitParams() initParams {
	return initParams{}
}

func initializeWorkspace() error {

	fmt.Printf("Initializing a new workspace for %s...\n", appName)

	ip := initParams{OutputDir: "/home"}

	out, err := yaml.Marshal(&ip)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new workspace",
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeWorkspace(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(initCmd)
	addInstallDirFlag(initCmd)
}
