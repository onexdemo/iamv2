// Package all imports every iam-watcher plugin so that their init functions
// register them with the onex watch registry.
package all

import (
	_ "github.com/onexdemo/iamv2/internal/watcher/watcher/clean"
	_ "github.com/onexdemo/iamv2/internal/watcher/watcher/task"
)
