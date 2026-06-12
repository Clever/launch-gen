package main

import (
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

const wagClientSuffix = "/gen-go/client"

func cleverImportPath(depName, pathSuffix string) string {
	return "github.com/Clever/" + depName + pathSuffix
}

type varOverride struct {
	old string
	new string
}

var varOverrides = []varOverride{
	{old: "Url", new: "URL"},
	{old: "Id", new: "ID"},
	{old: "Api", new: "API"},
}

// FOO_BAR => FooBar
// foo-bar => FooBar
// foo.bar => FooBar
// foo => Foo
func toPublicVar(s string) string {
	s = strings.ReplaceAll(strings.ToUpper(s), ".", "_")
	s = strings.ReplaceAll(strings.ToUpper(s), "-", "_")
	list := strings.Split(s, "_")

	titledVar := ""
	for _, word := range list {
		word = strings.ToLower(word)
		if word != "" {
			titledVar += strings.ToUpper(word[:1]) + word[1:]
		}
	}

	out := titledVar
	for _, override := range varOverrides {
		out = strings.Replace(out, override.old, override.new, 1)
	}

	return out
}

// FOO_BAR => fooBar
// foo-bar => fooBar
// foo => foo
func toPrivateVar(s string) string {
	if s == "" {
		return s
	}
	out := toPublicVar(s)
	return strings.ToLower(string(out[0])) + out[1:]
}

func contains(many []string, one string) bool {
	for _, m := range many {
		if m == one {
			return true
		}
	}
	return false
}

func parseOverrideDependencies(overrideDependenciesString *string, dependencies []string) map[string]string {
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

		if !contains(dependencies, depReplacementArr[0]) {
			log.Fatal(depReplacementArr[0], " is not a dependency specified in the provided yaml file")
		}

		overrideDependenciesMap[depReplacementArr[0]] = depReplacementArr[1]
	}

	return overrideDependenciesMap
}

func generateDependencies(f *jen.File, deps []string, skip map[string]bool, overrides map[string]string) (jen.Dict, []jen.Code) {
	depsStruct := []jen.Code{}
	depsInitDict := jen.Dict{}
	for _, d := range deps {
		if _, ok := skip[d]; ok {
			continue
		}
		importPackage, pathSuffix := resolveDepImport(d, overrides)
		depsStruct = append(depsStruct, jen.Id(toPublicVar(d)).Qual(cleverImportPath(importPackage, pathSuffix), "Client"))
		depsInitDict[jen.Id(toPublicVar(d))] = jen.Id(toPrivateVar(d))
	}
	f.Comment("Dependencies has clients for the service's dependencies")
	f.Type().Id("Dependencies").Struct(depsStruct...)
	return depsInitDict, buildDepInitLines(deps, skip, overrides)
}

func resolveDepImport(dep string, overrides map[string]string) (depName, pathSuffix string) {
	if override, hasOverride := overrides[dep]; hasOverride {
		return override, ""
	}
	return dep, wagClientSuffix
}

func buildDepInitLines(deps []string, skip map[string]bool, overrides map[string]string) []jen.Code {
	atLeastOneDep := false
	for _, d := range deps {
		if _, skipped := skip[d]; !skipped {
			atLeastOneDep = true
			break
		}
	}
	if !atLeastOneDep {
		return []jen.Code{}
	}

	initLines := []jen.Code{
		jen.Id("var exporter ").Qual("go.opentelemetry.io/otel/sdk/trace", "SpanExporter"),
		jen.If(jen.Id("exp").Op("==").Nil().Block(
			jen.Id("exporter").Op("=").Qual("go.opentelemetry.io/otel/sdk/trace/tracetest", "NewNoopExporter").Call(),
		)).Else().Block(
			jen.Id("exporter").Op("=").Id("*exp"),
		),
	}

	for _, d := range deps {
		if _, ok := skip[d]; ok {
			continue
		}
		depName, pathSuffix := resolveDepImport(d, overrides)
		initLines = append(initLines, []jen.Code{
			jen.List(jen.Id(toPrivateVar(d)), jen.Err()).Op(":=").
				Qual(cleverImportPath(depName, pathSuffix), "NewFromDiscovery").
				Call(jen.Qual("github.com/Clever/wag/clientconfig/v9", "WithTracing").Call(jen.Lit(d), jen.Id("exporter"))),
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Qual("log", "Fatalf").Call(jen.List(jen.Lit("discovery error: %s"), jen.Err())),
			),
		}...)
	}

	return initLines
}

func emitRequireEnvVar(f *jen.File) {
	f.Comment(`requireEnvVar exits the program immediately if an env var is not set`)
	f.Func().Id("requireEnvVar").Params(jen.Id("s").String()).String().Block(
		jen.List(jen.Id("val"), jen.Id("present")).Op(":=").Qual("os", "LookupEnv").Call(jen.Id("s")),
		jen.If(jen.Op("!").Id("present")).Block(
			jen.Qual("log", "Fatalf").Call(jen.List(jen.Lit("env var %s is not defined"), jen.Id("s"))),
		),
		jen.Return(jen.Id("val")),
	)
}
