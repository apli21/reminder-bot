package main

import (
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type user_data struct {
	status   int
	chat_id  int64
	reminder string
	loc      *time.Location
	timer    int64
}

type record struct {
	reminder string
	loc      *time.Location
	timer    int64
	chat_id  tgbotapi.MessageConfig
}

func main() {
	bot, err := tgbotapi.NewBotAPI("7212937598:AAFbYryUxv8iw0fWaOEgRd-te-WJpOIKaF8")
	if err != nil {
		log.Panic(err)
	}

	var users_data = make(map[int64]user_data)
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var tmp record
	records := new([]record)
	var tmp_time time.Time
	updates := bot.GetUpdatesChan(u)
	go remind_sender(records, bot)

	for update := range updates {
		if update.Message == nil { //пустые обновления игнорируем
			continue
		}
		if update.Message.IsCommand() { //это команда?
			switch update.Message.Command() {
			case "help":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "To create new timer, use /new_timer command"))

			case "new_timer": //создаём новый таймер, статус пользователя переводим на один- 0 я зарезервировал для пользователя который ещё ничего не делает
				tmp_loc, _ := time.LoadLocation("Local")
				users_data[update.Message.Chat.ID] = user_data{1, update.Message.Chat.ID, "", tmp_loc, 0}
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "First- what do i need to remind you about?"))

			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Fine, i'll tell you how to use this abomination - to create new timer, use /new_timer command. And next time you need help, use /help command, will you?"))
			}
		} else if update.Message.Text != "" { //если это текстовое сообщение
			if users_data[update.Message.Chat.ID].status == 1 { // если пользователь начал создание таймера, то его текстовое сообщение- напоминание
				users_data[update.Message.Chat.ID] = user_data{2, users_data[update.Message.Chat.ID].chat_id, update.Message.Text, users_data[update.Message.Chat.ID].loc, 0}
				//дальше, запрашиваем дату и время для напоминания
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Now, date of reminder in format dd:mm:yyyy hh:mm:ss. Or you can enter in incorrect format and break it, if you wish"))
			} else {
				if users_data[update.Message.Chat.ID].status == 2 { //если пользователь на этапе отправки даты
					tmp_time, _ = time.ParseInLocation("02:01:2006 15:04:05", strings.TrimSpace(update.Message.Text), users_data[update.Message.Chat.ID].loc)
					if tmp_time.Year() == 0001 { //проверяем, правильно ли введена дата, и дата ли это вообще
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Uh, you either wrote in incorrect format, or you didn't input date at all"))
					} else { //если всё верно, добавляем новый таймер в записи
						tmp.chat_id = tgbotapi.NewMessage(update.Message.Chat.ID, " ")
						tmp.loc = users_data[update.Message.Chat.ID].loc
						tmp.reminder = users_data[update.Message.Chat.ID].reminder
						tmp.timer = tmp_time.Unix()
						log.Println(tmp.timer)
						log.Println(time.Now().Unix())
						*records = append(*records, tmp)
						delete(users_data, update.Message.Chat.ID)
					}
				} else { //в любом другом случае, пишем пользователю напоминание о том как пользоваться этой штукой
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "To use this cursed abomination, use /new_timer command"))
				}
			}
		} else { //в любом другом случае, пишем пользователю напоминание о том как пользоваться этой штукой
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "To use this cursed abomination, use /new_timer command"))
		}
	}
}

func remind_sender(records *[]record, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		log.Println("ticker works")
		log.Println(len(*records))
		for i, rec := range *records {
			if rec.timer < time.Now().Unix() {
				rec.chat_id.Text = rec.reminder
				bot.Send(rec.chat_id)
				(*records)[i] = (*records)[len(*records)-1]
				(*records) = (*records)[:len(*records)-1]
			}
		}
	}
}

/*func new_timer(update tgbotapi.Update, msg tgbotapi.MessageConfig, u tgbotapi.UpdateConfig, updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) (record, bool) {
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
	bot.Send(msg)
	correct = false
	update = <-updates
	for !correct {
		if update.Message.IsCommand() || update.Message == nil {
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
}*/
