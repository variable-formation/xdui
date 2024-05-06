package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// Function to get all the available characters to pull from.
func charactersList() []string {
	var character string
	var characters []string
	// Creating a query to get distinct character names from the cards table.
	query := `SELECT DISTINCT character FROM cards;`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("%vERROR%v - COULD NOT RETRIEVE CHARACTERS FROM DATABASE:\n\t%v", Red, Reset, err)
	}

	for rows.Next() {
		err := rows.Scan(&character)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE CHARACTER FROM ROW:\n\t%v", Red, Reset, err)
		}

		characters = append(characters, character)
	}

	return characters
}

// Function to check and see if the user is register.
func userIsRegistered(session *discordgo.Session, interaction *discordgo.InteractionCreate) bool {
	var id int64
	authorID := interaction.Member.User.ID

	// Perform a single row query to check if the user is registered.
	query := fmt.Sprintf(`SELECT id FROM users WHERE id = %s;`, authorID)
	err := db.QueryRow(query).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
		return false
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false
	} else {
		return true
	}
}

func userIsRegisteredByID(userID string) bool {
	var id int64

	// Perform a single row query to check if the user is registered.
	query := fmt.Sprintf(`SELECT id FROM users WHERE id = %s;`, userID)
	err := db.QueryRow(query).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("%vERROR%v - COULD NOT RETRIEVE USER FROM REGISTRATION DATABASE:\n\t%v", Red, Reset, err)
		return false
	}

	if err == sql.ErrNoRows {
		return false
	} else {
		return true
	}
}

// Function to get the credits of a user.
func getCredits(authorID string) int64 {
	var credits int64
	// Snagging the amount of credits so that they can be checked against.
	query := fmt.Sprintf(`SELECT credits FROM users WHERE id = %v;`, authorID)
	err := db.QueryRow(query).Scan(&credits)
	if err != nil {
		log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
	}
	log.Printf("%vSUCCESS%v - GRABBED USER CREDITS", Green, Reset)

	return credits
}

// Function to update the credits of a user.
func updateCredits(amount int64, authorID string) {
	// Updating the amount of credits in the database for the user.
	query := fmt.Sprintf(`UPDATE users SET credits = %v WHERE id = %v;`,
		getCredits(authorID)+int64(amount), authorID)
	result, err := db.Exec(query)
	if err != nil {
		log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
		return
	}
	log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)
}

