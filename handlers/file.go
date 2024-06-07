package handlers

import (
	"bufio"
	"bytes"
	"chain/database"
	"chain/structs"
	"chain/utils"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
	"io"
	"os"
	"strconv"
	"time"
)

func pkcs5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func encryptAndSaveFile(src io.Reader, destPath string, ctx *fiber.Ctx) error {
	// 创建目标文件
	dest, err := os.Create(destPath)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "无法保存文件", nil)
	}
	defer dest.Close()
	// 使用 SM4 分组对称加密算法进行加密
	block, err := sm4.NewCipher([]byte(utils.CipherKey))
	if err != nil {
		return err
	}
	// 创建加密流
	blockSize := block.BlockSize()
	plaintext, err := io.ReadAll(src)
	origData := pkcs5Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, []byte(utils.IV))
	cryted := make([]byte, len(origData))
	blockMode.CryptBlocks(cryted, origData)
	if _, err := dest.Write(cryted); err != nil {
		return sendCommonResponse(ctx, 500, "加密失败", nil)
	}
	return nil
}

func handleFileUpload(ctx *fiber.Ctx) error {
	start := time.Now()
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	file, err := ctx.FormFile("file")
	if err != nil {
		return sendCommonResponse(ctx, 400, "", nil)
	}
	src, err := file.Open()
	if err != nil {
		return sendCommonResponse(ctx, 500, "无法打开文件", nil)
	}
	defer src.Close()
	hash := sm3.New()
	if _, err := io.Copy(hash, src); err != nil {
		return sendCommonResponse(ctx, 500, "无法生成文件哈希", nil)
	}
	hashStr := hex.EncodeToString(hash.Sum(nil))
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return sendCommonResponse(ctx, 500, "无法重置文件指针", nil)
	}
	fileuser, err := database.GetUserByUid(uid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	fileusername := fileuser.Username
	destPath := "./uploads/" + fileusername
	err = os.MkdirAll(destPath, 0755)
	if err != nil {
		return sendCommonResponse(ctx, 500, "创建用户目录失败", nil)
	}
	destPath = destPath + "/" + file.Filename
	if err := encryptAndSaveFile(src, destPath, ctx); err != nil {
		return sendCommonResponse(ctx, 500, "无法保存文件", nil)
	}
	str, _ := generateRandomString(16)
	fileRecord := structs.File{
		Hash:      hashStr,
		Path:      destPath,
		Uid:       uid,
		ShareCode: str,
		Name:      file.Filename,
	}
	fid, err := database.SaveFile(fileRecord)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	err = database.GrantFileAccess(uid, fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	defer func() {
		cost := time.Since(start)
		fmt.Println("cost=", cost)
	}()
	err = database.AddFileChangeLog(uid, fid, file.Filename, time.Now().Format("2006-01-02 15:04:05"), "Upload")
	if err != nil {
		return sendCommonResponse(ctx, 500, "日志插入失败", nil)
	}
	return sendCommonResponse(ctx, 200, "成功上传文件", nil)
}

func checkShareCode(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	type ShareCodeRequest struct {
		ShareCode string `json:"share_code"`
	}
	var share_code ShareCodeRequest
	err := jsoniter.Unmarshal(ctx.Body(), &share_code)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	err = database.GrantFileAccessIfValidShareCode(user.Uid, share_code.ShareCode)
	if err == sql.ErrNoRows {
		return sendCommonResponse(ctx, 403, "不存在对应文件", nil)
	} else if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "服务器内部错误", nil)
	} else {
		return sendCommonResponse(ctx, 200, "文件分享成功", nil)
	}
}

func getUserCreatedFiles(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页数错误", nil)
	}
	// 获取前端传来的 pageSize 参数，默认为 10
	pageSize, err := strconv.Atoi(ctx.Query("pagesize", "10"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页大小错误", nil)
	}
	files, err := database.GetFilesCreatedByUid(uid, page, pageSize)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
		"total": len(files),
		"files": files,
	})
}

func getUserAvailableFiles(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页数错误", nil)
	}
	// 获取前端传来的 pageSize 参数，默认为 10
	pageSize, err := strconv.Atoi(ctx.Query("pagesize", "10"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页大小错误", nil)
	}
	files, err := database.GetFilesAvailableByUid(uid, page, pageSize)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
		"total": len(files),
		"files": files,
	})
}

