package database

import (
	"chain/structs"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

func GetFilesCreatedByUid(uid int, pagenum int, pagesize int) ([]structs.File, error) {
	offset := (pagenum - 1) * pagesize
	rows, err := db.Query(`
	SELECT 
		f.hash, f.path, f.share_code, f.name, u.username ,f.fid
	FROM 
		file f 
	INNER JOIN 
		user u ON f.uid = u.uid 
	WHERE 
		f.uid = ? 
	LIMIT ? OFFSET ?
	`, uid, pagesize, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query for getting user files")
		return nil, err
	}
	defer rows.Close()
	var files []structs.File
	for rows.Next() {
		var file structs.File
		err := rows.Scan(&file.Hash, &file.Path, &file.ShareCode, &file.Name, &file.Username, &file.Fid)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan file row")
			return nil, err
		}
		file.Uid = uid
		err = AddFileChangeLog(uid, file.Fid, file.Name, time.Now().Format("2006-01-02 15:04:05"), "Check")
		if err != nil {
			fmt.Println("here database")
			return nil, err
		}
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

func GetFilesAvailableByUid(uid int, pagenum int, pagesize int) ([]structs.File, error) {
	var files []structs.File
	offset := (pagenum - 1) * pagesize
	rows, err := db.Query(`
	SELECT 
		f.hash, f.path, f.name, f.fid, f.uid, u.username,f.share_code
	FROM 
		file f 
	INNER JOIN 
		user_access ua ON f.fid = ua.file_id 
	INNER JOIN 
		user u ON ua.user_id = u.uid 
	WHERE 
		ua.user_id = ? 
	LIMIT ? OFFSET ?
	`, uid, pagesize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file structs.File
		if err := rows.Scan(&file.Hash, &file.Path, &file.Name, &file.Fid, &file.Uid, &file.Username, &file.ShareCode); err != nil {
			return nil, err
		}
		err = AddFileChangeLog(uid, file.Fid, file.Name, time.Now().Format("2006-01-02 15:04:05"), "Check")
		if err != nil {
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

func DeleteFileByFid(fid int) error {
	_, err := db.Exec(`DELETE FROM file WHERE fid = ?`, fid)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM user_access where user_access.file_id=?`, fid)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFileByFid(f structs.File) error {
	_, err := db.Exec(`
	UPDATE file 
	SET hash = ?,path=?,name=?,share_code=?
	WHERE fid = ?
	`, f.Hash, f.Path, f.Name, f.ShareCode, f.Fid)
	if err != nil {
		return err
	}
	return nil
}

func GetFileByPartialName(partialname string, pagenum int, pagesize int, uid int) ([]structs.File, error) {
	offset := (pagenum - 1) * pagesize
	rows, err := db.Query("SELECT hash, path, share_code, name FROM file WHERE uid=? AND name LIKE ? LIMIT ? OFFSET ?", uid, "%"+partialname+"%", pagesize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 处理查询结果
	var files []structs.File
	for rows.Next() {
		var file structs.File
		err := rows.Scan(&file.Hash, &file.Path, &file.ShareCode, &file.Name)
		if err != nil {
			return nil, err
		}
		err = AddFileChangeLog(uid, file.Fid, file.Name, time.Now().Format("2006-01-02 15:04:05"), "Check")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		files = make([]structs.File, 0)
	}
	return files, nil
}

func GetFileChangeLogByUid(uid int, pagenum int, pagesize int) ([]structs.FileLog, error) {
	var filelog []structs.FileLog
	offset := (pagenum - 1) * pagesize
	rows, err := db.Query(`
	SELECT 
		fc.user_id, fc.change_time,fc.operation,u.username,fc.file_name
	FROM 
		file_changelog fc
	INNER JOIN 
		user u ON fc.user_id = u.uid 
	WHERE 
		fc.user_id = ? 
	LIMIT ? OFFSET ?
	`, uid, pagesize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var log structs.FileLog
		if err := rows.Scan(&log.Uid, &log.Times, &log.Operation, &log.Username, &log.Name); err != nil {
			return nil, err
		}
		filelog = append(filelog, log)
	}
	return filelog, nil
}

func AddFileChangeLog(uid int, fid int, name string, times string, operation string) error {
	_, err := db.Exec("INSERT INTO file_changelog (user_id, file_id,file_name,change_time,operation) VALUES (?, ?,?,?,?)", uid, fid, name, times, operation)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
