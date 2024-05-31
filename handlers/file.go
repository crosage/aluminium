package handlers

import (
	"chain/database"
	"chain/structs"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"io"
	"os"
	"strconv"
)

func handleFileUpload(ctx *fiber.Ctx) error {
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

	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return sendCommonResponse(ctx, 500, "无法生成文件哈希", nil)
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return sendCommonResponse(ctx, 500, "无法重置文件指针", nil)
	}

	destPath := "./uploads/" + file.Filename
	dest, err := os.Create(destPath)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "无法保存文件", nil)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
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
		return sendCommonResponse(ctx, 200, "", nil)
	}
}

func getUserCreatedFiles(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	user := getSessionUser(ctx)
	uid := user.Uid
	files, err := database.GetFilesCreatedByUid(uid)
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
	files, err := database.GetFilesAvailableByUid(uid)
	if err != nil {
		fmt.Println(err)
		return sendCommonResponse(ctx, 500, "", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
		"total": len(files),
		"files": files,
	})
}

func getFileByFid(ctx *fiber.Ctx) error {
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
		return sendCommonResponse(ctx, 500, "获取文件错误", nil)
	}
	filePath := file.Path
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return sendCommonResponse(ctx, 404, "文件不存在", nil)
	}
	ctx.Set("Content-Disposition", "attachment; filename="+filePath)
	return ctx.SendFile(filePath)
}
