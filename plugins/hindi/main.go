package main

import (
	"github.com/unsuman/greeter/pkg/plugin/external"
	hindi "github.com/unsuman/greeter/plugins/hindi/pkg"
)

func main() {
	plugin := hindi.New()
	external.Run(plugin)
}
