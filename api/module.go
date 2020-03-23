package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

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

type ModuleService struct {
	ModuleID int64 `json:"module_id" form:"module_id"`
	UserID int64 `json:"user_id" form:"user_id"`
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

func ModuleDetail(c *gin.Context) {
	var moduleSrv ModuleService
	if err := c.ShouldBind(&moduleSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}

	fmt.Println("modele detail service: ", moduleSrv)

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

func GetDefaultDomain(c *gin.Context) {

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

	var instance models.ModuleInstance
	instance.ModuleName = createModSrv.ModuleName
	instance.Domains = createModSrv.Domains

	instance.Company = createModSrv.Company
	instance.ModuleID = createModSrv.ModuleID

	// 生成一个instance id
	instance.InstanceID = GetLatestID("module_instance")
	// com id 从user表中获取
	//userID := c.Get("userID")
	collection := models.Client.Collection("users")
	filter := bson.M{}
	filter["user_id"] = createModSrv.UserID
	var user models.User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		fmt.Println("Can't get user information: ", err)
		return
	}
	//SmartPrint(user)

	instance.UserID = user.UserID
	instance.Admin = user.Username
	instance.Password = user.Password
	instance.ComID = user.ComID
	instance.Telephone = user.Telephone
	instance.CreateAt = time.Now().Unix()
	instance.ExpireAt = time.Now().Unix() + (7 * 86400) // 首次开通用户有7天的免费使用时间
	instance.Using = true
	instance.IsTry = true

	// 查找模块的价格
	collection = models.Client.Collection("modules")
	var module models.Module
	err = collection.FindOne(context.TODO(), bson.D{{"id", createModSrv.ModuleID}}).Decode(&module)
	if err != nil {
		fmt.Println("Can't get module: ", err)
		return
	}
	instance.Price = module.Price

	// TODO: 域名要独立存到一张表中，然后还要做域名重复的判定，如果没有填写域名，则为它提供一个默认的域名
	// TODO: 如果是自己填写的域名，需要让它指向服务器的ip
	if len(createModSrv.Domains) == 0 { // 没有填写域名
		instance.Domains = append(instance.Domains, strconv.Itoa(int(instance.InstanceID)) + os.Getenv("DOMAINOFJXC"))
	}

	collection = models.Client.Collection("Domain")
	for _, domain := range instance.Domains {
		var existDomain models.Domain
		err := collection.FindOne(context.TODO(), bson.M{"value": bson.M{"$eq": domain}}).Decode(&existDomain)
		if err == nil {
			// 说明已经存在此域名
			c.JSON(http.StatusOK, serializer.Response{
				Code: -1,
				Msg:  "填写的域名已经存在",
			})
			return
		}
	}
	for _, domain := range instance.Domains {
		var d models.Domain
		d.ID = GetLatestID("domain")
		d.InstanceID = instance.InstanceID
		d.Value = domain
		insertResult, err := collection.InsertOne(context.TODO(), d)
		if err != nil {
			fmt.Println("Can't insert domain: ", err)
			return
		}
		fmt.Println("insert result: ", insertResult.InsertedID)
	}



	SmartPrint(instance)

	collection = models.Client.Collection("module_instance")
	insertResult, err := collection.InsertOne(context.TODO(), instance)
	if err != nil {
		fmt.Println("Create module failed: ", err)
		return
	}
	fmt.Println("insert result: ", insertResult.InsertedID)

	// TODO：把这条记录同时添加到对应模块实例的company表中，目前是进销存模块
	// 这是两个服务之间的通信，需要调用新的接口
	AddRecordToRemoteModule("http://127.0.0.1:3000/add_superadmin", instance)

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
	SmartPrint(insSrv)
	collection := models.Client.Collection("module_instance")
	var instance models.ModuleInstance
	err := collection.FindOne(context.TODO(), bson.D{{"id", insSrv.InstanceID}}).Decode(&instance)
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

func StopInstance(c *gin.Context) {
	var updSrv InstanceService
	if err := c.ShouldBindJSON(&updSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	SmartPrint(updSrv)
	collection := models.Client.Collection("module_instance")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"id", updSrv.InstanceID}}, bson.M{"$set": bson.M{"using": false}})
	if err != nil {
		fmt.Println("Can't stop instance: ", err)
		return
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)

	// TODO：实际的模块实例也不能访问了
	// TODO: 在对应进销存服务的域名表中加一个valid字段，表示当前域名是否有效，这里停用之后要到那边修改

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Stop module instance",
	})
}

func StartInstance(c *gin.Context) {
	var updSrv InstanceService
	if err := c.ShouldBindJSON(&updSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	SmartPrint(updSrv)
	collection := models.Client.Collection("module_instance")
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"id", updSrv.InstanceID}}, bson.M{"$set": bson.M{"using": true}})
	if err != nil {
		fmt.Println("Can't start instance: ", err)
		return
	}
	fmt.Println("Update result: ", updateResult.UpsertedID)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Start module instance",
	})
}

func DeleteInstance(c *gin.Context) {
	var updSrv InstanceService
	if err := c.ShouldBindJSON(&updSrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "Params error",
		})
		return
	}
	SmartPrint(updSrv)
	collection := models.Client.Collection("module_instance")
	DeleteResult, err := collection.DeleteOne(context.TODO(), bson.D{{"id", updSrv.InstanceID}})
	if err != nil {
		fmt.Println("Can't start instance: ", err)
		return
	}
	fmt.Println("Delete result: ", DeleteResult.DeletedCount)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Delete module instance",
	})
}