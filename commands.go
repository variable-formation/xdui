package main

import (
	"github.com/bwmarrin/discordgo"
)

// Permissions for commands.
var manageServerPermission int64 = discordgo.PermissionManageServer

var dmPermission bool = false

// The list of commands for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "add_cards",
		Description:              "Loops through './Card Art' folder and registers all the cards in there.",
		DefaultMemberPermissions: &manageServerPermission,
		DMPermission:             &dmPermission,
	},
	{
		Name:         "register",
		Description:  "This command registers you to play!",
		DMPermission: &dmPermission,
	},
	{
		Name:         "daily",
		Description:  "This command claims your daily credits!",
		DMPermission: &dmPermission,
	},
	{
		Name:         "credits",
		Description:  "This command tells you how many credits you have.",
		DMPermission: &dmPermission,
	},
	{
		Name:         "characters",
		Description:  "This command lists the available gacha pools to pull from.",
		DMPermission: &dmPermission,
	},
	{
		Name:         "collection_list",
		Description:  "This command lists the cards in your collection.",
		DMPermission: &dmPermission,

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to list.",
				Required:    false,
			},
		},
	},
	{
		Name:         "collection_amount",
		Description:  "This command lists the amount of the cards in your collection.",
		DMPermission: &dmPermission,

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to list.",
				Required:    false,
			},
		},
	},
	{
		Name:         "single_pull",
		Description:  "This command pulls one random card from the gacha pool.",
		DMPermission: &dmPermission,

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to draw for.",
				Required:    false,
			},
		},
	},
	{
		Name:         "ten_pull",
		Description:  "This command pulls ten random cards from the gacha pool.",
		DMPermission: &dmPermission,

		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "The name of the character you wish to draw for.",
				Required:    false,
			},
		},
	},
}
