package lang

// English implements the Greeter interface in English language
type English struct{}

func NewEnglish() *English {
	return &English{}
}

func (e English) Hello() string {
	return "Hello!"
}

func (e English) GoodMorning() string {
	return "Good morning!"
}

func (e English) GoodAfternoon() string {
	return "Good afternoon!"
}

func (e English) GoodNight() string {
	return "Good night!"
}

func (e English) GoodBye() string {
	return "Goodbye!"
}
