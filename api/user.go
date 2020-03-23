package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"weqi_service/models"
	"weqi_service/serializer"
	"weqi_service/util"

	"encoding/json"
	"weqi_service/cache/redis"
)

type UserInfoService struct {
	UserID int64 `json:"user_id"`
}

func UserInfo(c *gin.Context) {
	var usrInfoSrv UserInfoService
	if err := c.ShouldBindJSON(&usrInfoSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数错误",
		})
	}
	collection := models.Client.Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.D{{"user_id", usrInfoSrv.UserID}}).Decode(&user)
	if err != nil {
		fmt.Println("Can't find user: ", err)
		return
	}
	collection = models.Client.Collection("module_instance")
	var instances []models.ModuleInstance
	cur, err := collection.Find(context.TODO(), bson.D{{"user_id", usrInfoSrv.UserID}})
	if err != nil {
		fmt.Println("Can't find user's module instance: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.ModuleInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode model instance: ", err)
			return
		}
		instances = append(instances, res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get module instance",
		Data: map[string]interface{}{
			"user":      user,
			"instances": instances,
		},
	})
}

type ResetPasswordService struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
	UserID      int64  `json:"user_id"`
}

// TODO: test this function
func ResetUserPassword(c *gin.Context) {
	var resetPassSrv ResetPasswordService
	if err := c.ShouldBindJSON(&resetPassSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数错误",
		})
	}

	collection := models.Client.Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.D{{"user_id", resetPassSrv.UserID}}).Decode(&user)
	if err != nil {
		fmt.Println("can't find reset password user: ", err)
		return
	}

	code := strings.Trim(strings.Split(redis.Client.Get(user.Telephone).String(),":")[1], " ")
	if code == "" {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "验证码已经过期",
		})
		return
	}
	if code != resetPassSrv.Code {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "验证码不正确",
		})
		return
	}

	if util.GenMD5Password(resetPassSrv.NewPassword) == user.Password {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "不能与原密码相同",
		})
		return
	}

	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"user_id", user.UserID}}, bson.M{
		"$set": bson.M{"password": util.GenMD5Password(resetPassSrv.NewPassword)}})
	if err != nil {
		fmt.Println("Can't update user's password: ", err)
		return
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "更新密码成功",
	})

	// 远程更新服务的超级管理员密码
	UpdateSuperAdminPassword("http://127.0.0.1:3000/adminpasswd_update", user)


}

func UpdateSuperAdminPassword(url string, admin models.User) {
	b, err := json.Marshal(admin)
	if err != nil {
		fmt.Println("fail to convert struct to json: ", err)
		return
	}
	//fmt.Println(string(b))
	buffer := bytes.NewBuffer(b)

	request, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		fmt.Printf("http.NewRequest%v", err)
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")  //添加请求头
	client := http.Client{} //创建客户端
	resp, err := client.Do(request.WithContext(context.TODO())) //发送请求
	if err != nil {
		fmt.Printf("client.Do %v", err)
		return
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll %v", err)
		return
	}
	fmt.Println(string(respBytes))
}




// 发送验证码
func SendResetCode(c *gin.Context) {

	//session := sessions.Default(c)
	// 获取手机号
	var teleSrv TelephoneService
	if err := c.ShouldBindJSON(&teleSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "参数解释错误",
		})
		return
	}

	// 生成6位数验证码，保存到redis中
	code := GenRandomDigitCode(6)
	redis.Client.Set(teleSrv.Telephone, code, 2 * time.Minute) //2分钟过期

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "get code",
		Data: map[string]string{
			"code" : code,
		},
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)

	userID := session.Get("userID")
	fmt.Println("Get user id: ", userID)
	if userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("userID")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	userID = session.Get("userID")
	fmt.Println("Get user id: ", userID)
	if userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
