package main

import (
	"../src"
	"../src/logger"
	"../src/sender"
	"../src/startup"
	"../src/storage"
	// "sync"
	"fmt"
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
	configurationFile := "configuration.json"

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

	outputEntryChannel := make(chan storage.OutputEntry)
	startProcessors(configuration, outputEntryChannel)

	startStorageListener(outputEntryChannel)

	startEntrySender(configuration)

	startEntryCleanup(configuration)

	// Exit the main thread. Application keeps on running.
	runtime.Goexit()

}

// Start running the processors.
//
// This will create a separate goroutine for each processor.
func startProcessors(configuration src.Configuration, outputEntryChannel chan storage.OutputEntry) {

	// Loop and run all the processors
	for _, processor := range configuration.RuntimeData.Processors {
		// Add some time between starting to make log reading easier
		time.Sleep(time.Second * 5)

		// Run processor in a new goroutine
		logger.Logger.Log(".. Running " + processor.Handler.Name)
		go processor.Run(outputEntryChannel)
	}
}

// Start the storage listener - will listen to channel messages and store
// data to the database.
func startStorageListener(outputEntryChannel chan storage.OutputEntry) {
	go func(outputEntryChannel chan storage.OutputEntry) {
		var err error
		var outputEntries []storage.OutputEntry
		storeAmount := 1
		for {
			outputEntry := <-outputEntryChannel
			outputEntries = append(outputEntries, outputEntry)
			if len(outputEntries) >= storeAmount {
				err = storage.StoreOutputEntries(outputEntries)

				if err != nil {
					src.ExitApplicationWithMessage(
						fmt.Sprintf("Error storing entries: %s", err),
					)
				}

				outputEntries = outputEntries[0:0]

			}
		}
	}(outputEntryChannel)
}

// Starts the entry sender - send all flagged entries to a remote API.
func startEntrySender(configuration src.Configuration) {
	if !configuration.SendData || configuration.SenderSettings.Url == "" {
		return
	}

	logger.Logger.Log(". Starting entry sender")

	dataSender := sender.Sender{
		ApiUrl:         configuration.SenderSettings.Url,
		ApplicationKey: configuration.SenderSettings.ApplicationKey,
		ServerHandler:  configuration.SenderSettings.ServerHandler,
	}

	go func(configuration src.Configuration, dataSender sender.Sender) {

		ticker := time.NewTicker(
			time.Duration(configuration.SenderSettings.CycleFrequency) * time.Second,
		)
		for {
			// Parses unsent entries
			entries, err := storage.GetUnsentEntries(configuration.SenderSettings.EntriesPerCycle)

			if err != nil {
				src.ExitApplicationWithMessage(
					fmt.Sprintf("Error getting unsent entries: %s", err),
				)
			}

			logger.Logger.Log(fmt.Sprintf("Sending entries. Count: %d", len(entries)))

			result := false

			if len(entries) != 0 {
				result = dataSender.SendEntries(entries)
			}

			if result {
				storage.MarkEntriesSent(entries)
			}

			// time.Sleep(time.Duration(cycleFrequency) * time.Second)
			<-ticker.C
		}
	}(configuration, dataSender)
}

// Starts entry cleanup. This targets old entries and removes them from the db
func startEntryCleanup(configuration src.Configuration) {
	logger.Logger.Log(". Starting entry cleanup")

	go func(configuration src.Configuration) {

		cleanupFrequency := 120
		cleanupOldestEntry := 259200

		if configuration.CleanupFrequency != 0 {
			cleanupFrequency = configuration.CleanupFrequency
		}
		if configuration.CleanupOldestEntry != 0 {
			cleanupOldestEntry = configuration.CleanupOldestEntry
		}

		ticker := time.NewTicker(
			time.Duration(cleanupFrequency) * time.Second,
		)

		for {

			// Deletes entries older than the defined oldest entry
			err := storage.DeleteOldEntries(cleanupOldestEntry)

			if err != nil {
				src.ExitApplicationWithMessage(
					fmt.Sprintf("Error getting unsent entries: %s", err),
				)
			}

			<-ticker.C
		}
	}(configuration)

}
