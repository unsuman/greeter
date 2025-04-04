package japanese

import "github.com/unsuman/greeter/pkg/greetings"

type Greeter struct{}

func New() greetings.Plugin {
	return &Greeter{}
}

func (g *Greeter) Hello() string {
	return "こんにちは! (Konnichiwa)"
}

func (g *Greeter) GoodMorning() string {
	return "おはようございます! (Ohayou gozaimasu)"
}

func (g *Greeter) GoodAfternoon() string {
	return "こんにちは! (Konnichiwa)"
}

func (g *Greeter) GoodNight() string {
	return "おやすみなさい! (Oyasumi nasai)"
}

func (g *Greeter) GoodBye() string {
	return "さようなら! (Sayounara)"
}

func (g *Greeter) Name() string {
	return "japanese"
}

func (g *Greeter) Init() error {
	return nil
}

func (g *Greeter) Close() error {
	return nil
}
