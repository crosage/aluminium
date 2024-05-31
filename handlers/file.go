package handlers

import (
	"chain/database"
	"chain/structs"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gofiber/fiber/v2"
	"io"
	"os"
)

func handleFileUpload(ctx *fiber.Ctx) error {
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
		return sendCommonResponse(ctx, 500, "", nil)
	}
	defer dest.Close()

	fileRecord := structs.File{
		Hash: hashStr,
		Path: destPath,
		Uid:  uid,
	}

	err = database.SaveFile(fileRecord)
	if err != nil {
		return sendCommonResponse(ctx, 500, "", nil)
	}
	return sendCommonResponse(ctx, 200, "成功上传文件", nil)
}
