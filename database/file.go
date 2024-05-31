package database

import (
	"chain/structs"
	"github.com/rs/zerolog/log"
)

func GetFilesCreatedByUid(uid int) ([]structs.File, error) {
	rows, err := db.Query("SELECT hash, path, share_code,name FROM file WHERE uid = ?", uid)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query for getting user files")
		return nil, err
	}
	defer rows.Close()

	var files []structs.File
	for rows.Next() {
		var file structs.File
		err := rows.Scan(&file.Hash, &file.Path, &file.ShareCode, &file.Name)
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
		file(hash,path,uid,share_code,name)
		VALUES (?,?,?,?,?)
	`, file.Hash, file.Path, file.Uid, file.ShareCode, file.Name)
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
	_, err := db.Exec("INSERT OR REPLACE INTO user_access (user_id, file_id) VALUES (?, ?)", uid, fid)
	if err != nil {
		return err
	}
	return nil
}

func GetFileAccess(uid, fid int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_access WHERE user_id = ? AND file_id = ?)", uid, fid).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GrantFileAccessIfValidShareCode(uid int, shareCode string) error {
	var fid int
	err := db.QueryRow("SELECT fid FROM file WHERE share_code = ?", shareCode).Scan(&fid)
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

func GetFilesAvailableByUid(uid int) ([]structs.File, error) {
	var files []structs.File

	rows, err := db.Query(`
		SELECT 
		f.hash,f.path,f.name,f.fid,f.uid
		FROM file f 
		INNER JOIN user_access ua
		ON f.fid=ua.file_id WHERE ua.user_id=?
	`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file structs.File
		if err := rows.Scan(&file.Hash, &file.Path, &file.Name, &file.Fid, &file.Fid); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func GetFileByFid(fid int) (structs.File, error) {
	var file structs.File
	err := db.QueryRow(`
		SELECT 
		f.hash,f.path,f.name,f.fid,f.uid
		FROM file f 
		INNER JOIN user_access ua
		ON f.fid=ua.file_id WHERE f.fid=?
	`, fid).Scan(&file.Hash, &file.Path, &file.Name, &file.Fid, &file.Uid)
	if err != nil {
		return file, err
	}
	return file, nil
}
