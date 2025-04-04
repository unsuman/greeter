package hindi

import "github.com/unsuman/greeter/pkg/greetings"

type Greeter struct{}

func New() greetings.Plugin {
	return &Greeter{}
}

func (g *Greeter) Hello() string {
	return "नमस्ते! (Namaste!)"
}

func (g *Greeter) GoodMorning() string {
	return "शुभ प्रभात! (Shubh Prabhat!)"
}

func (g *Greeter) GoodAfternoon() string {
	return "शुभ दोपहर! (Shubh Dophar)"
}

func (g *Greeter) GoodNight() string {
	return "शुभ रात्रि! (Shubh Ratri!)"
}

func (g *Greeter) GoodBye() string {
	return "अलविदा! (Alvida!)"
}

func (g *Greeter) Name() string {
	return "hindi"
}

func (g *Greeter) Init() error {
	return nil
}

func (g *Greeter) Close() error {
	return nil
}
