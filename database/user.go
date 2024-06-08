package database

import (
	"chain/structs"
	"chain/utils"
	"fmt"
)

func CreateUser(user structs.User) error {
	user.Password = utils.GeneratePassHash(user.Password)
	_, err := db.Exec("INSERT INTO user (`username`,`passhash`,`type`) VALUES (?, ?, ?)", user.Username, user.Password, user.Type)
	return err
}

func UpdateUser(user structs.User) error {
	fields := make(map[string]interface{})
	if user.Username != "" {
		fields["username"] = user.Username
	}
	if user.Password != "" {
		fields["passhash"] = utils.GeneratePassHash(user.Password)
	}
	if user.Type != 0 {
		fields["type"] = user.Type
	}
	if len(fields) == 0 {
		return nil
	}
	print("ffffffffffffff")
	fmt.Println(fields)
	query := "UPDATE user SET "
	args := []interface{}{}
	i := 0
	for k, v := range fields {
		if i > 0 {
			query += ", "
		}
		query += "`" + k + "` = ?"
		args = append(args, v)
		i++
	}
	query += " WHERE `uid` = ?"
	args = append(args, user.Uid)

	_, err := db.Exec(query, args...)
	return err
}

func DeleteUser(uid int) error {
	_, err := db.Exec("DELETE FROM user WHERE `uid`=?", uid)
	return err
}

func GetAllUsers(pagenum int, pagesize int) (int, []structs.User, error) {
	// 计算 OFFSET
	offset := (pagenum - 1) * pagesize
	// 执行带有 LIMIT 和 OFFSET 的 SQL 查询
	rows, err := db.Query("SELECT `uid`,`username`,`type` FROM user LIMIT ? OFFSET ?", pagesize, offset)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	var users []structs.User
	for rows.Next() {
		var user structs.User
		err := rows.Scan(&user.Uid, &user.Username, &user.Type)
		if err != nil {
			return 0, nil, err
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		users = make([]structs.User, 0)
	}

	err = rows.Err()
	if err != nil {
		return 0, nil, err
	}
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
	if err != nil {
		return 0, nil, err
	}
	return count, users, nil
}

func GetUserByUsername(username string) (structs.User, error) {
	var user structs.User
	err := db.QueryRow("SELECT uid, username, type, passhash FROM user WHERE `username`=?", username).Scan(&user.Uid, &user.Username, &user.Type, &user.Password)
	if err != nil {
		return structs.User{}, err
	}
	return user, nil
}

func GetUserByUid(uid int) (structs.User, error) {
	var user structs.User
	err := db.QueryRow("SELECT uid, username, type, passhash FROM user WHERE `uid`=?", uid).Scan(&user.Uid, &user.Username, &user.Type, &user.Password)
	if err != nil {
		return structs.User{}, err
	}
	return user, nil
}

func GetUserByPartialName(partialname string, pagenum int, pagesize int) ([]structs.User, error) {
	offset := (pagenum - 1) * pagesize
	rows, err := db.Query("SELECT DISTINCT uid, username FROM user WHERE username LIKE ? LIMIT ? OFFSET ?", "%"+partialname+"%", pagesize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []structs.User
	for rows.Next() {
		var user structs.User
		err := rows.Scan(&user.Uid, &user.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		users = make([]structs.User, 0)
	}
	return users, nil
}
