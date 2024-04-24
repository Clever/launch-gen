# launch-gen

Generate code from launch YML

1. [Running](#running)
2. [Migrating to use in a Golang repo](#migrating-to-use-in-a-golang-repo)
3. [Dependency overrides](#dependency-overrides)

```sh
$ launch-gen --help
Usage of ./bin/launch-gen:
  -d string
        Dependency name to override. You can provide multiple dependencies in the format dep1:replacementDep1,dep2:replacementDep2,...
  -o string
        optional output to file. Default is stdout
  -p string
        optional package name (default "main")
  -skip-dependency value
        Dependency to skip generating wag clients. Can be added mulitple times e.g. -skip-dependency a -skip-dependency b
```

## Running

Build it

```
make build
```

Run it

```
./bin/launch-gen <path-to-launch-yml>
```

## Migrating to use in a Golang repo

This assumes you have a `go mod` repo.

1. Grab the latest version of `launch-gen`.

```bash
go get github.com/Clever/launch-gen
```

2. Within your repo's `Makefile`, declare `launch-gen` to be built on every `install_deps` invocation. Example:

```bash
install_deps:
    go mod vendor
    go build -o bin/launch-gen -mod=vendor ./vendor/github.com/Clever/launch-gen
```

3. Include as a blank import, to make sure `launch-gen` is declared as a dependency in `go.mod`. See [Clever/template-wag](https://github.com/Clever/template-wag/blob/2cbdec713ce3787970c25aa6c08a32ceb4c12336/tools/tools.go) for an example.

4. On the entry point of your application, include a stanza to generate the config. This should match the launch YAML in `launch/`.

```golang
// generate launch config
//go:generate sh -c "$PWD/bin/launch-gen -o launch.go -p main $PWD/launch/<your-application>.yml"
```

If you scope the current project's `/bin/` to the Makefile's path, then you can simplify the stanza:

```golang
// generate launch config
//go:generate launch-gen -o launch.go -p main $PWD/launch/<your-application>.yml
```

5. Ensure you call `go generate ./...` in the Makefile for paths that would be relevant for building or running the application. This allows the `//go:generate` stanza to run, which helps with building.

For example:

```bash
PKGS := $(shell go list ./... | grep -v /vendor | grep -v tools)
$(PKGS): generate golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)

# Before the `test` recipe is called, `make generate` is called as a prerequisite.
test: $(PKGS)

# Before the `build` recipe is called, `make generate` is called as a prerequisite.
build: generate
	$(call golang-build,$(PKG),$(EXECUTABLE))

# Create a target that will run `go generate` when `make generate` is called.
generate:
	go generate ./...
```

6. Run `make generate`. `launch.go` should be within the specified directory.

7. Call `InitLaunchConfig()` during startup of your program, and use it when needed.

## Dependency overrides

You can optionally override the package name for a service client dependency with the `-d` arg. Specifying `-d` will change the import statement in the generated `launch.go`. The primary use case for this arg is to handle modules which have the major version in their module name, because they will not be generated correctly by default.

Dependency mappings are defined as a colon-separated pair of strings, as in `configuredDependency:replacementDependency`. You can specify multiple overrides with a comma-separated string of these mappings: `dependency1:replacementDependency1,dependency2:replacementDependency2`.

Since `launch.go` depends on the `client` package within each module, `launch-gen` appends a `/gen-go/client` suffix to the module name to specific the service's client package. If you override the dependency, you **must** specify the whole package path yourself; `launch-gen` will not automatically insert the `/gen-go/client` path into the override package name, since it cannot assume any particular format.

```sh
$ launch-gen -d "servica-a:service-a/gen-go/client/v4,service-b:api-b-client/special/unique/client"
# service-a -> github.com/Clever/service-a/gen-go/client/v4
# service-b -> github.com/Clever/api-b-client/special/unique/client
```
