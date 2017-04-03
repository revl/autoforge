package main

import "fmt"

func queryPackages(workspacedir, pkgpath string) {
	fmt.Println("List of packages:")

	pd := loadPackageDefinition("examples/packages/greeting/greeting.yaml")

	fmt.Println(pd.Name)
}
