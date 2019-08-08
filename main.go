package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	token := os.Getenv("TELEGRAM_APITOKEN")
	if len(token) == 0 {
		log.Fatal("main: environment variable TELEGRAM_APITOKEN is empty")
	}

	quotesDB, err := Parse()
	if err != nil {
		log.Fatalf("main: %s", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("main: %s", err)
	}

	log.Printf("main: authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			command := update.Message.Command()
			switch command {
			case "help":
				res := strings.Builder{}
				res.WriteString("You can get wise quote by sending command to this bot.\nList of commands:\n")
				res.WriteString("/get - get completely random quote\n")
				for k, v := range quotesDB.GetThemes() {
					res.WriteString(fmt.Sprintf("/%s - get random quote from category \"%s\"\n", k, v))
					if res.Len() > 3000 {
						msg.Text = res.String()
						bot.Send(msg)
						res.Reset()
					}
				}
				res.WriteString("\n/help - show this message\n")
				msg.Text = res.String()
			case "start":
				msg.Text = fmt.Sprintf("Greetings %s.\n Try /get to get completely random quote or /help to get list of all commands", update.Message.From.UserName)
			case "stop":
				msg.Text = fmt.Sprintf("Goodbye %s. I hope you was satisfied with wise quotes", update.Message.From.UserName)
			case "get":
				msg.Text = quotesDB.GetRandomQuote()
			default:
				quote, err := quotesDB.GetRandomQuoteByTheme(command)
				if err != nil {
					msg.Text = "Unfortunately, I don`t know this command, maybe you should check out our /help ?"
				} else {
					msg.Text = quote
				}
			}
			bot.Send(msg)
		}

	}

}
