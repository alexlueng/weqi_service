package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"time"
	"weqi_service/auth"
	"weqi_service/conf"
	"weqi_service/models"
	"weqi_service/serializer"
	"weqi_service/util"

	"gopkg.in/go-playground/validator.v9"
)

type LoginService struct {
	Telephone string `json:"telephone" form:"telephone" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

// TODO: 一个账号当日不能无限制地登录，需要加入登录次数限制

func Login(c *gin.Context) {

	fmt.Println("Get request uri: ", c.Request.Host + c.Request.RequestURI)

	validate := validator.New()

	var loginSrv LoginService
	if err := c.ShouldBind(&loginSrv); err != nil {
		fmt.Println("login request err: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	SmartPrint(loginSrv)
	loginSrv.Telephone = strings.Trim(loginSrv.Telephone, " ")
	// 验证前端传过来的参数
	err := validate.Struct(loginSrv)
	if err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
			Data: err.Error(),
		})
		return
	}

	var user models.User
	collection := models.Client.Collection("users")
	err = collection.FindOne(context.TODO(), bson.D{{"telephone", loginSrv.Telephone}}).Decode(&user)
	if err != nil {
		fmt.Println("Can't find user: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "用户未注册",
		})
		return
	}

	SmartPrint(user)

	// 比较密码
	if user.Password != util.GenMD5Password(loginSrv.Password) {

		if _, ok := conf.RestrictLogin[user.Telephone]; ok { // 已经输错过密码了

			if conf.RestrictLogin[user.Telephone]["expire_time"].(int64) < time.Now().Unix() {
				// 更新密码可输次数和过期时间
				conf.RestrictLogin[user.Telephone]["times"] = 3 // 剩余尝试次数
				conf.RestrictLogin[user.Telephone]["expire_time"] = time.Now().Unix() + 86400
			}

			if conf.RestrictLogin[user.Telephone]["times"].(int) == 0 {
				// 当日尝试输入密码的次数已经用完
				// 账号被冻结，请联系管理员
				// 用户表中加一个状态
				fmt.Println("账号被冻结")
			}
			conf.RestrictLogin[user.Telephone]["times"] = conf.RestrictLogin[user.Telephone]["times"].(int) - 1
		} else {
			conf.RestrictLogin[user.Telephone] = make(map[string]interface{})
			conf.RestrictLogin[user.Telephone]["times"] = 3 // 剩余尝试次数
			conf.RestrictLogin[user.Telephone]["expire_time"] = time.Now().Unix() + 86400
		}

		fmt.Println("Password doesn't match")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "密码错误",
			Data: map[string]interface{}{
				"try_time" : conf.RestrictLogin[user.Telephone]["times"],
			},
		})
		return
	}

	// TODO：生成jwt Token返回给客户端
	token, err := auth.GenerateToken(user.Username, user.Telephone)
	if err != nil {
		fmt.Println("get user token error: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "登录成功",
		Data: map[string]interface{}{
			"user_id" : user.UserID,
			"token" : token,
		},
	})

}
