package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	packageName := flag.String("p", "main", "optional package name")
	outputFile := flag.String("o", "", "optional output to file. Default is stdout")
	skipDependencies := map[string]bool{}
	flag.Func("skip-dependency", "Dependency to skip generating wag clients. Can be added multiple times e.g. -skip-dependency a -skip-dependency b", func(s string) error {
		skipDependencies[s] = true
		return nil
	})
	overrideDependenciesString := flag.String("d", "", "Dependency name to override. You can provide multiple dependencies in the format dep1:replacementDep1,dep2:replacementDep2,...")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("usage: launch-gen [-p <package_name>] <file>")
	}

	output := os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("error opening file '%s': %s", *outputFile, err)
		}
		defer f.Close()
		output = f
	}

	data, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := generateFargate(*packageName, skipDependencies, *overrideDependenciesString, data, output); err != nil {
		log.Fatal(err)
	}
}
