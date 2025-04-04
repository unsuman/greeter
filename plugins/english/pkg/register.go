package english

import "github.com/unsuman/greeter/pkg/plugin/registry"

func init() {
	registry.Register(New())
}
