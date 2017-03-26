package src

import (
	"../src/logger"
	"../src/storage"
	"fmt"
	"strings"
	"time"
)

type Processor struct {
	Command      Command
	Handler      Handler
	FinalCommand string
	Outputs      []ProcessorOutput
}

type ProcessorOutput struct {
	Output    string
	Timestamp int32
}

func (self *Processor) GenerateFinalCommand() {
	self.FinalCommand = self.Command.Command
	for _, placeholder := range self.Handler.Placeholders {
		self.FinalCommand = strings.Replace(
			self.FinalCommand,
			"<"+placeholder.Name+">",
			placeholder.Value,
			-1,
		)
	}

}

func (self Processor) Run(channel chan storage.OutputEntry) {
	frequencyDuration := time.Second * time.Duration(self.Handler.Frequency)
	ticker := time.NewTicker(frequencyDuration)
	for {
		output := self.RunOnce()
		// self.Outputs = append(
		// 	self.Outputs,
		// 	ProcessorOutput{
		// 		Output:    output,
		// 		Timestamp: int32(time.Now().Unix()),
		// 	},
		// )

		// Notify channel with a new entry
		channel <- storage.OutputEntry{
			HandlerIdentifier: self.Handler.Identifier,
			CommandName:       self.Command.Name,
			Output:            output,
			Timestamp:         int32(time.Now().Unix()),
		}

		// time.Sleep(time.Second * time.Duration(self.Handler.GetCheckFrequency()))
		<-ticker.C
	}
}

func (self Processor) RunOnce() string {

	logger.Logger.Log(
		fmt.Sprintf(
			"[%s] Running '%s'",
			time.Now().Format(time.RFC3339),
			self.Handler.Name,
		),
	)

	output := RunCommand(self.FinalCommand)

	logger.Logger.Log(
		fmt.Sprintf(
			"[%s] Finished running '%s'. Output: '%s'",
			time.Now().Format(time.RFC3339),
			self.Handler.Name,
			output,
		),
	)

	return output
}
