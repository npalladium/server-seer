package startup

import (
	"../../src"
	"../../src/logger"
	"../../src/storage"
	"encoding/json"
	"fmt"
	"strconv"
)

// Read and parse all configuration variables from the configuration file.
//
// This will include general config (ex: target server for sending the data)
// as well as all the enabled handlers
func GetConfiguration(fileName string) src.Configuration {

	input := src.GetFileContents(fileName)

	// Set default values
	configuration := src.Configuration{
		LogHandler:   "screen",
		CommandsFile: "commands.json",
		DatabaseFile: "storage.db",
		SendData:     false,
		SenderSettings: src.SenderSettings{
			Url:             "",
			EntriesPerCycle: 10,
			CycleFrequency:  30,
		},
		RuntimeData: src.RuntimeData{
			Commands:   make([]src.Command, 0),
			Processors: make([]src.Processor, 0),
		},
	}

	// Parse the JSON input to an interface
	json.Unmarshal(input, &configuration)

	// Handlers are not defined. JSON is most likely invalid
	if configuration.Handlers == nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Configuration file invalid"),
		)
	}

	return configuration
}

func InitializeLogger(configuration src.Configuration) {
	switch configuration.LogHandler {
	case "screen":
		logger.Logger = logger.ScreenLogger{}
	default:
		logger.Logger = logger.ScreenLogger{}
	}
}

// Creates the database connection.
//
// Creates the structure from scratch, if not defined already.
func SetupDatabase(configuration src.Configuration) {
	err := storage.OpenDatabase(configuration.DatabaseFile)

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error opening database: %s", err),
		)
	}

	storage.CreateStructure()

	if err != nil {
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Error creating structure: %s", err),
		)
	}
}

// Parses the commands from a JSON file and does basic validation
func InitializeCommands(configuration *src.Configuration) {

	input := src.GetFileContents(configuration.CommandsFile)

	var commands []src.Command
	json.Unmarshal(input, &commands)

	validateCommands(commands)

	configuration.RuntimeData.Commands = commands

}

func validateCommands(commands []src.Command) {
	var commandNames []string

	for _, command := range commands {
		if src.ContainsString(commandNames, command.Name) {
			src.ExitApplicationWithMessage(
				fmt.Sprintf("Command %s duplicated", command.Name),
			)
		}
		commandNames = append(commandNames, command.Name)
	}
}

// Create status processors.
//
// Will connect the configuration enabled commands and change the placeholders
func SetupProcessors(configuration *src.Configuration) {
	var processors []src.Processor

	// To create processors, loop through all handlers and find the command belonging
HandlerLoop:
	for _, handler := range configuration.Handlers {
		commandName := handler.CommandName

		for _, command := range configuration.RuntimeData.Commands {

			if command.Name == commandName {
				// Found the command, create processor and go to next handler

				processor := src.Processor{
					Command: command,
					Handler: handler,
				}

				// Generates the command from the handler placeholders
				processor.GenerateFinalCommand()

				processors = append(processors, processor)

				logger.Logger.Log(
					fmt.Sprintf(
						".. Added processor: %s (every %ds)",
						handler.Name,
						int32(handler.Frequency),
					),
				)

				logger.Logger.Log(
					fmt.Sprintf(
						"... Final command: %s",
						strconv.Quote(processor.FinalCommand),
					),
				)

				continue HandlerLoop

			}

		}

		// Command not found for the handler, error and exit
		src.ExitApplicationWithMessage(
			fmt.Sprintf("Command name '%s' not found", commandName),
		)

	}

	configuration.RuntimeData.Processors = processors

	// Test status processors for possible errors in commands
	testProcessors(processors)
}

func testProcessors(processors []src.Processor) {
	for _, processor := range processors {

		logger.Logger.Log(
			fmt.Sprintf(
				".. Processor: %s (%s)",
				processor.Handler.Name,
				processor.Command.Name,
			),
		)

		processor.RunOnce()

	}
}
