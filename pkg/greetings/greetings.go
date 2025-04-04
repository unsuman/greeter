package greetings

// Greeter defines the interface for different greeting types
type Greeter interface {
	Hello() string
	GoodMorning() string
	GoodAfternoon() string
	GoodNight() string
	GoodBye() string
}

type Plugin interface {
	Greeter
	Name() string
	Init() error
	Close() error
}
