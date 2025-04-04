package main

import (
	"github.com/unsuman/greeter/pkg/plugin/external"
	japanese "github.com/unsuman/greeter/plugins/japanese/pkg"
)

func main() {
	plugin := japanese.New()
	external.Run(plugin)
}
