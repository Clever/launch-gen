package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/go-yaml/yaml"
)

// LaunchYML Schema
type LaunchYML struct {
	Env          []string `yaml:"env"`
	Dependencies []string `yaml:"dependencies"`
	Aws          struct {
		S3 struct {
			Read  []string `json:"read"`
			Write []string `json:"write"`
		} `json:"s3"`
	} `json:"aws"`
}

type flagsSet map[string]bool

func (fs *flagsSet) String() string {
	if fs == nil {
		return "nil"
	}

	values := make([]string, len(*fs))
	i := 0
	for k := range *fs {
		values[i] = k
		i++
	}

	return fmt.Sprintf("%s", values)
}

func (fs *flagsSet) Set(value string) error {
	(*fs)[value] = true
	return nil
}

const (
	funcGetS3NameByEnv = "getS3NameByEnv"
)

func sortedKeys(m map[string]struct{}) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func parseOverrideDependencies(overrideDependenciesString string, dependencies []string) map[string]string {

	// parsing through the list of overrides to make an original:new string map
	var overrideDependenciesList []string

	overrideDependenciesList = strings.Split(overrideDependenciesString, ",")
	overrideDependenciesMap := make(map[string]string)

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

// helper function that takes in a flag name and returns true if the flag was passed as an argument
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	t := LaunchYML{}
	packageName := flag.String("p", "main", "optional package name")
	outputFile := flag.String("o", "", "optional output to file. Default is stdout")
	var skipDependencies flagsSet = map[string]bool{}
	flag.Var(&skipDependencies, "skip-dependency", "Dependency to skip generating wag clients. Can be added mulitple times e.g. -skip-dependency a -skip-dependency b")
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

	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	f := NewFile(*packageName)
	f.Comment("Code generated by launch-gen DO NOT EDIT.")
	f.Id("")

	f.Comment("LaunchConfig is auto-generated based on the launch YML file")
	f.Type().Id("LaunchConfig").Struct(
		Id("Deps").Id("Dependencies"),
		Id("Env").Id("Environment"),
		Id("AwsResources"),
	)

	// parseOverrideDependencies parses the list of dependencies to be overwritten into a map
	var overrideDependenciesMap map[string]string

	if isFlagPassed("d") {
		if overrideDependenciesString != nil {
			overrideDependenciesMap = parseOverrideDependencies(*overrideDependenciesString, t.Dependencies)
		} else {
			log.Fatal("invalid dependency override arguments provided.")
		}
	}

	// Dependencies
	depsStruct := []Code{}
	depsInitDict := Dict{}
	for _, d := range t.Dependencies {
		if _, ok := skipDependencies[d]; ok {
			continue
		}

		importPackage, ok := overrideDependenciesMap[d]
		if !ok {
			importPackage = d
		}

		depsStruct = append(depsStruct, Id(strings.Title(toPublicVar(d))).Qual(fmt.Sprintf("github.com/Clever/%s/gen-go/client", importPackage), "Client"))
		depsInitDict[Id(strings.Title(toPublicVar(d)))] = Id(toPrivateVar(d))
	}
	f.Comment("Dependencies has clients for the service's dependencies")
	f.Type().Id("Dependencies").Struct(depsStruct...)

	// Environment
	envStruct := []Code{}
	envInitDict := Dict{}
	optionalEnvVars := []string{
		// Not used in dev
		"TRACING_ACCESS_TOKEN",
	}
	for _, s := range t.Env {
		envStruct = append(envStruct, List(Id(toPublicVar(s))).String())
		if contains(optionalEnvVars, s) {
			envInitDict[Id(toPublicVar(s))] = Id("os.Getenv").Call(Lit(s))
		} else {
			envInitDict[Id(toPublicVar(s))] = Id("requireEnvVar").Call(Lit(s))
		}
	}

	f.Comment("Environment has environment variables and their values")
	f.Type().Id("Environment").Struct(envStruct...)

	// AWS Resources
	awsStruct := []Code{}
	awsInitDict := Dict{}

	s3Buckets := map[string]struct{}{}
	for _, bucket := range t.Aws.S3.Read {
		s3Buckets[bucket] = struct{}{}
	}
	for _, bucket := range t.Aws.S3.Write {
		s3Buckets[bucket] = struct{}{}
	}

	for _, a := range sortedKeys(s3Buckets) {
		name := "S3" + toPublicVar(a)
		awsStruct = append(awsStruct, List(Id(name)).String())
		awsInitDict[Id(name)] = Id(funcGetS3NameByEnv).Call(Lit(a))
	}

	f.Comment("AwsResources contains string IDs that will help for accessing various AWS resources")
	f.Type().Id("AwsResources").Struct(awsStruct...)

	////////////////////
	// InitLaunchConfig() function
	////////////////////
	lines := []Code{}
	// Setup a wag client for each dependency
	for _, d := range t.Dependencies {
		if _, ok := skipDependencies[d]; ok {
			continue
		}
		c := []Code{}

		// checking to see if the dependency name has to be overwritten
		replacementString, ok := overrideDependenciesMap[d]
		if !ok {
			replacementString = d
		}

		c = []Code{
			List(Id(toPrivateVar(d)), Err()).Op(":=").Qual(fmt.Sprintf("github.com/Clever/%s/gen-go/client", replacementString), "NewFromDiscovery").Call(),
			If(Err().Op("!=").Nil()).Block(
				Qual("log", "Fatalf").Call(List(Lit("discovery error: %s"), Err())),
			),
		}

		lines = append(lines, c...)
	}

	// Return the full launch Config
	ret := Return(Id("LaunchConfig").Values(Dict{
		Id("Deps"):         Id("Dependencies").Values(depsInitDict),
		Id("Env"):          Id("Environment").Values(envInitDict),
		Id("AwsResources"): Id("AwsResources").Values(awsInitDict),
	}))

	lines = append(lines, ret)

	f.Comment("InitLaunchConfig creates a LaunchConfig")
	f.Func().Id("InitLaunchConfig").Params().Id("LaunchConfig").Block(lines...)

	f.Comment(`requireEnvVar exits the program immediately if an env var is not set`)
	f.Func().Id("requireEnvVar").Params(Id("s").String()).String().Block(
		List(Id("val"), Id("present")).Op(":=").Qual("os", "LookupEnv").Call(Id("s")),
		If(Op("!").Id("present")).Block(
			Qual("log", "Fatalf").Call(List(Lit("env var %s is not defined"), Id("s"))),
		),
		Return(Id("val")),
	)

	f.Comment(`getS3NameByEnv adds "-dev" to an env var name unless we're in "production" deploy env`)
	f.Comment(`We check both DEPLOY_ENV and _DEPLOY_ENV env vars, which are injected by our deployment system for Lambda and non-Lambda deployments, respectively`)
	f.Func().Id(funcGetS3NameByEnv).Params(Id("s").String()).String().Block(
		Id("env").Op(":=").Qual("os", "Getenv").Call(Lit("DEPLOY_ENV")),
		If(Id("env").Op("==").Lit("")).Block(
			Id("env").Op("=").Qual("os", "Getenv").Call(Lit("_DEPLOY_ENV")),
		),
		If(Id("env").Op("==").Lit("")).Block(
			Qual("log", "Fatal").Call(List(Lit("Unable to determine deployment environment (DEPLOY_ENV and _DEPLOY_ENV are undefined)"))),
		),
		If(Id("env").Op("==").Lit("production")).Block(
			Return(Id("s")),
		),
		Return(Id("s").Op("+").Lit("-dev")),
	)

	err = f.Render(output)
	if err != nil {
		panic(err)
	}
}

var (
	varOverrides []varOverride
)

type varOverride struct {
	old string
	new string
}

func init() {
	// Golint complains if certain strings are not capitalized
	varOverrides = []varOverride{
		varOverride{
			old: "Url",
			new: "URL",
		},
		varOverride{
			old: "Id",
			new: "ID",
		},
		varOverride{
			old: "Api",
			new: "API",
		},
	}
}

// FOO_BAR => FooBar
// foo-bar => FooBar
// foo => Foo
func toPublicVar(s string) string {
	list := []string{}
	if strings.Contains(s, "_") {
		list = strings.Split(s, "_")
	} else if strings.Contains(s, "-") {
		list = strings.Split(s, "-")
	} else {
		list = []string{s}
	}

	titledVar := ""
	for _, i := range list {
		titledVar += strings.Title(strings.ToLower(i))
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
