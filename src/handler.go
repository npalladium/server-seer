package src

type Handler struct {
	Name         string
	Identifier   string
	CommandName  string `json:"command"`
	Frequency    float64
	Placeholders []Placeholder
}
