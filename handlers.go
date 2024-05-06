package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// A map of command handlers for interactions.
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"add_cards": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Getting all the files in the directory.
		filesList, err := os.ReadDir("./cards")
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT LIST CARDS: %v", Red, Reset, err)
			return
		}

		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Now registering %d cards...", len(filesList)),
			},
		})
		if err != nil {
			return
		}

		for _, file := range filesList {
			// Grabbing the image file path.
			filePath := fmt.Sprintf("./cards/%v", file.Name())

			// Reading the file into memory.
			imageBytes, err := os.Open(filePath)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT READ IMAGE: %v", Red, Reset, err)
				return
			}

			// Uploading that image to discord for saving.
			msg, err := session.ChannelFileSend(interaction.ChannelID, file.Name(), imageBytes)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPLOAD IMAGE: %v", Red, Reset, err)
				return
			}

			// Getting all the variables for the cards.
			name := strings.ReplaceAll(file.Name(), ".png", "")
			nameParts := strings.Split(name, " ")
			log.Println(nameParts)

			var character string
			switch nameParts[0] {
			case "SG01":
				character = "Hibiki"
			case "SG02":
				character = "Tsubasa"
			case "SG03":
				character = "Chris"
			case "SG04":
				character = "Maria"
			case "SG05":
				character = "Shirabe"
			case "SG06":
				character = "Kirika"
			case "SG07":
				character = "Kanade"
			case "SG08":
				character = "Miku"
			case "SG09":
				character = "Serena"
			case "SG10":
				character = "Fine"
			case "SG11":
				character = "Dr. Ver"
			case "SG12":
				character = "Genjuro"
			case "SG13":
				character = "Ogawa"
			case "SG14":
				character = "St. Germain"
			case "SG15":
				character = "Cagliostro"
			case "SG16":
				character = "Prelati"
			case "SG18":
				character = "Adam"
			case "SG19":
				character = "Carol"
			case "SG21":
				character = "Phara"
			case "SG22":
				character = "Leiur"
			case "SG23":
				character = "Garie"
			case "SG24":
				character = "Micha"
			case "SG25":
				character = "Andou"
			case "SG26":
				character = "Shiori"
			case "SG27":
				character = "Yumi"
			case "SG28":
				character = "Vanessa"
			case "SG29":
				character = "Millaarc"
			case "SG30":
				character = "Elsa"
			case "SG31":
				character = "Shem-Ha"
			case "SG32":
				character = "Elfnein"
			}

			cardID := fmt.Sprintf("%v_%v", nameParts[0], nameParts[1])
			evolution := nameParts[2]
			cardImage := msg.Attachments[0].URL

			// Creating a query to insert the cards into the card database.
			query := fmt.Sprintf(`INSERT INTO cards(character, id, evolution, image) VALUES("%v", "%v", %v, "%v");`,
				character, cardID, evolution, cardImage)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT REGISTER CARD IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - REGISTERED CARD IN CARD DATABASE: %v", Green, Reset, result)

			time.Sleep(time.Millisecond * 10)
		}
	},
	"register": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID

		if userIsRegistered(session, interaction) {
			// Notify the user that they are already registerd.
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You're already registered to play!",
				},
			})
			if err != nil {
				return
			}
		} else {
			// Creating a query to insert the user into the database with a phony unix timestamp and no credits.
			query := fmt.Sprintf(`INSERT INTO users(id, timestamp, credits) VALUES("%s", 0, 10000);`,
				authorID)

			// Executing that query.
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT PLACE USER IN REGISTRATION DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - PLACED USER INTO REGISTRATION DATABASE: %v", Green, Reset, result)

			// Notify the user that they are now registered.
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You are now registered to play. Here's 10,000 credits to get you started!",
				},
			})
			if err != nil {
				return
			}
		}
	},
	"daily": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		currentTimestamp := time.Now().Unix()
		var databaseTimestamp int64
		var credits int64

		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// Perform a single row query in the database to retrieve the timestamp.
		query := fmt.Sprintf(`SELECT timestamp FROM users WHERE id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&databaseTimestamp)
		if err != nil {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE USER'S TIMESTAMP FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// Checking to see if the user is on cooldown or if it is just an outdated entry.
		if currentTimestamp >= databaseTimestamp+int64(86400) {
			// It was an outdated entry, so we should give the user their reward and place them on cooldown again.

			// Updating the timestamp in the database so that the user can't use the command again for a certain amount of time.
			query = fmt.Sprintf(`UPDATE users SET timestamp = %v WHERE id = %v;`, currentTimestamp, authorID)
			result, err := db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE UNIX TIMESTAMP IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER COOLDOWN: %v", Green, Reset, result)

			// Snagging the amount of credits so that they can be updated.
			query := fmt.Sprintf(`SELECT credits FROM users WHERE id = %v;`, authorID)
			err = db.QueryRow(query).Scan(&credits)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT GET CREDITS OF USER IN DATABASE: %v", Red, Reset, err)
				return
			}

			// Updating the amount of credits in the database for the user.
			query = fmt.Sprintf(`UPDATE users SET credits = %v WHERE id = %v;`, credits+int64(100), authorID)
			result, err = db.Exec(query)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT UPDATE CREDITS IN DATABASE: %v", Red, Reset, err)
				return
			}
			log.Printf("%vSUCCESS%v - UPDATED USER CREDITS: %v", Green, Reset, result)

			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Here's your daily reward of 100 credits!",
				},
			})
			if err != nil {
				return
			}
		} else {
			// The user is actually on cooldown so we should let them know to comeback later.
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Come back on <t:%v:D> at <t:%v:T> to claim your daily reward!",
						databaseTimestamp+int64(86400), databaseTimestamp+int64(86400)),
				},
			})
			if err != nil {
				return
			}
		}

	},
	"credits": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var credits int64
		authorID := interaction.Member.User.ID
		printer := message.NewPrinter(language.English)

		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// Perform a single row query to get the amount of credits a user has.
		query := fmt.Sprintf(`SELECT credits FROM users WHERE id = %s;`, authorID)
		err := db.QueryRow(query).Scan(&credits)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE CREDITs FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: printer.Sprintf("You currently have %d credits!", credits),
			},
		})
		if err != nil {
			return
		}
	},
	"characters": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: strings.Join(charactersList(), ", "),
			},
		})
		if err != nil {
			return
		}
	},
	"collection_list": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID
		var id string
		var characterName string
		var customName string
		var evolution int8
		var query string
		// var webhookParams []discordgo.WebhookParams
		var embeds []*discordgo.MessageEmbed

		// Checking to make sure the user is registered.
		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// I don't even know at this point. Check whether a character is specified or something.
		if len(interaction.ApplicationCommandData().Options) == 0 {
			query = fmt.Sprintf(`SELECT card_id, character, custom_name, evolution FROM collections WHERE user_id = %v;`, authorID)
		} else {
			if inArray(strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()), charactersList()) {
				query = fmt.Sprintf(`SELECT card_id, character, custom_name, evolution FROM collections WHERE user_id = %v AND character_name = "%v";`,
					authorID, interaction.ApplicationCommandData().Options[0].StringValue())
			} else {
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "I couldn't find a character with that name.",
					},
				})
				if err != nil {
					return
				}
			}

		}

		// Executing the query.
		rows, err := db.Query(query)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE CARDS FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// Creating a struct to hold query results.
		type Card struct {
			id            string
			characterName string
			customName    string
			evolution     int8
		}

		// Making an slice of card structs to hold results.
		var cards []Card

		// Iterating over the results and appending to an array of cards.
		for rows.Next() {
			err := rows.Scan(&id, &characterName, &customName, &evolution)
			if err != nil {
				log.Printf("%vERROR%v - COULD NOT RETRIEVE CHARACTER FROM ROW:\n\t%v", Red, Reset, err)
				return
			}

			var card Card
			card.id = id
			card.characterName = characterName
			card.customName = customName
			card.evolution = evolution

			cards = append(cards, card)
		}

		if len(cards) == 0 {
			// If there were no rows returned, let the user know that they don't have any cards.
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "It looks like you don't have any cards that match that criteria!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// Sorting the slice.
		sort.SliceStable(cards[:], func(i, j int) bool {
			return cards[i].characterName < cards[j].characterName
		})

		// SUPER funky shit to chop up array.
		var chunkedCards [][]Card
		chunkSize := 25

		for i := 0; i < len(cards); i += chunkSize {
			end := i + chunkSize

			if end > len(cards) {
				end = len(cards)
			}

			chunkedCards = append(chunkedCards, cards[i:end])
		}

		// Printing the results to the user. Need to clean it up...
		for _, values := range chunkedCards {
			buffer := new(bytes.Buffer)
			writer := tabwriter.NewWriter(buffer, 0, 0, 4, ' ', 0)
			fmt.Fprintln(writer, "Character:\tCard Name:\tEvolution:")

			for _, value := range values {
				// content += fmt.Sprintf("%-10s\t%12s\n", value.characterName, value.customName)
				_, err := fmt.Fprintf(writer, "%v\t%v\t%v\n", value.characterName, value.customName, value.evolution)
				if err != nil {
					log.Println(err)
				}
			}

			writer.Flush()

			content := "```" + buffer.String() + "```"

			embeds = append(embeds, &discordgo.MessageEmbed{
				Description: content,
			})
		}

		// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Now listing cards...",
			},
		})
		if err != nil {
			return
		}

		paginator := dgwidgets.NewPaginator(session, interaction.ChannelID)

		for _, embed := range embeds {
			paginator.Add(embed)
		}

		paginator.SetPageFooters()

		paginator.Widget.Timeout = time.Minute * 5

		paginator.Widget.UserWhitelist = append(paginator.Widget.UserWhitelist, authorID)

		paginator.Spawn()
	},
	"collection_amount": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var query string
		var userAmount int64
		var totalAmount int64

		authorID := interaction.Member.User.ID

		printer := message.NewPrinter(language.English)

		// Checking to make sure the user is registered.
		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// I don't even know at this point. Check whether a character is specified or something.
		if len(interaction.ApplicationCommandData().Options) == 0 {
			query = fmt.Sprintf(`SELECT COUNT(DISTINCT card_id) FROM collections WHERE user_id = %v;`, authorID)
		} else {
			if inArray(strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()), charactersList()) {
				query = fmt.Sprintf(`SELECT COUNT(DISTINCT card_id) FROM collections WHERE user_id = %v AND character = "%v";`,
					authorID, strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()))
			} else {
				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "I couldn't find a character with that name.",
					},
				})
				if err != nil {
					return
				}
			}
		}

		err := db.QueryRow(query).Scan(&userAmount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE AMOUNT OF CARDS FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		// Grabbing the amount of those total cards in the cards table.
		if len(interaction.ApplicationCommandData().Options) == 0 {
			query = `SELECT COUNT(DISTINCT card_id) FROM cards;`
		} else {
			query = fmt.Sprintf(`SELECT COUNT(DISTINCT card_id) FROM cards WHERE character = "%v";`,
				interaction.ApplicationCommandData().Options[0].StringValue())
		}

		err = db.QueryRow(query).Scan(&totalAmount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("%vERROR%v - COULD NOT RETRIEVE AMOUNT OF CARDS FROM DATABASE:\n\t%v", Red, Reset, err)
			return
		}

		if len(interaction.ApplicationCommandData().Options) == 0 {
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: printer.Sprintf("You have currently collected %d out of %d cards!", userAmount, totalAmount),
				},
			})
			if err != nil {
				return
			}
		} else {
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: printer.Sprintf("You have currently collected %d out of %d cards of %s!",
						userAmount, totalAmount, strings.Title(interaction.ApplicationCommandData().Options[0].StringValue())),
				},
			})
			if err != nil {
				return
			}
		}
	},
	"single_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID

		// Checking to make sure the user is registered.
		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// Snagging the amount of credits so that they can be checked against.
		credits := getCredits(authorID)

		// Making sure the user has the correct amount of credits.
		if credits >= int64(200) {
			// Checking to see if the user specified a character to pull for.
			if len(interaction.ApplicationCommandData().Options) != 0 {
				// They did, so we need to check if that that character is available to pull from.
				if inArray(strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()), charactersList()) {
					// Updating the amount of credits in the database for the user.
					updateCredits(-200, authorID)

					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: "I've deducted 200 credits from your wallet, let's see what you drew!",
						},
					})
					if err != nil {
						return
					}

					time.Sleep(time.Second / 10)

					webhook := pullCard(session, interaction)
					_, err = session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
					if err != nil {
						return
					}
				} else {
					// Could not find a character pool with that name.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: "I couldn't find a character pool with that name.",
						},
					})
					if err != nil {
						return
					}
				}
			} else {
				// The player did not specify a character to draw from.
				// Updating the amount of credits in the database for the user.
				updateCredits(-200, authorID)

				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
						Content: "I've deducted 200 credits from your wallet, let's see what you drew!",
					},
				})
				if err != nil {
					return
				}

				//time.Sleep(time.Second / 10)

				webhook := pullCard(session, interaction)
				_, err = session.FollowupMessageCreate(interaction.Interaction, true, &webhook)
				if err != nil {
					return
				}
			}
		} else {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough credits to draw a card.",
				},
			})
			if err != nil {
				return
			}
		}
	},
	"ten_pull": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		authorID := interaction.Member.User.ID

		// Checking to make sure the user is registered.
		if !userIsRegistered(session, interaction) {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey! You aren't registered to play yet! Remember to use the command `/register` before trying to play!",
				},
			})
			if err != nil {
				return
			}

			return
		}

		// Snagging the amount of credits so they can be checked against.
		credits := getCredits(authorID)

		// Making sure the user has the correct amount of credits.
		if credits >= int64(1800) {
			// Checking to see if the user specified a character to pull for.
			if len(interaction.ApplicationCommandData().Options) != 0 {
				// They did, so we need to check if that that character is available to pull from.
				if inArray(strings.Title(interaction.ApplicationCommandData().Options[0].StringValue()), charactersList()) {
					// Updating the amount of credits in the database for the user.
					updateCredits(-1800, authorID)

					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: "I've deducted 1,800 credits from your wallet, let's see what you drew!",
						},
					})
					if err != nil {
						return
					}

					// Conducting the ten pull.
					var webhookParams []discordgo.WebhookParams

					for i := 0; i < 10; i++ {
						webhook := pullCard(session, interaction)
						webhookParams = append(webhookParams, webhook)
					}

					paginator := dgwidgets.NewPaginator(session, interaction.ChannelID)

					for _, webhook := range webhookParams {
						paginator.Add(webhook.Embeds[0])
					}

					paginator.SetPageFooters()

					paginator.Widget.Timeout = time.Minute * 5

					paginator.Widget.UserWhitelist = append(paginator.Widget.UserWhitelist, authorID)
					err = paginator.Spawn()
					if err != nil {
						return
					}
				} else {
					// Could not find a character pool with that name.
					// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
					err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
							Content: "I couldn't find a character pool with that name.",
						},
					})
					if err != nil {
						return
					}
				}
			} else {
				// The user did not specify a character to pull for.
				// Updating the amount of credits in the database for the user.
				updateCredits(-1800, authorID)

				// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
				err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						//Content: fmt.Sprintf("Successfully added %v to your collection. You can rename this card at anytime by using `/rename [original_name] [new_name]", drawnCardID),
						Content: "I've deducted 1,800 credits from your wallet, let's see what you drew!",
					},
				})
				if err != nil {
					return
				}

				// Conducting the ten pull.
				var webhookParams []discordgo.WebhookParams

				for i := 0; i < 10; i++ {
					webhook := pullCard(session, interaction)
					webhookParams = append(webhookParams, webhook)
				}

				paginator := dgwidgets.NewPaginator(session, interaction.ChannelID)

				for _, webhook := range webhookParams {
					paginator.Add(webhook.Embeds[0])
				}

				paginator.SetPageFooters()

				paginator.Widget.Timeout = time.Minute * 5

				paginator.Widget.UserWhitelist = append(paginator.Widget.UserWhitelist, authorID)

				err = paginator.Spawn()
				if err != nil {
					return
				}
			}
		} else {
			// https: //pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough credits to draw a card.",
				},
			})
			if err != nil {
				return
			}
		}
	},
}
