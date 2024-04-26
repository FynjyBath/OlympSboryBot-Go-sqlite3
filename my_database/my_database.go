package my_database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DataBaseSites struct {
	DB *sql.DB
}

func (dbs *DataBaseSites) Init() {
	DB, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	dbs.DB = DB
}

func (dbs *DataBaseSites) IsAdmin(id int) (bool, error) {
	ids, err := dbs.GetAdmins()
	if err != nil {
		return false, err
	}
	for _, admin_id := range ids {
		if id == admin_id {
			return true, nil
		}
	}
	return false, nil
}

func (dbs *DataBaseSites) AddAdmin(id int) error {
	query := "INSERT INTO admins (id) VALUES (?)"
	_, err := dbs.DB.Exec(query, id)
	return err
}

func (dbs *DataBaseSites) GetGroupIDs() ([]int, error) {
	rows, err := dbs.DB.Query("SELECT id FROM chats")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (dbs *DataBaseSites) GetCountQuestions() (int, error) {
	var count int
	err := dbs.DB.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (dbs *DataBaseSites) GetQuestion(id int) (string, error) {
	var question string
	err := dbs.DB.QueryRow("SELECT question FROM questions WHERE id = ?", id).Scan(&question)
	if err != nil {
		return "", err
	}
	return question, nil
}

func (dbs *DataBaseSites) GetParameter(id int) (string, error) {
	var parameter string
	err := dbs.DB.QueryRow("SELECT parameter FROM questions WHERE id = ?", id).Scan(&parameter)
	if err != nil {
		return "", err
	}
	return parameter, nil
}

func (dbs *DataBaseSites) GetAdmins() ([]int, error) {
	rows, err := dbs.DB.Query("SELECT id FROM admins")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (dbs *DataBaseSites) AddGroupID(id int) error {
	query := "INSERT INTO chats (id) VALUES (?)"
	_, err := dbs.DB.Exec(query, id)
	return err
}

func (dbs *DataBaseSites) AddUser(id int) error {
	query := "INSERT INTO users (id, status) VALUES (?, ?)"
	_, err := dbs.DB.Exec(query, id, 0)
	return err
}

func (dbs *DataBaseSites) DeleteUser(id int) error {
	query := "DELETE FROM users WHERE id=?"
	_, err := dbs.DB.Exec(query, id)
	return err
}

func (dbs *DataBaseSites) BanUser(id int) error {
	return dbs.SetStatus(id, -404)
}

func (dbs *DataBaseSites) UnbanUser(id int) error {
	return dbs.SetStatus(id, -1)
}

func (dbs *DataBaseSites) AddMessage(id, user_id, message_id int) error {
	query := "INSERT INTO messages (id, user_id, message_id) VALUES (?, ?, ?)"
	_, err := dbs.DB.Exec(query, id, user_id, message_id)
	return err
}

func (dbs *DataBaseSites) GetMessage(id int) (int, int) {
	var user_id, message_id int
	err := dbs.DB.QueryRow("SELECT user_id, message_id FROM messages WHERE id = ?", id).Scan(&user_id, &message_id)
	if err != nil {
		return -2, -2
	}
	return user_id, message_id
}

func (dbs *DataBaseSites) GetStatus(id int) int {
	var status int
	err := dbs.DB.QueryRow("SELECT status FROM users WHERE id = ?", id).Scan(&status)
	if err != nil {
		return -2
	}
	return status
}

func (dbs *DataBaseSites) SetStatus(id int, status int) error {
	query := "UPDATE users SET status = ? WHERE id = ?"
	_, err := dbs.DB.Exec(query, status, id)
	return err
}

func (dbs *DataBaseSites) SetParameter(id int, par, val string) error {
	query := "UPDATE users SET " + par + " = ? WHERE id = ?"
	_, err := dbs.DB.Exec(query, val, id)
	return err
}

func (dbs *DataBaseSites) GetUsers() ([]int, error) {
	rows, err := dbs.DB.Query("SELECT id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (dbs *DataBaseSites) GetUsersParameter(parameter string, delimeter string) (string, error) {
	if delimeter == "\\n" {
		delimeter = "\n"
	}
	if delimeter == "\\t" {
		delimeter = "\t"
	}
	rows, err := dbs.DB.Query("SELECT " + parameter + " FROM users")
	if err != nil {
		return "", err
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
		return "", err
	}
	res := ""
	for i, val := range values {
		res += val
		if i != len(values)-1 {
			res += delimeter
		}
	}
	return res, nil
}

func (dbs *DataBaseSites) AddQuestion(question, parameter string) (int, error) {
	query := "ALTER TABLE users ADD COLUMN " + parameter
	_, err := dbs.DB.Exec(query)
	if err != nil {
		return 0, err
	}

	question_id, err := dbs.GetCountQuestions()
	if err != nil {
		return 0, err
	}
	query = "INSERT INTO questions (id, question, parameter) VALUES (?, ?, ?)"
	_, err = dbs.DB.Exec(query, question_id, question, parameter)
	if err != nil {
		return 0, err
	}

	return question_id, nil
}