func pkcs5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func sm4Decrypt(key, iv, cipherText []byte) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cipherText))
	blockMode.CryptBlocks(origData, cipherText)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func getFileByFid(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	fidStr := ctx.Params("fid")
	fmt.Println("get" + fidStr)
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		return sendCommonResponse(ctx, 500, "服务器内部错误", nil)
	}
	exist, err := database.GetFileAccess(uid, fid)
	if exist == false {
		return sendCommonResponse(ctx, 403, "该用户没有该文件权限", nil)
	}
	file, err := database.GetFileByFid(fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "获取文件错误", nil)
	}
	filePath := file.Path
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return sendCommonResponse(ctx, 404, "文件不存在", nil)
	}
	filetext, err := os.Open(filePath)
	if err != nil {
		return sendCommonResponse(ctx, 500, "读取文件错误", nil)
	}
	defer filetext.Close()
	reader := bufio.NewReader(filetext)
	// 读取文件内容
	var encryptedData []byte
	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return sendCommonResponse(ctx, 500, "读取文件过程错误", nil)
		}
		if n == 0 || err == io.EOF {
			break
		}
		encryptedData = append(encryptedData, buf[:n]...)
	}
	// 使用解密函数解密文件数据
	decryptedData, err := sm4Decrypt([]byte(utils.CipherKey), []byte(utils.IV), encryptedData)
	if err != nil {
		return sendCommonResponse(ctx, 500, "文件解密错误", nil)
	}
	tempFile, err := os.CreateTemp("", file.Name)
	if err != nil {
		return sendCommonResponse(ctx, 500, "创建临时文件错误", nil)
	}
	defer tempFile.Close()
	// 将解密后的数据写入到临时文件中
	if _, err := tempFile.Write(decryptedData); err != nil {
		return sendCommonResponse(ctx, 500, "写入临时文件错误", nil)
	}
	ctx.Set("Content-Disposition", "attachment; filename="+file.Name)
	err = database.AddFileChangeLog(uid, fid, file.Name, time.Now().Format("2006-01-02 15:04:05"), "Download")
	if err != nil {
		return sendCommonResponse(ctx, 500, "日志插入失败", nil)
	}
	return ctx.SendFile(tempFile.Name())
}

func handleFileDelete(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)

	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	fidStr := ctx.Params("fid")
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		return sendCommonResponse(ctx, 500, "服务器内部错误", nil)
	}
	exist, err := database.GetFileAccess(uid, fid)
	if exist == false {
		return sendCommonResponse(ctx, 403, "该用户没有该文件权限", nil)
	}
	file, err := database.GetFileByFid(fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	err = os.Remove(file.Path)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "无法删除文件", nil)
	}
	err = database.DeleteFileByFid(fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "删除文件错误", nil)
	}
	err = database.AddFileChangeLog(uid, fid, file.Name, time.Now().Format("2006-01-02 15:04:05"), "Delete")
	if err != nil {
		return sendCommonResponse(ctx, 500, "日志插入失败", nil)
	}
	return sendCommonResponse(ctx, 200, "成功删除文件", nil)
}

func handleFileUpdate(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	fidStr := ctx.Params("fid")
	fmt.Println(fidStr)
	fid, err := strconv.Atoi(fidStr)
	fmt.Println(err)
	if err != nil {
		return sendCommonResponse(ctx, 500, "服务器内部错误", nil)
	}
	exist, err := database.GetFileAccess(uid, fid)
	if exist == false {
		return sendCommonResponse(ctx, 403, "该用户没有该文件权限", nil)
	}
	file, err := ctx.FormFile("file")
	fmt.Println("thereup;date?")
	if err != nil {
		return sendCommonResponse(ctx, 400, "", nil)
	}
	src, err := file.Open()
	if err != nil {
		return sendCommonResponse(ctx, 500, "无法打开文件", nil)
	}
	username, err := database.GetUserByUid(uid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	tmpfile, err := database.GetFileByFid(fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	destPath := tmpfile.Path
	err = os.Remove(destPath)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "无法删除更新前的文件", nil)
	}
	defer src.Close()
	hash := sm3.New()
	hashStr := hex.EncodeToString(hash.Sum(nil))
	NewdestPath := "./uploads/" + username.Username + "/" + file.Filename
	dest, err := os.Create(NewdestPath)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "无法保存更新后的文件", nil)
	}
	defer dest.Close()
	str, _ := generateRandomString(16)

	fmt.Println(NewdestPath, file.Filename)

	fileRecord := structs.File{
		Hash:      hashStr,
		Path:      NewdestPath,
		Uid:       uid,
		ShareCode: str,
		Name:      file.Filename,
	}
	err = database.UpdateFileByFid(fileRecord)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	err = database.GrantFileAccess(uid, fid)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	err = database.AddFileChangeLog(uid, fid, file.Filename, time.Now().Format("2006-01-02 15:04:05"), "Update")
	if err != nil {
		return sendCommonResponse(ctx, 500, "日志插入失败", nil)
	}
	return sendCommonResponse(ctx, 200, "成功更新文件", nil)
}

func searchFileByFilename(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	partname := ctx.Query("searchstring")
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页数错误", nil)
	}
	pageSize, err := strconv.Atoi(ctx.Query("pagesize", "10"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页大小错误", nil)
	}
	files, err := database.GetFileByPartialName(partname, page, pageSize, uid)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "", nil)
	}
	return sendCommonResponse(ctx, 200, "查询成功", map[string]interface{}{
		"total": len(files),
		"files": files,
	})
}

func getFileChangeLogByUid(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页数错误", nil)
	}
	pageSize, err := strconv.Atoi(ctx.Query("pagesize", "10"))
	if err != nil {
		return sendCommonResponse(ctx, 500, "接收页大小错误", nil)
	}
	//根据操作类型获取log，类型：Check,Delete,Update,Upload,Download,ALL
	operationtype := ctx.Query("operationtype", "Check")
	filelog, err := database.GetFileChangeLogByUid(uid, page, pageSize)
	if err != nil {
		return sendCommonResponse(ctx, 500, "获取文件日志错误", nil)
	}
	if operationtype == "All" {
		return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
			"total":   len(filelog),
			"filelog": filelog,
		})
	} else {
		var filelogs []structs.FileLog
		for _, log := range filelog {
			if log.Operation == operationtype {
				filelogs = append(filelogs, log)
			}
		}
		return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
			"total":   len(filelogs),
			"filelog": filelogs,
		})
	}
}
