package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"weqi_service/models"
	"weqi_service/serializer"
	"weqi_service/util"
)

type LoginService struct {
	Telephone string `json:"telephone" form:"telephone"`
	Password string `json:"password" form:"password"`
}

func Login(c *gin.Context) {

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

	var user models.User
	collection := models.Client.Collection("users")
	err := collection.FindOne(context.TODO(), bson.D{{"telephone", loginSrv.Telephone}}).Decode(&user)
	if err != nil {
		fmt.Println("Can't find user: ", err)
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "找不到此用户",
		})
		return
	}

	// 比较密码
	if user.Password != util.GenMD5Password(loginSrv.Password) {
		fmt.Println("Password doesn't match")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "密码错误",
		})
		return
	}

	// TODO：生成jwt Token返回给客户端


	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "登录成功",
		Data: user.UserID,
	})

}
