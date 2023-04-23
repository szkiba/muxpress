// SPDX-FileCopyrightText: 2023 Iván Szkiba
//
// SPDX-License-Identifier: MIT

//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/princjef/mageutil/bintool"
	"github.com/princjef/mageutil/shellcmd"
)

var Default = All

var linter = bintool.Must(bintool.New(
	"golangci-lint{{.BinExt}}",
	"1.51.1",
	"https://github.com/golangci/golangci-lint/releases/download/v{{.Version}}/golangci-lint-{{.Version}}-{{.GOOS}}-{{.GOARCH}}{{.ArchiveExt}}",
))

func Lint() error {
	if err := linter.Ensure(); err != nil {
		return err
	}

	return linter.Command(`run`).Run()
}

var docgen = bintool.Must(bintool.New(
	"gomarkdoc{{.BinExt}}",
	"0.4.1",
	"https://github.com/princjef/gomarkdoc/releases/download/v{{.Version}}/gomarkdoc_{{.Version}}_{{.GOOS}}_{{.GOARCH}}{{.ArchiveExt}}",
))

func Doc() error {
	if err := docgen.Ensure(); err != nil {
		return err
	}

	gomd, err := docgen.Command(
		`--config magefiles/.gomarkdoc.yml github.com/szkiba/muxpress`,
	).Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(gomd))

		return err
	}

	if err := yarn("typedoc").Run(); err != nil {
		return err
	}

	tsmd, err := yarn("concat-md  --decrease-title-levels --start-title-level-at 2 dist/docs").Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(tsmd))

		return err
	}

	api := []byte("# muxpress API\n")

	api = append(api, tsmd...)

	out := append(gomd, tsmd...)

	if err := os.WriteFile("README.md", out, 0o644); err != nil {
		return err
	}

	return os.WriteFile("api/README.md", api, 0o644)
}

func Prepare() error {
	if err := yarn("install").Run(); err != nil {
		return err
	}

	return shellcmd.Command("pipx install reuse").Run()
}

func Test() error {
	return shellcmd.Command(`go test -count 1 -coverprofile=coverage.txt ./...`).Run()
}

func Coverage() error {
	return shellcmd.Command(`go tool cover -html=coverage.txt`).Run()
}

func glob(patterns ...string) (string, error) {
	buff := new(strings.Builder)

	for _, p := range patterns {
		m, err := filepath.Glob(p)
		if err != nil {
			return "", err
		}

		_, err = buff.WriteString(strings.Join(m, " ") + " ")
		if err != nil {
			return "", err
		}
	}

	return buff.String(), nil
}

func License() error {
	all, err := glob("*.go", "*/*.go", ".*.yml", ".gitignore", "*/.gitignore", "*.ts", "*/*ts", ".github/workflows/*")
	if err != nil {
		return err
	}

	return shellcmd.Command(
		`reuse annotate --copyright "Iván Szkiba" --merge-copyrights --license MIT --skip-unrecognised ` + all,
	).Run()
}

func yarn(arg string) shellcmd.Command {
	return shellcmd.Command("yarn --silent --cwd magefiles " + arg)
}

func Clean() error {
	sh.Rm("magefiles/dist")
	sh.Rm("magefiles/bin")
	sh.Rm("magefiles/node_modules")
	sh.Rm("magefiles/yarn.lock")
	sh.Rm("coverage.txt")
	sh.Rm("bin")

	return nil
}

func All() error {
	if err := Lint(); err != nil {
		return err
	}

	return Test()
}
