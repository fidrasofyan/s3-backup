package tasks

import (
	"fmt"
	"runtime"
)

var AppVersion = "dev"
var AppCommit = "commit"
var AppBuildTime = "build time"

func Version() {
	fmt.Println("Runtime:", runtime.Version())
	fmt.Println("Version:", AppVersion)
	fmt.Println("Commit:", AppCommit)
	fmt.Println("Build time:", AppBuildTime)
}
