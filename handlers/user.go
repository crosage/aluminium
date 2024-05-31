package handlers

import (
	"chain/database"
	"chain/structs"
	"chain/utils"
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/json-iterator/go"
	"strconv"
)

func userRegister(ctx *fiber.Ctx) error {
	user := structs.User{}
	err := jsoniter.Unmarshal(ctx.Body(), &user)
	//println(err)
	if err != nil || len(user.Username) == 0 || len(user.Password) == 0 {
		return sendCommonResponse(ctx, 403, "非法输入", nil)
	}

	user.Type = 0

	err = database.CreateUser(user)
	if err != nil {
		return sendCommonResponse(ctx, 403, "非法输入", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", nil)
}

func userLogin(ctx *fiber.Ctx) error {
	user := structs.User{}
	err := jsoniter.Unmarshal(ctx.Body(), &user)
	if err != nil {
		return sendCommonResponse(ctx, 403, "非法输入", nil)
	}
	queriedUser, err := database.GetPassHash(user.Username)
	if err == sql.ErrNoRows {
		return sendCommonResponse(ctx, 403, "用户不存在", nil)
	}
	if err != nil {
		return sendCommonResponse(ctx, 500, "内部服务器错误", nil)
	}
	if utils.GeneratePassHash(user.Password) == queriedUser.Password {
		claims := jwt.MapClaims{
			"uid":      queriedUser.Uid,
			"username": user.Username,
			"usertype": queriedUser.Type,
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(utils.JwtKey))
		if err != nil {
			return sendCommonResponse(ctx, 500, "内部服务器错误", nil)
		}
		return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
			"uid":      queriedUser.Uid,
			"username": user.Username,
			"usertype": queriedUser.Type,
			"token":    token,
		})
	} else {
		return sendCommonResponse(ctx, 403, "账号或密码错误", nil)
	}
}

func getAllUsers(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	getSessionUser(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	total, users, err := database.GetAllUsers()
	if err != nil {
		return sendCommonResponse(ctx, 403, "非法输入", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", map[string]interface{}{
		"total": total,
		"users": users,
	})
}

func deleteUser(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	uid, err := strconv.Atoi(ctx.Params("uid"))
	if err != nil {
		return sendCommonResponse(ctx, 403, "非法路径", nil)
	}
	err = database.DeleteUser(uid)
	return sendCommonResponse(ctx, 200, "成功", nil)
}

func updateUser(ctx *fiber.Ctx) error {
	hasPermission := validatePermission(ctx)
	if !hasPermission {
		return sendCommonResponse(ctx, 403, "无权限", nil)
	}
	uid, err := strconv.Atoi(ctx.Params("uid"))
	if err != nil {
		return sendCommonResponse(ctx, 403, "非法路径", nil)
	}
	user := structs.User{}
	err = jsoniter.Unmarshal(ctx.Body(), &user)
	if err != nil || user.Uid != uid {
		return sendCommonResponse(ctx, 403, "非法输入", nil)
	}
	err = database.UpdateUser(user)
	if err != nil {
		return sendCommonResponse(ctx, 500, "内部服务器错误", nil)
	}
	return sendCommonResponse(ctx, 200, "成功", nil)
}
