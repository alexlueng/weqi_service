package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	//"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
	"weqi_service/models"
	"weqi_service/serializer"
)

type UserIDService struct {
	UserID int64 `json:"user_id"`
}

func ModuleList(c *gin.Context) {

	var uIDSrv UserIDService
	if err := c.ShouldBindJSON(&uIDSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	// 所有的页面都需要考虑分页，排序

	// 列出所有的模块
	var modules []models.Module
	collections := models.Client.Collection("modules")
	cur, err := collections.Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println("Can't get module list: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.Module
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't not decode into module: ", err)
			return
		}
		modules = append(modules, res)
	}
	moduleNums := make(map[int64]int64)
	// TODO：显示用户拥有模块的数量 module_instance
	collections = models.Client.Collection("module_instance")
	//collections.Find(context.TODO(), bson.D{{"user_id", userID}})
	filter := bson.M{}
	filter["user_id"] = uIDSrv.UserID
	for _, module := range modules {
		filter["module_id"] = module.ModuleID
		total, err := collections.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println("Can't get user module number: ", err)
			return
		}
		moduleNums[module.ModuleID] = total
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get module list",
		Data: map[string]interface{}{
			"modules" : modules,
			"module_nums" : moduleNums,
		},
	})
}

type ModuleService struct {
	ModuleID int64 `form:"module_id"`
	UserID int64 `form:"module_id"`
}

func ModuleDetail(c *gin.Context) {
	var moduleSrv ModuleService
	if err := c.ShouldBind(&moduleSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}

	instanceList := []models.ModuleInstance{}
	// 从Module_instance表中找出所有的模块实例
	collection := models.Client.Collection("module_instance")
	filter := bson.M{}
	filter["user_id"] = moduleSrv.UserID
	filter["module_id"] = moduleSrv.ModuleID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find user instances: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.ModuleInstance
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode module instance: ", err)
			return
		}
		instanceList = append(instanceList, res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get user instance list",
		Data: instanceList,
	})
}

type CreateModService struct {
	UserID int64 `json:"user_id" form:"user_id"`
	Domains []string `json:"domains" form:"domains"`
	Company string `json:"company" form:"company"`
	Developer string `json:"developer" form:"developer"`
	ModuleName string `json:"module_name" form:"module_name"`
	ModuleID int64 `json:"module_id" form:"module_id"`
}

func CreateModule(c *gin.Context) {

	var createModSrv CreateModService
	if err := c.ShouldBindJSON(&createModSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}

	SmartPrint(createModSrv)

	var module models.ModuleInstance
	module.ModuleName = createModSrv.ModuleName
	module.Domains = createModSrv.Domains
	module.Company = createModSrv.Company
	module.ModuleID = createModSrv.ModuleID

	// 生成一个instance id
	module.InstanceID = GetLatestID("module_instance")
	// com id 从user表中获取
	//userID := c.Get("userID")


	collection := models.Client.Collection("users")
	filter := bson.M{}
	//filter["user_id"] = userID
	var user models.User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		fmt.Println("Can't get user information: ", err)
		return
	}
	SmartPrint(user)
	module.Admin = user.Username
	module.ComID = user.ComID
	module.CreateAt = time.Now().Unix()
	module.ExpireAt = time.Now().Unix() + (7 * 86400)
	module.Using = true
	module.IsTry = true

	collection = models.Client.Collection("module_instance")
	insertResult, err := collection.InsertOne(context.TODO(), module)
	if err != nil {
		fmt.Println("Create module failed: ", err)
		return
	}
	fmt.Println("insert result: ", insertResult.InsertedID)

	// TODO：把这条记录同时添加到对应模块实例的company表中，目前是进销存模块
	// 这是两个服务之间的通信，需要调用新的接口
	AddRecordToRemoteModule("localhost:3000/system/add_record", module)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Module create success",
	})
}

func AddRecordToRemoteModule(url string, instance models.ModuleInstance) {
	b, err := json.Marshal(instance)
	if err != nil {
		fmt.Println("fail to convert struct to json: ", err)
		return
	}
	fmt.Println(string(b))
	buffer:= bytes.NewBuffer(b)
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

type InstanceService struct {
	InstanceID int64 `json:"instance_id" form:"instance_id"`
}

func InstanceDetail(c *gin.Context) {
	var insSrv InstanceService
	if err := c.ShouldBindJSON(&insSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	collection := models.Client.Collection("module_instance")
	var instance models.ModuleInstance
	err := collection.FindOne(context.TODO(), bson.D{{"instance_id", insSrv.InstanceID}}).Decode(&instance)
	if err != nil {
		fmt.Println("Can't find module instance: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get module instance",
		Data: instance,
	})
}