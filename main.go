package main

import (
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type record struct {
	reminder string
	loc      *time.Location
	timer    int64
	chat_id  tgbotapi.MessageConfig
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("7212937598:AAFbYryUxv8iw0fWaOEgRd-te-WJpOIKaF8"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	loc, _ := time.LoadLocation("Local")
	for update := range updates {

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		if !update.Message.IsCommand() { // ignore any non-command Messages
			msg.Text = "If you do not know how to use this cursed abomination, use /help command"
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "To initiate timer use /new_timer"
		case "sayhi":
			msg.Text = "Hi :)"
		case "status":
			msg.Text = "I'm ok."
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func new_timer(update tgbotapi.Update, msg tgbotapi.MessageConfig, u tgbotapi.UpdateConfig, updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) (record, bool) {
	loc, _ := time.LoadLocation("Local")
	var cancel bool
	cancel = false
	var tmp record
	msg.Text = "what do i need to remind you about? If you wish to cancel, write 'cancel'"
	bot.Send(msg)
	update = <-updates
	tmp.chat_id = msg
	var correct bool
	correct = false
	var message string
	for !correct {
		if !update.Message.IsCommand() && update.Message != nil {
			if strings.TrimSpace(update.Message.Text) == "cancel" {
				cancel = true
				return tmp, cancel
			} else {
				message = update.Message.Text
				correct = true
			}
		} else {
			msg.Text = "Chief, i can't take stickers or whatever you send as reminder. Just plain text, please"
			bot.Send(msg)
			update = <-updates
		}
	}
	tmp.reminder = message
	msg.Text = "Next, date of reminder- when do i need to send you a reminder? Please, write it in format dd:mm:yyyy hh:mm:ss, unless you want to break this bot"
	correct = false
	update = <-updates
	for !correct {
		if !update.Message.IsCommand() || update.Message == nil {
			msg.Text = "Incorrect format"
			bot.Send(msg)
			update = <-updates
		} else {
			if strings.TrimSpace(update.Message.Text) == "cancel" {
				cancel = true
				return tmp, cancel
			} else {
				message = update.Message.Text
				message = strings.TrimSpace(message)
				remind_date, _ := time.ParseInLocation("02:01:2006 15:04:05", message, loc)
				if remind_date.Year() == 0001 {
					msg.Text = "You wrote in incorrect format, please write date and time in format dd:mm:yyyy hh:mm:ss"
					bot.Send(msg)
					update = <-updates
				} else {
					tmp.timer = remind_date.Unix()
					correct = true
				}
			}
		}
	}
	tmp.chat_id = msg
	tmp.loc = loc
	return tmp, cancel
}
