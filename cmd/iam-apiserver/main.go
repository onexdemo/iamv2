package main

import (
	"github.com/onexdemo/iamv2/cmd/iam-apiserver/app"
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	// Build and launch the iam-apiserver application.
	app.NewApp().Run()
}
