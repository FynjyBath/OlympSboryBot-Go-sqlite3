package main

import (
	"config"
	"fmt"
	"log"
	"my_database"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"tgbot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

var bot tgbot.TGBot
var DB my_database.DataBaseSites
var MU sync.Mutex
var cfg config.Config

func SendMessage(id int, text string) int {
	id, err := bot.SendMessage(id, text)
	if err != nil {
		catchError(err)
	}
	return id
}

func SendForward(id1, id2, id3 int) int {
	id, err := bot.SendForward(id1, id2, id3)
	if err != nil {
		catchError(err)
	}
	return id
}

func catchError(err error) {
	ids, _ := DB.GetGroupIDs()
	pc, _, _, _ := runtime.Caller(1)
	callerName := runtime.FuncForPC(pc).Name()
	for _, id := range ids {
		SendMessage(id, "Произошла ошибка в функции '"+callerName+"':\n"+err.Error())
	}
}

func addInNewGroup(update tgbotapi.Update) {
	flag, err := DB.IsAdmin(update.Message.From.ID)
	if err != nil {
		catchError(err)
	}
	if flag {
		SendMessage(int(update.Message.Chat.ID), "Here we go again")
		SendMessage(int(update.Message.Chat.ID), `Команды, доступные только админам:
		/send_message [string message] - отправить сообщение всем пользователям
		/get_parameter [string parameter_name] [string delimeter] - получить список этого параметра всех пользователей с разделителем
		/add_admin [int id] - добавить админа
		/delete_user [int id] - удалить пользователя
		/add_question [string question] [string parameter] - добавить вопрос
		/ban_user [string question] [string parameter] - забанить пользователя (совсем)`)
		err = DB.AddGroupID(int(update.Message.Chat.ID))
		if err != nil {
			catchError(err)
		}
	} else {
		detectYoungHacker(update)
		leaveChatConfig := tgbotapi.ChatConfig{
			ChatID: update.Message.Chat.ID,
		}
		_, err = bot.Bot.LeaveChat(leaveChatConfig)
		if err != nil {
			catchError(err)
		}
	}
}

func detectYoungHacker(update tgbotapi.Update) {
	SendMessage(update.Message.From.ID, "По ФЗ-35 гл. 23 за попытку доступа к приватным функциям вы приговариваетесь к бану.")
	ids, err := DB.GetGroupIDs()
	if err != nil {
		catchError(err)
	}
	for _, id := range ids {
		SendMessage(id, "Пользователь с ID "+fmt.Sprint(update.Message.From.ID)+" и ником @"+fmt.Sprint(update.Message.From.UserName)+" попытался использовать функционал, на который у него нет прав:")
		if update.Message.NewChatMembers == nil && !update.Message.GroupChatCreated {
			SendForward(id, int(update.Message.Chat.ID), update.Message.MessageID)
		} else {
			SendMessage(id, "Добавление в чат.")
		}
	}
}

func handleCommand(update tgbotapi.Update) {
	command := update.Message.Command()
	if command == "start" {
		status := DB.GetStatus(update.Message.From.ID)
		if status != -2 {
			SendMessage(update.Message.From.ID, "Уже зарегистрирован")
		} else {
			err := DB.AddUser(update.Message.From.ID)
			if err != nil {
				catchError(err)
			}
			question, err := DB.GetQuestion(0)
			if err != nil {
				catchError(err)
			}
			SendMessage(update.Message.From.ID, question)
		}
		return
	}
	flag, err := DB.IsAdmin(update.Message.From.ID)
	if err != nil {
		catchError(err)
	}
	switch command {
	case "send_message":
		if flag {
			text := update.Message.CommandArguments()
			ids, err := DB.GetUsers()
			if err != nil {
				catchError(err)
			}
			for _, id := range ids {
				SendMessage(id, text)
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "get_parameter":
		if flag {
			arguments := strings.Fields(update.Message.CommandArguments())
			text := ""
			if len(arguments) == 2 {
				text, err = DB.GetUsersParameter(arguments[0], arguments[1])
				if err != nil {
					catchError(err)
				}
			} else {
				SendMessage(int(update.Message.Chat.ID), "Ожидается два аргумента")
			}
			SendMessage(int(update.Message.Chat.ID), text)
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "add_admin":
		if flag {
			argument := update.Message.CommandArguments()
			id, err := strconv.Atoi(argument)
			if err != nil {
				catchError(err)
			} else {
				err := DB.AddAdmin(id)
				if err != nil {
					catchError(err)
				}
				SendMessage(int(update.Message.Chat.ID), "Админ с id "+argument+" успешно добавлен")
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "delete_user":
		if flag {
			argument := update.Message.CommandArguments()
			id, err := strconv.Atoi(argument)
			if err != nil {
				catchError(err)
			} else {
				err := DB.DeleteUser(id)
				if err != nil {
					catchError(err)
				}
				SendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно удалён")
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "add_question":
		if flag {
			arguments := strings.Fields(update.Message.CommandArguments())
			if len(arguments) < 2 {
				SendMessage(int(update.Message.Chat.ID), "Ожидается хотя бы два аргумента")
			} else {
				question := ""
				for i := 0; i+1 < len(arguments); i++ {
					question += arguments[i]
					if i+1 != len(arguments) {
						question += " "
					}
				}
				question_id, err := DB.AddQuestion(question, arguments[len(arguments)-1])
				if err != nil {
					catchError(err)
				}
				ids, err := DB.GetUsers()
				if err != nil {
					catchError(err)
				}
				for _, id := range ids {
					status := DB.GetStatus(id)
					if status != -1 {
						continue
					}
					err = DB.SetStatus(id, question_id)
					if err != nil {
						catchError(err)
					}
					SendMessage(id, question)
				}
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "ban_user":
		if flag {
			argument := update.Message.CommandArguments()
			id, err := strconv.Atoi(argument)
			if err != nil {
				catchError(err)
			} else {
				err := DB.BanUser(id)
				if err != nil {
					catchError(err)
				}
				SendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно забанен")
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	case "unban_user":
		if flag {
			argument := update.Message.CommandArguments()
			id, err := strconv.Atoi(argument)
			if err != nil {
				catchError(err)
			} else {
				err := DB.UnbanUser(id)
				if err != nil {
					catchError(err)
				}
				SendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно восстановлен в правах")
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		} else {
			detectYoungHacker(update)
		}
	}
}

func handlePrivateMessage(update tgbotapi.Update) {
	countQuestions, err := DB.GetCountQuestions()
	if err != nil {
		catchError(err)
	}
	status := DB.GetStatus(update.Message.From.ID)
	if status == -404 {
		SendMessage(update.Message.From.ID, "Вы были заблокированы, обратитесь к админу для разблокировки")
	} else if status == -2 {
		SendMessage(update.Message.From.ID, "Зарегистрируйтесь с помощью /start перед использованием")
	} else if status == -1 {
		ids, err := DB.GetGroupIDs()
		if err != nil {
			catchError(err)
		}
		for _, id := range ids {
			msg_id := SendForward(id, int(update.Message.Chat.ID), update.Message.MessageID)
			DB.AddMessage(msg_id, update.Message.From.ID, update.Message.MessageID)
			if err != nil {
				catchError(err)
			}
		}
		if len(ids) > 0 {
			SendMessage(update.Message.From.ID, "Вопрос отправлен")
		} else {
			SendMessage(update.Message.From.ID, "Нет доступных чатов, обратитесь к администратору")
		}
	} else if status == countQuestions-1 {
		parameter, err := DB.GetParameter(status)
		if err != nil {
			catchError(err)
		}
		err = DB.SetParameter(update.Message.From.ID, parameter, update.Message.Text)
		if err != nil {
			catchError(err)
		}
		err = DB.SetStatus(update.Message.From.ID, -1)
		if err != nil {
			catchError(err)
		}
		SendMessage(update.Message.From.ID, "Регистрация пройдена успешно")
		ids, err := DB.GetGroupIDs()
		if err != nil {
			catchError(err)
		}
		for _, id := range ids {
			SendMessage(id, "Пользователь с ником @"+update.Message.From.UserName+" и ID "+fmt.Sprint(update.Message.From.ID)+" успешно зарегистрирован")
		}
	} else {
		parameter, err := DB.GetParameter(status)
		if err != nil {
			catchError(err)
		}
		err = DB.SetParameter(update.Message.From.ID, parameter, update.Message.Text)
		if err != nil {
			catchError(err)
		}
		err = DB.SetStatus(update.Message.From.ID, status+1)
		if err != nil {
			catchError(err)
		}
		question, err := DB.GetQuestion(status + 1)
		if err != nil {
			catchError(err)
		}
		SendMessage(update.Message.From.ID, question)
	}
}

func handleGroupAnswer(update tgbotapi.Update) {
	user_id, message_id := DB.GetMessage(update.Message.ReplyToMessage.MessageID)
	if user_id != -2 {
		if update.Message.Text != "" {
			msg := tgbotapi.NewMessage(int64(user_id), update.Message.Text)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Photo != nil {
			msg := tgbotapi.NewPhotoUpload(int64(user_id), update.Message.Photo)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Document != nil {
			msg := tgbotapi.NewDocumentUpload(int64(user_id), update.Message.Document)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Sticker != nil {
			msg := tgbotapi.NewStickerUpload(int64(user_id), update.Message.Sticker)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Audio != nil {
			msg := tgbotapi.NewAudioUpload(int64(user_id), update.Message.Audio)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Voice != nil {
			msg := tgbotapi.NewVoiceUpload(int64(user_id), update.Message.Voice)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Video != nil {
			msg := tgbotapi.NewVideoUpload(int64(user_id), update.Message.Video)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		} else if update.Message.Animation != nil {
			msg := tgbotapi.NewAnimationUpload(int64(user_id), update.Message.Animation)
			msg.ReplyToMessageID = message_id
			_, err := bot.Bot.Send(msg)
			if err != nil {
				catchError(err)
			}
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if update.Message.From.ID == bot.Bot.Self.ID {
			return
		}
		if update.Message.IsCommand() {
			handleCommand(update)
		} else if update.Message.Chat.Type == "private" {
			handlePrivateMessage(update)
		} else if update.Message.NewChatMembers != nil {
			for _, member := range *update.Message.NewChatMembers {
				if member.ID == bot.Bot.Self.ID {
					addInNewGroup(update)
				}
			}
		} else if update.Message.GroupChatCreated {
			addInNewGroup(update)
		} else if update.Message.Chat.Type == "group" {
			if update.Message.ReplyToMessage != nil {
				handleGroupAnswer(update)
			}
			SendMessage(int(update.Message.Chat.ID), "Запрос обработан")
		}
	}
}

func main() {
	cfg = config.LoadConfig("config.json")

	DB.Init()
	defer DB.DB.Close()
	log.Println("Connected to the database")

	bot.Init(cfg.TGBotKey)
	log.Printf("Authorized on account %s", bot.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
		return
	}

	for update := range updates {
		handleUpdate(update)
	}
}
