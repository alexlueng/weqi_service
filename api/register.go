package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"time"
	"weqi_service/models"
	"weqi_service/serializer"
	"weqi_service/util"

	"weqi_service/cache/redis"
)

type RegisterService struct {
	Username    string `json:"username" form:"username"`
	Password    string `json:"password" form:"password"`
	PassConform string `json:"pass_confirm" form:"pass_confirm"`
	Telephone   string `json:"telephone" form:"telephone"`
}

func Register(c *gin.Context) {

	var regSrv RegisterService
	if err := c.ShouldBind(&regSrv); err != nil {
		fmt.Println("register request err: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	if regSrv.Password != regSrv.PassConform {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "两次输入的密码不一样",
		})
		return
	}

	user := models.User{}

	// TODO：需要一个全局唯一ID生成算法
	user.UserID = GetLatestID("user")

	// TODO:password 需要经过加密
	// 用户名不用检测重名 因为手机号是唯一的
	user.Password = util.GenMD5Password(regSrv.Password)

	// 为客户生成一个ComID, 这个ComID会关联到客户开通的所有其他模块
	user.ComID = GetLatestID("com")
	user.Level = 1 // 初始等级为1
	user.CreateAt = time.Now().Unix()
	user.Username = regSrv.Username
	user.Telephone = regSrv.Telephone

	SmartPrint(user)

	collection := models.Client.Collection("users")
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		fmt.Println("Can't insert into user table: ", err)
		return
	}
	fmt.Println("insert user table success: ", insertResult.InsertedID)

	// 注册成功，重定向到登录页
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "注册成功",
	})
}

type TelephoneService struct {
	Telephone string `json:"telephone"`
	Code string `json:"code"`
}

func SendCode(c *gin.Context) {
	// 获取手机号
	var teleSrv TelephoneService
	if err := c.ShouldBindJSON(&teleSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	// 检查手机号是否已经注册过
	var user models.User
	collections := models.Client.Collection("users")
	err := collections.FindOne(context.TODO(), bson.D{{"telephone", teleSrv.Telephone}}).Decode(&user)
	if err == nil {
		// 说明手机号已经注册过了
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "手机号已经注册",
		})
		return
	}
	// 生成6位数验证码，保存到redis中
	code := GenRandomDigitCode(6)
	redis.Client.Set(teleSrv.Telephone, code, 1 * time.Minute) //1分钟过期

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "get code",
		Data: map[string]string{
			"code" : code,
		},
	})
}

func CheckCode(c *gin.Context) {
	var teleSrv TelephoneService
	if err := c.ShouldBindJSON(&teleSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}
	SmartPrint(teleSrv)
	code := strings.Trim(strings.Split(redis.Client.Get(teleSrv.Telephone).String(),":")[1], " ")
	fmt.Println("get code from redis: ", code)
	if code == "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "验证码已经过期",
		})
		return
	}
	if code != teleSrv.Code {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "验证码不正确",
		})
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "验证成功",
	})
	return
}
