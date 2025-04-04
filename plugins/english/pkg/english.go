package english

import "github.com/unsuman/greeter/pkg/greetings"

// Greeter implements the Greeter interface in English language
type Greeter struct{}

// New creates a new English greeter
func New() greetings.Plugin {
	return &Greeter{}
}

func (g *Greeter) Hello() string {
	return "Hello!"
}

func (g *Greeter) GoodMorning() string {
	return "Good morning!"
}

func (g *Greeter) GoodAfternoon() string {
	return "Good afternoon!"
}

func (g *Greeter) GoodNight() string {
	return "Good night!"
}

func (g *Greeter) GoodBye() string {
	return "Goodbye!"
}

func (g *Greeter) Name() string {
	return "english"
}

func (g *Greeter) Init() error {
	return nil
}

func (g *Greeter) Close() error {
	return nil
}
