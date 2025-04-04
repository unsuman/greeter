package main

import (
	"github.com/unsuman/greeter/pkg/plugin/external"
	english "github.com/unsuman/greeter/plugins/english/pkg"
)

func main() {
	plugin := english.New()
	external.Run(plugin)
}
