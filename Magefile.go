//go:build mage
// +build mage

package main

import (
	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
	"github.com/magefile/mage/mg"
)

func Build4() { //revive:disable-line
	b := build.Build{}
	mg.Deps(b.Windows, b.Darwin, b.DarwinARM64, b.Linux, b.LinuxARM64, b.GenerateManifestFile)
}

// Default configures the default target.
var Default = Build4
