package main

import (
	"../src"
	"../src/logger"
	"../src/sender"
	"../src/startup"
	"../src/storage"
	// "sync"
	"runtime"
	"time"
)

// The main function is responsible for parsing and validating the settings.
//
// From there, the main function calls 'startup' functions to start various
// parts of the application.
//
// In no way, this function should contain any calculations or calls to anywhere
// but the startup package or this file
//
// This function should contain ALL channel variables
func main() {

	// General application defaults
	configurationFile := "configuration_small.json"

	// Configuration
	configuration := startup.GetConfiguration(configurationFile)

	// Logger
	startup.InitializeLogger(configuration)

	logger.Logger.Log("Configuration parsed, logger initialized. Starting.")

	// Database
	logger.Logger.Log(". Setting up local database")
	startup.SetupDatabase(configuration)

	// Commands - all possible actions that handler will be able to use
	//
	// Will update configuration.RuntimeData.Commands
	logger.Logger.Log(". Parsing commands")
	startup.InitializeCommands(&configuration)

	// Processors - wraps the handlers of commands for processing.
	//
	// Will test each processor for possible runtime errors.
	//
	// Will update configuration.RuntimeData.Processors
	logger.Logger.Log(". Setting up status processors")
	startup.SetupProcessors(&configuration)

	/*
	 *	Run the application
	 */
	logger.Logger.Log(". Running")

	channelOutputedProcessor := make(chan *src.Processor)
	startProcessors(configuration, channelOutputedProcessor)

	startStorageListener(channelOutputedProcessor)

	startEntrySender(configuration)

	// Exit the main thread. Application keeps on running.
	runtime.Goexit()

}

// Start running the processors.
//
// This will create a separate goroutine for each processor.
func startProcessors(configuration src.Configuration, channelOutputedProcessor chan *src.Processor) {

	// Loop and run all the processors
	for _, processor := range configuration.RuntimeData.Processors {
		// Add some time between starting to make log reading easier
		time.Sleep(time.Second * 1)

		// Run processor in a new goroutine
		logger.Logger.Log(".. Running " + processor.Handler.Name)
		go processor.Run(channelOutputedProcessor)
	}
}

// Start the storage listener - will listen to channel messages and store
// data to the database.
func startStorageListener(channelOutputedProcessor chan *src.Processor) {
	go func(channelOutputedProcessor chan *src.Processor) {
		for {
			processor := <-channelOutputedProcessor

			// Create the output entries and store them in the local sqlite database
			outputEntries := processProcessorToOutputEntry(processor)

			storage.StoreOutputEntries(outputEntries)

			// outputEntry.Store()

		}
	}(channelOutputedProcessor)
}

// Starts the entry sender - send all flagged entries to a remote API.
func startEntrySender(configuration src.Configuration) {
	if !configuration.SendData || configuration.SenderSettings.Url == "" {
		return
	}

	logger.Logger.Log(". Starting entry sender")

	dataSender := sender.Sender{
		ApiUrl: configuration.SenderSettings.Url,
	}

	go func(configuration src.Configuration, dataSender sender.Sender) {

		ticker := time.NewTicker(
			time.Duration(configuration.SenderSettings.CycleFrequency) * time.Second,
		)
		for {
			// Parses unsent entries
			entries := storage.GetUnsentEntries(configuration.SenderSettings.EntriesPerCycle)

			if len(entries) != 0 {
				dataSender.SendEntries(entries)

			}

			// time.Sleep(time.Duration(cycleFrequency) * time.Second)
			<-ticker.C
		}
	}(configuration, dataSender)
}

// Processes all piled up output entries from a single processor to be saved
//
// TODO: Move this somewhere else
func processProcessorToOutputEntry(processor *src.Processor) []storage.OutputEntry {
	numOfProcessorOutputs := len(processor.Outputs)

	var outputEntries []storage.OutputEntry

	if numOfProcessorOutputs == 0 {
		return outputEntries
	}

	i := 0
	for i < numOfProcessorOutputs {
		outputEntries = append(
			outputEntries,
			storage.OutputEntry{
				HandlerIdentifier: processor.Handler.Identifier,
				CommandName:       processor.Command.Name,
				Output:            processor.Outputs[i].Output,
				Timestamp:         processor.Outputs[i].Timestamp,
			},
		)
		i = i + 1
	}
	// If the array has not changed, empty slice. Otherwise, slice off the entries processed
	if len(processor.Outputs) == numOfProcessorOutputs {
		processor.Outputs = processor.Outputs[0:0]
	} else {
		processor.Outputs = processor.Outputs[numOfProcessorOutputs-1:]
	}

	return outputEntries
}
