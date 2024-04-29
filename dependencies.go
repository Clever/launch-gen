package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	. "github.com/dave/jennifer/jen"
)

type ServiceDependency struct {
	name     string
	version  string
	override string
}

func parseOverrideDependencies(overrideDependenciesString *string, dependencies []string) map[string]string {

	// parsing through the list of overrides to make an original:new string map

	overrideDependenciesMap := make(map[string]string)

	if overrideDependenciesString == nil || *overrideDependenciesString == "" {
		return overrideDependenciesMap
	}

	overrideDependenciesList := strings.Split(*overrideDependenciesString, ",")
	for _, overrideRule := range overrideDependenciesList {
		depReplacementArr := strings.Split(overrideRule, ":")

		if len(depReplacementArr) != 2 || depReplacementArr[1] == "" {
			log.Fatal("usage: invalid formatting for the -d flag")
		}

		flag := 0
		for _, d := range dependencies {
			if d == depReplacementArr[0] {
				flag = 1
				break
			}
		}

		if flag == 0 {
			log.Fatal(depReplacementArr[0], " is not a dependency specified in the provided yaml file")
		}

		overrideDependenciesMap[depReplacementArr[0]] = depReplacementArr[1]
	}

	return overrideDependenciesMap
}

func parseDependencies(
	launchConfig *LaunchYML,
	skipDependencies flagsSet,
	overridesMap map[string]string,
) []ServiceDependency {
	versionedModuleRegex := regexp.MustCompile("(?i)^([a-z][a-z0-9-_]+)@(v[0-9]+)$")
	parsed := []ServiceDependency{}
	for _, d := range launchConfig.Dependencies {
		// skip
		if _, ok := skipDependencies[d]; ok {
			continue
		}

		// check for override
		overridePath, hasOverride := overridesMap[d]
		if hasOverride {
			parsed = append(parsed, ServiceDependency{name: d, override: overridePath})
			continue
		}

		// check for configured version
		submatches := versionedModuleRegex.FindStringSubmatch(d)
		if len(submatches) == 3 {
			parsed = append(parsed, ServiceDependency{name: submatches[1], version: submatches[2]})
			continue
		}

		// default: just the name
		parsed = append(parsed, ServiceDependency{name: d})
	}

	return parsed
}

func (d ServiceDependency) packageName() string {
	// with wagv9 onwards /gen-go/client is after the service name in the package path
	importPackage := fmt.Sprintf("%s/gen-go/client", d.name)
	if d.override != "" {
		importPackage = d.override
	} else if d.version != "" {
		importPackage = fmt.Sprintf("%s/%s", importPackage, d.version)
	}

	return fmt.Sprintf("github.com/Clever/%s", importPackage)
}

func mapDependenciesToCode(dependencies []ServiceDependency) []Code {
	imports := []Code{}
	for _, d := range dependencies {
		identifier := Id(strings.Title(toPublicVar(d.name)))

		// // with wagv9 onwards /gen-go/client is after the service name in the package path
		// importPackage := fmt.Sprintf("%s/gen-go/client", d.name)
		// if d.override != "" {
		// 	importPackage = d.override
		// } else if d.version != "" {
		// 	importPackage = fmt.Sprintf("%s/%s", importPackage, d.version)
		// }

		statement := identifier.Qual(d.packageName(), "Client")
		imports = append(imports, statement)
	}

	return imports
}
