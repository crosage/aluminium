package database

import (
	"chain/structs"
	"github.com/rs/zerolog/log"
)

func GetFilesByUid(uid int) ([]structs.File, error) {
	rows, err := db.Query("SELECT hash, path, share_code FROM file WHERE uid = ?", uid)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query for getting user files")
		return nil, err
	}
	defer rows.Close()

	var files []structs.File
	for rows.Next() {
		var file structs.File
		err := rows.Scan(&file.Hash, &file.Path, &file.ShareCode)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan file row")
			return nil, err
		}
		file.Uid = uid
		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		log.Error().Err(err).Msg("Row iteration error")
		return nil, err
	}

	if len(files) == 0 {
		files = make([]structs.File, 0)
	}

	return files, nil
}

func SaveFile(file structs.File) (int, error) {
	result, err := db.Exec(`
		INSERT INTO 
		file(hash,path,uid,share_code)
		VALUES (?,?,?,?)
	`, file.Hash, file.Path, file.Uid, file.ShareCode)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save file path")
		return 0, err
	}

	fid, err := result.LastInsertId()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get last insert ID")
		return 0, err
	}

	return int(fid), nil
}

func GrantFileAccess(uid, fid int) error {
	_, err := db.Exec("INSERT INTO user_access (user_id, file_id) VALUES (?, ?)", uid, fid)
	if err != nil {
		return err
	}
	return nil
}

func GrantFileAccessIfValidShareCode(uid int, shareCode string) error {
	var fid int
	err := db.QueryRow("SELECT id FROM file WHERE share_code = ?", shareCode).Scan(&fid)
	if err != nil {
		return err
	}

	// 更新用户文件访问权限
	err = GrantFileAccess(uid, fid)
	if err != nil {
		return err
	}

	return nil
}
