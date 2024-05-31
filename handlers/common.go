package handlers

import (
	"chain/structs"
	"chain/utils"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
)

func InitHandlers(app *fiber.App) {

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	app.Post("/api/user", userRegister)
	app.Post("/api/user/login", userLogin)

	// Register middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			Key: []byte(utils.JwtKey),
		},
		ContextKey: "user",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return sendCommonResponse(ctx, 401, "过期或非法JWT", map[string]interface{}{
				"path": ctx.Path(),
			})
		},
	}))

	app.Get("/api/user", getAllUsers)
	app.Post("/api/user/:uid", updateUser)
	app.Delete("/api/user/:uid", deleteUser)
	app.Post("/api/file/upload", handleFileUpload)
	app.Post("/api/file/check-share", checkShareCode)
	app.Get("/api/file/created-files", getUserCreatedFiles)
	app.Get("/api/file/available-files", getUserAvailableFiles)
	app.Get("/api/file/:fid", getFileByFid)
}

func sendCommonResponse(ctx *fiber.Ctx, code int, message string, data map[string]interface{}) error {
	response := structs.Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	json, err := jsoniter.Marshal(response)
	if err != nil {
		// THIS SHOULD NOT HAPPEN
		// If this happens, just stop the server and wait for further investigation

	}
	return ctx.Status(code).Send(json)
}

func validatePermission(ctx *fiber.Ctx) bool {
	userLocal := ctx.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	userType := int(claims["usertype"].(float64))
	fmt.Println("*********")
	fmt.Println(userType)
	return userType > 0
}

func getSessionUser(ctx *fiber.Ctx) structs.User {
	userLocal := ctx.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	user := structs.User{}

	user.Type = int(claims["usertype"].(float64))
	user.Uid = int(claims["uid"].(float64))
	user.Username = claims["username"].(string)
	return user
}

func generateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(randomBytes)[:length], nil
}
