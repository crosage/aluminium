package database

import (
	"chain/structs"
	"chain/utils"
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
		file.ShareCode = utils.Decrypt(file.ShareCode)
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

func SaveFile(file structs.File) error {
	encryptedShareCode := utils.Encrypt(file.ShareCode)
	_, err := db.Exec(`
		INSERT INTO 
		file(hash,path,uid,share_code)
		VALUES (?,?,?,?)
	`, file.Hash, file.Path, file.Uid, encryptedShareCode)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save file path")
		return err
	}
	return nil
}
