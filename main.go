package main

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name        string
	Codeforces  string
	Informatics string
	School      string
	About       string
}

var bot *tgbotapi.BotAPI
var err error
var DB *sql.DB

func isAdmin(id int) bool {
	ids := getAdmins()
	if err != nil {
		catchError(err)
		return false
	}
	for _, admin_id := range ids {
		if id == admin_id {
			return true
		}
	}
	return false
}

func addAdmin(id int) {
	query := "INSERT INTO admins (id) VALUES (?)"
	_, err := DB.Exec(query, id)
	if err != nil {
		catchError(err)
	}
}

func getAdmins() []int {
	rows, err := DB.Query("SELECT id FROM admins")
	if err != nil {
		catchError(err)
		return nil
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			catchError(err)
			return nil
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		catchError(err)
		return nil
	}
	return ids
}

func addGroupID(id int) {
	query := "INSERT INTO chats (id) VALUES (?)"
	_, err := DB.Exec(query, id)
	if err != nil {
		catchError(err)
	}
}

func addUser(id int) {
	query := "INSERT INTO users (id, status) VALUES (?, ?)"
	_, err := DB.Exec(query, id, 0)
	if err != nil {
		catchError(err)
	}
}

func deleteUser(id int) {
	query := "DELETE FROM users WHERE id=?"
	_, err := DB.Exec(query, id)
	if err != nil {
		catchError(err)
	}
}

func banUser(id int) {
	setStatus(id, -404)
}

func unbanUser(id int) {
	setStatus(id, -1)
}

func addMessage(id, user_id, message_id int) {
	query := "INSERT INTO messages (id, user_id, message_id) VALUES (?, ?, ?)"
	_, err := DB.Exec(query, id, user_id, message_id)
	if err != nil {
		catchError(err)
	}
}

func getMessage(id int) (int, int) {
	var user_id, message_id int
	err = DB.QueryRow("SELECT user_id, message_id FROM messages WHERE id = ?", id).Scan(&user_id, &message_id)
	if err != nil {
		return -2, -2
	}
	return user_id, message_id
}

func getStatus(id int) int {
	var status int
	err = DB.QueryRow("SELECT status FROM users WHERE id = ?", id).Scan(&status)
	if err != nil {
		return -2
	}
	return status
}

func setStatus(id int, status int) {
	query := "UPDATE users SET status = ? WHERE id = ?"
	_, err := DB.Exec(query, status, id)
	if err != nil {
		catchError(err)
	}
}

func setParameter(id int, par, val string) {
	query := "UPDATE users SET " + par + " = ? WHERE id = ?"
	_, err := DB.Exec(query, val, id)
	if err != nil {
		catchError(err)
	}
}

func getUsers() []int {
	rows, err := DB.Query("SELECT id FROM users")
	if err != nil {
		catchError(err)
		return nil
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			catchError(err)
			return nil
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		catchError(err)
		return nil
	}
	return ids
}

func getUsersParameter(parameter string, delimeter string) string {
	if delimeter == "\\n" {
		delimeter = "\n"
	}
	if delimeter == "\\t" {
		delimeter = "\t"
	}
	rows, err := DB.Query("SELECT " + parameter + " FROM users")
	if err != nil {
		catchError(err)
		return ""
	}
	defer rows.Close()
	var values []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			continue
		}
		values = append(values, val)
	}
	if err := rows.Err(); err != nil {
		catchError(err)
		return ""
	}
	res := ""
	for i, val := range values {
		res += val
		if i != len(values)-1 {
			res += delimeter
		}
	}
	return res
}

func getGroupIDs() []int {
	rows, err := DB.Query("SELECT id FROM chats")
	if err != nil {
		catchError(err)
		return nil
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			catchError(err)
			return nil
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		catchError(err)
		return nil
	}
	return ids
}

func addInNewGroup(update tgbotapi.Update) {
	flag := isAdmin(update.Message.From.ID)
	if flag {
		sendMessage(int(update.Message.Chat.ID), "Here we go again")
		sendMessage(int(update.Message.Chat.ID), `Команды, доступные только админам:
		/send_message [string message] - отправить сообщение всем пользователям
		/get_parameter [string parameter_name] [string delimeter] - получить список этого параметра всех пользователей с разделителем
		/add_admin [int id] - добавить админа
		/delete_user [int id] - удалить пользователя
		/add_question [string question] [string parameter] - добавить вопрос
		/ban_user [string question] [string parameter] - забанить пользователя (совсем)`)
		addGroupID(int(update.Message.Chat.ID))
	} else {
		detectYoungHacker(update)
		leaveChatConfig := tgbotapi.ChatConfig{
			ChatID: update.Message.Chat.ID,
		}
		_, err = bot.LeaveChat(leaveChatConfig)
		if err != nil {
			catchError(err)
		}
	}
}

