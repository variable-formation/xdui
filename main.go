package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

// Creating a struct to hold the Discord token.
type Config struct {
	Discord_Token  string
	MySQL_Username string
	MySQL_Password string
	MySQL_Database string
}

// Creating a variable to hold the Config struct.
var config Config

// Global variable to hold database connection, because why not?
var db *sql.DB

// Global variable to hold regex string.
var re *regexp.Regexp

func main() {
	log.Printf("%vBOT IS STARTING UP.%v", Blue, Reset)

	// Retrieve the tokens from the tokens.json file.
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("%vERROR%v - COULD NOT READ 'config.json' FILE:\n\t%v", Red, Reset, err)
	}

	// Unmarshal the tokens from tokensFile.
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return
	}

	// Open a connection to the database.
	db, err = sql.Open("sqlite3", "cards.db")
	if err != nil {
		log.Fatalf("%vERROR%v - COULD NOT CONNECT TO DATABASE:\n\t%v", Red, Reset, err)
	}

	// Compile regex string.
	re, err = regexp.Compile(`^[A-Za-z0-9 _]*[A-Za-z0-9][A-Za-z0-9 _]*$`)
	if err != nil {
		log.Fatal("COULD NOT COMPILE REGEX: ", err)
	}

	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + config.Discord_Token)
	if err != nil {
		log.Fatalf("%vERROR%v - PROBLEM CREATING DISCORD SESSION:\n\t%v", Red, Reset, err)
	}

	// Identify that we want all intents.
	session.Identify.Intents = discordgo.IntentsAll

	// Now we open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatalf("%vERROR%v - PROBLEM OPENING WEBSOCKET:\n\t%v", Red, Reset, err)
	}

	log.Println("Registering commands...")
	// Making a map of registered commands.
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	// Looping through the commands array and registering them.
	// https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.ApplicationCommandCreate
	for i, command := range commands {
		registeredCommand, err := session.ApplicationCommandCreate(session.State.User.ID, "990405675022700564", command)
		if err != nil {
			log.Printf("CANNOT CREATE '%v' COMMAND: %v", command.Name, err)
		}
		registeredCommands[i] = registeredCommand
	}

	// Looping through the array of interaction handlers and adding them to the session.
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})

	// Wait here until CTRL-C or other term signal is received.
	log.Printf("%vBOT IS NOW RUNNING.%v", Blue, Reset)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Lopping through the registeredCommands array and deleting all the commands.
	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "990405675022700564", v.ID)
		if err != nil {
			log.Printf("CANNOT DELETE '%v' COMMAND: %v", v.Name, err)
		}
	}

	// Cleanly close down the Discord session.
	err = session.Close()
	if err != nil {
		return
	}
	fmt.Println("\nHave a good day!")
}
