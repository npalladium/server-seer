package src

type Configuration struct {
	LogHandler     string
	CommandsFile   string
	DatabaseFile   string
	Handlers       []Handler
	SendData       bool
	SenderSettings SenderSettings
	RuntimeData    RuntimeData
}

type SenderSettings struct {
	Url             string
	ApplicationKey  string
	ServerHandler   string
	EntriesPerCycle int
	CycleFrequency  int
}

type RuntimeData struct {
	Commands   []Command
	Processors []Processor
}
