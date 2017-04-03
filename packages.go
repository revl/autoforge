package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type packageDefinition struct {
	Name string `yaml:"name"`
}

func loadPackageDefinition(pathname string) packageDefinition {
	data, err := ioutil.ReadFile(pathname)

	if err != nil {
		panic(err)
	}

	var pd packageDefinition

	err = yaml.Unmarshal(data, &pd)

	if err != nil {
		panic(err)
	}

	return pd
}