func detectYoungHacker(update tgbotapi.Update) {
	sendMessage(update.Message.From.ID, "По ФЗ-35 гл. 23 за попытку доступа к приватным функциям вы приговариваетесь к бану.")
	ids := getGroupIDs()
	for _, id := range ids {
		sendMessage(id, "Пользователь с ID "+fmt.Sprint(update.Message.From.ID)+" и ником @"+fmt.Sprint(update.Message.From.UserName)+" попытался использовать функционал, на который у него нет прав:")
		if update.Message.NewChatMembers == nil && !update.Message.GroupChatCreated {
			sendForward(id, int(update.Message.Chat.ID), update.Message.MessageID)
		} else {
			sendMessage(id, "Добавление в чат.")
		}
	}
}

func sendMessage(id int, text string) int {
	msg, err := bot.Send(tgbotapi.NewMessage(int64(id), text))
	if err != nil {
		catchError(err)
	}
	return msg.MessageID
}

func sendForward(id1, id2, id3 int) int {
	msg, err := bot.Send(tgbotapi.NewForward(int64(id1), int64(id2), id3))
	if err != nil {
		catchError(err)
	}
	return msg.MessageID
}

func catchError(err error) {
	ids := getGroupIDs()
	pc, _, _, _ := runtime.Caller(1)
	callerName := runtime.FuncForPC(pc).Name()
	for _, id := range ids {
		sendMessage(id, "Произошла ошибка в функции '"+callerName+"':\n"+err.Error())
	}
}

func getCountQuestions() int {
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	if err != nil {
		catchError(err)
	}
	return count
}

func getQuestion(id int) string {
	var question string
	err = DB.QueryRow("SELECT question FROM questions WHERE id = ?", id).Scan(&question)
	if err != nil {
		catchError(err)
	}
	return question
}

func getParameter(id int) string {
	var parameter string
	err = DB.QueryRow("SELECT parameter FROM questions WHERE id = ?", id).Scan(&parameter)
	if err != nil {
		catchError(err)
	}
	return parameter
}

func addQuestion(question, parameter string) {
	query := "ALTER TABLE users ADD COLUMN " + parameter
	_, err = DB.Exec(query)
	if err != nil {
		catchError(err)
		return
	}

	question_id := getCountQuestions()
	query = "INSERT INTO questions (id, question, parameter) VALUES (?, ?, ?)"
	_, err = DB.Exec(query, question_id, question, parameter)
	if err != nil {
		catchError(err)
	}

	ids := getUsers()
	for _, id := range ids {
		status := getStatus(id)
		if status != -1 {
			continue
		}
		setStatus(id, question_id)
		sendMessage(id, question)
	}
}

