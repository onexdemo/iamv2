package main

import (
	"github.com/onexdemo/iamv2/cmd/iam-authz-server/app"
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	// Build and launch the iam-authz-server application.
	app.NewApp().Run()
}
