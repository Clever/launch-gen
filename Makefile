include golang.mk
.DEFAULT_GOAL := test

SHELL := /bin/bash
PKG := github.com/Clever/launch-gen
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run install_deps build fixtures

$(eval $(call golang-version-check,1.24))

fixtures: build
	rm -f fixtures/*.expected
	./bin/launch-gen -p packagename -skip-dependency dependency-to-skip fixtures/launch1.yml > fixtures/launch1.expected
	./bin/launch-gen -p packagename -skip-dependency dependency-to-skip -d dapple:dapple/gen-go/client/v5 fixtures/launch2.yml > fixtures/launch2.expected

test: build $(PKGS)
	diff <(./bin/launch-gen -p packagename -skip-dependency dependency-to-skip fixtures/launch1.yml) fixtures/launch1.expected
	diff <(./bin/launch-gen -p packagename -skip-dependency dependency-to-skip -d dapple:dapple/gen-go/client/v5 fixtures/launch2.yml) fixtures/launch2.expected

build:
	$(call golang-build,$(PKG),$(EXECUTABLE))

run: build
	bin/launch-gen

$(PKGS): golang-test-all-strict-cover-deps golang-setup-coverage
	$(call golang-test-all-strict-cover,$@)

install_deps:
	go mod vendor