func main() {
	DB, err = sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Panic(err)
		return
	}
	defer DB.Close()

	bot, err = tgbotapi.NewBotAPI("7031978149:AAEgExfnWUmJDJyydevedkA3Vr_-THRQ-P8")
	if err != nil {
		log.Panic(err)
		return
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
		return
	}

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.From.ID == bot.Self.ID {
				continue
			}
			if update.Message.IsCommand() {
				command := update.Message.Command()
				if command == "start" {
					status := getStatus(update.Message.From.ID)
					if status != -2 {
						sendMessage(update.Message.From.ID, "Уже зарегистрирован")
					} else {
						addUser(update.Message.From.ID)
						sendMessage(update.Message.From.ID, getQuestion(0))
					}
					continue
				}
				flag := isAdmin(update.Message.From.ID)
				switch command {
				case "send_message":
					if flag {
						text := update.Message.CommandArguments()
						ids := getUsers()
						for _, id := range ids {
							sendMessage(id, text)
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
					} else {
						detectYoungHacker(update)
					}
				case "get_parameter":
					if flag {
						arguments := strings.Fields(update.Message.CommandArguments())
						text := ""
						if len(arguments) == 2 {
							text = getUsersParameter(arguments[0], arguments[1])
						} else {
							sendMessage(int(update.Message.Chat.ID), "Ожидается два аргумента")
						}
						sendMessage(int(update.Message.Chat.ID), text)
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
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
							addAdmin(id)
							sendMessage(int(update.Message.Chat.ID), "Админ с id "+argument+" успешно добавлен")
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
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
							deleteUser(id)
							sendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно удалён")
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
					} else {
						detectYoungHacker(update)
					}
				case "add_question":
					if flag {
						arguments := strings.Fields(update.Message.CommandArguments())
						if len(arguments) < 2 {
							sendMessage(int(update.Message.Chat.ID), "Ожидается хотя бы два аргумента")
						} else {
							question := ""
							for i := 0; i+1 < len(arguments); i++ {
								question += arguments[i]
								if i+1 != len(arguments) {
									question += " "
								}
							}
							addQuestion(question, arguments[len(arguments)-1])
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
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
							banUser(id)
							sendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно забанен")
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
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
							unbanUser(id)
							sendMessage(int(update.Message.Chat.ID), "Пользователь с ID "+argument+" успешно восстановлен в правах")
						}
						sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
					} else {
						detectYoungHacker(update)
					}
				}
			} else if update.Message.Chat.Type == "private" {
				status := getStatus(update.Message.From.ID)
				if status == -404 {
					sendMessage(update.Message.From.ID, "Вы были заблокированы, обратитесь к админу для разблокировки")
				} else if status == -2 {
					sendMessage(update.Message.From.ID, "Зарегистрируйтесь с помощью /start перед использованием")
				} else if status == -1 {
					ids := getGroupIDs()
					for _, id := range ids {
						msg_id := sendForward(id, int(update.Message.Chat.ID), update.Message.MessageID)
						addMessage(msg_id, update.Message.From.ID, update.Message.MessageID)
					}
					if len(ids) > 0 {
						sendMessage(update.Message.From.ID, "Вопрос отправлен")
					} else {
						sendMessage(update.Message.From.ID, "Нет доступных чатов, обратитесь к администратору")
					}
				} else if status == getCountQuestions()-1 {
					setParameter(update.Message.From.ID, getParameter(status), update.Message.Text)
					setStatus(update.Message.From.ID, -1)
					sendMessage(update.Message.From.ID, "Регистрация пройдена успешно")
					ids := getGroupIDs()
					for _, id := range ids {
						sendMessage(id, "Пользователь с ником @"+update.Message.From.UserName+" и ID "+fmt.Sprint(update.Message.From.ID)+" успешно зарегистрирован")
					}
				} else {
					setParameter(update.Message.From.ID, getParameter(status), update.Message.Text)
					setStatus(update.Message.From.ID, status+1)
					sendMessage(update.Message.From.ID, getQuestion(status+1))
				}
			} else if update.Message.NewChatMembers != nil {
				for _, member := range *update.Message.NewChatMembers {
					if member.ID == bot.Self.ID {
						addInNewGroup(update)
					}
				}
			} else if update.Message.GroupChatCreated {
				addInNewGroup(update)
			} else if update.Message.Chat.Type == "group" {
				if update.Message.ReplyToMessage != nil {
					user_id, message_id := getMessage(update.Message.ReplyToMessage.MessageID)
					if user_id != -2 {
						if update.Message.Text != "" {
							msg := tgbotapi.NewMessage(int64(user_id), update.Message.Text)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Photo != nil {
							msg := tgbotapi.NewPhotoUpload(int64(user_id), update.Message.Photo)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Document != nil {
							msg := tgbotapi.NewDocumentUpload(int64(user_id), update.Message.Document)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Sticker != nil {
							msg := tgbotapi.NewStickerUpload(int64(user_id), update.Message.Sticker)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Audio != nil {
							msg := tgbotapi.NewAudioUpload(int64(user_id), update.Message.Audio)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Voice != nil {
							msg := tgbotapi.NewVoiceUpload(int64(user_id), update.Message.Voice)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Video != nil {
							msg := tgbotapi.NewVideoUpload(int64(user_id), update.Message.Video)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						} else if update.Message.Animation != nil {
							msg := tgbotapi.NewAnimationUpload(int64(user_id), update.Message.Animation)
							msg.ReplyToMessageID = message_id
							_, err = bot.Send(msg)
							if err != nil {
								catchError(err)
							}
						}
					}
				}
				sendMessage(int(update.Message.Chat.ID), "Запрос обработан")
			}
		}
	}
}
