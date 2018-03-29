include golang.mk
.DEFAULT_GOAL := test

SHELL := /bin/bash
PKG := github.com/Clever/launch-gen
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE = $(shell basename $(PKG))

.PHONY: test $(PKGS) run install_deps build

$(eval $(call golang-version-check,1.9))

test: $(PKGS)

build: 
	$(call golang-build,$(PKG),$(EXECUTABLE))

run: build
	bin/launch-gen

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)
