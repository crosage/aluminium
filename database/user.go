package database

import (
	"chain/structs"
	"chain/utils"
)

func CreateUser(user structs.User) error {
	user.Password = utils.GeneratePassHash(user.Password)
	_, err := db.Exec("INSERT INTO user (`username`,`passhash`,`type`) VALUES (?, ?, ?)", user.Username, user.Password, user.Type)
	return err
}

func UpdateUser(user structs.User) error {
	user.Password = utils.GeneratePassHash(user.Password)
	_, err := db.Exec("UPDATE user SET `username` = ?,`passhash` = ?,`type` = ? WHERE `uid` = ?", user.Username, user.Password, user.Type, user.Uid)
	return err
}

func DeleteUser(uid int) error {
	_, err := db.Exec("DELETE FROM user WHERE `uid`=?", uid)
	return err
}

func GetAllUsers() (int, []structs.User, error) {
	rows, err := db.Query("SELECT `uid`,`username`,`type` FROM user")
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