// Function that pulls a card from the card database and adds it to the user database. Does not handle cost.
func pullCard(session *discordgo.Session, interaction *discordgo.InteractionCreate) discordgo.WebhookParams {
	var drawnCardID string
	var characterName string
	var evolution int8
	var customName string
	var cardImage string
	var credits int64

	authorID := interaction.Member.User.ID

	// Performing a single row query to draw a card with optional character.
	var query string
	if len(interaction.ApplicationCommandData().Options) == 0 {
		query = `SELECT card_id, character, image FROM cards WHERE evolution = 1 ORDER BY RANDOM() LIMIT 1;`
	} else {
		query = fmt.Sprintf(`SELECT card_id, character, image FROM cards WHERE evolution = 1 AND character = "%v" ORDER BY  RANDOM() LIMIT 1;`,
			strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()))
	}

	err := db.QueryRow(query).Scan(&drawnCardID, &characterName, &cardImage)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("%vERROR%v - COULD NOT GET A CARD TO DRAW FROM DATABASE: %v", Red, Reset, err)
	}

	// Performing a single row query to check if the user already has the card in their collection.
	query = fmt.Sprintf(`SELECT card_id, evolution FROM collections WHERE card_id = "%v" AND user_id = %v;`, drawnCardID, authorID)
	err = db.QueryRow(query).Scan(&drawnCardID, &evolution)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("%vERROR%v - COULD NOT QUERY USER COLLECTION IN DATABASE: %v", Red, Reset, err)
	}

	// Checking if the user has the card in their collection or not.
	if err == sql.ErrNoRows {
		// The user does not have the card in their collection.
		// Query to insert the card into the user's collection.
		query = fmt.Sprintf(`INSERT INTO collections(user_id, character, card_id, evolution, custom_name) VALUES("%v", "%v", "%v", "%v", "%v");`,
			authorID, characterName, drawnCardID, 1, drawnCardID)

		// Executing that query.
		result, err := db.Exec(query)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT PLACE CARD IN USER COLLECTION DATABASE: %v", Red, Reset, err)
		}
		log.Printf("%vSUCCESS%v - PLACED CARD INTO USER COLLECTION DATABASE: %v", Green, Reset, result)

		// Constructing an embed to hold the card image.
		image := discordgo.MessageEmbedImage{
			URL: cardImage,
		}
		embeds := []*discordgo.MessageEmbed{
			{
				Image:       &image,
				Description: fmt.Sprintf("Successfully added %v to your collection.", drawnCardID),
			},
		}

		// Returning the webhook.
		return discordgo.WebhookParams{
			Embeds: embeds,
		}
	} else {
		// The user does have this card, the only question is what level do they have?
		if evolution == 3 {
			// If the evolution level is the max level, then we need to refund the user for this draw.
			// Updating the amount of credits in the database for the user.

			// Snagging the amount of credits so that they can be updated.
			query := fmt.Sprintf(`SELECT credits FROM users WHERE id = %v;`, authorID)
			err := db.QueryRow(query).Scan(&credits)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
			}

			query = fmt.Sprintf(`UPDATE users SET credits = %v WHERE id = %v;`, credits+int64(100), authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// Creating a query to retrieve the image of the maximum level card.
			query = fmt.Sprintf(`SELECT image FROM cards where card_id = "%v" AND evolution = %v`, drawnCardID, evolution)
			err = db.QueryRow(query).Scan(&cardImage)
			if err != nil {
				log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
			}

			// Creating a query to retrieve the user's custom name of the card.
			query = fmt.Sprintf(`SELECT custom_name FROM collections WHERE user_id = %v AND card_id = "%v"`, authorID, drawnCardID)
			err = db.QueryRow(query).Scan(&customName)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
			}

			// Constructing an embed to hold the card image.
			image := discordgo.MessageEmbedImage{
				URL: cardImage,
			}
			embeds := []*discordgo.MessageEmbed{
				{
					Image:       &image,
					Description: fmt.Sprintf("Whoah there! You've already maxed out your %v, I've refunded half a draw for you. Go ahead and try again!", customName),
				},
			}

			// Returning the webhook.
			return discordgo.WebhookParams{
				Embeds: embeds,
			}
		} else if evolution == 2 {
			// If the evolution level is 2, then we need to evolve the card to level 3.

			// Updating the evolution level in the database of the user's card.
			query = fmt.Sprintf(`UPDATE collections SET evolution = 3 WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE USER EVOLUTION IN DATABASE: %v", Red, Reset, err)
			}
			log.Printf("%vSUCCESS%v - UPDATED USER EVOLUTION: %v", Green, Reset, result)

			// Creating a query to retrieve the image of the maximum level card.
			query = fmt.Sprintf(`SELECT image FROM cards WHERE card_id = "%v" AND evolution = 3;`, drawnCardID)
			err = db.QueryRow(query).Scan(&cardImage)
			if err != nil {
				log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
			}

			// Creating a query to retrieve the user's custom name of the card.
			query = fmt.Sprintf(`SELECT custom_name FROM collections WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
			err = db.QueryRow(query).Scan(&customName)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
			}

			// Constructing an embed to hold the card image.
			image := discordgo.MessageEmbedImage{
				URL: cardImage,
			}
			embeds := []*discordgo.MessageEmbed{
				{
					Image:       &image,
					Description: fmt.Sprintf("Check it out! You've evolved your %v!", customName),
				},
			}

			// Returning the webhook.
			return discordgo.WebhookParams{
				Embeds: embeds,
			}
		} else if evolution == 1 {
			// If the evolution level is 1, then we need to evolve the card to level 2.

			// Updating the evolution level in the database of the user's card.
			query = fmt.Sprintf(`UPDATE collections SET evolution = 2 WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE USER EVOLUTION IN DATABASE: %v", Red, Reset, err)
			}
			log.Printf("%vSUCCESS%v - UPDATED USER EVOLUTION: %v", Green, Reset, result)

			// Creating a query to retrieve the image of the maximum level card.
			query = fmt.Sprintf(`SELECT image FROM cards where card_id = "%v" AND evolution = 2;`, drawnCardID)
			err = db.QueryRow(query).Scan(&cardImage)
			if err != nil {
				log.Printf("%vERROR%v - COULD QUERY CARDS IN DATABASE: %v", Red, Reset, err)
			}

			// Creating a query to retrieve the user's custom name of the card.
			query = fmt.Sprintf(`SELECT custom_name FROM collections WHERE user_id = %v AND card_id = "%v";`, authorID, drawnCardID)
			err = db.QueryRow(query).Scan(&customName)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CUSTOM NAME FROM DATABASE: %v", Red, Reset, err)
			}

			// Constructing an embed to hold the card image.
			image := discordgo.MessageEmbedImage{
				URL: cardImage,
			}
			embeds := []*discordgo.MessageEmbed{
				{
					Image:       &image,
					Description: fmt.Sprintf("Check it out! You've evolved your %v!", customName),
				},
			}

			// Returning the webhook.
			return discordgo.WebhookParams{
				Embeds: embeds,
			}
		}
	}

	return discordgo.WebhookParams{}
}

// Function to check if element is in array.
func inArray(str string, arr []string) bool {
	for _, value := range arr {
		if value == str {
			return true
		}
	}

	return false
}

// Function that converts a bool to an int. I don't know what else to say.
func boolToInt(truth bool) int {
	if truth {
		return 1
	} else {
		return 0
	}
}

// Function to return the integer that is less than the other.
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
