include golang.mk
.DEFAULT_GOAL := test

SHELL := /bin/bash
PKG := github.com/Clever/launch-gen
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run install_deps build fixtures

$(eval $(call golang-version-check,1.24))

fixtures:
	go test ./... -update

test: $(PKGS)

build:
	$(call golang-build,$(PKG),$(EXECUTABLE))

run: build
	bin/launch-gen

$(PKGS): golang-test-all-strict-cover-deps golang-setup-coverage
	$(call golang-test-all-strict-cover,$@)

install_deps:
	go mod vendor
