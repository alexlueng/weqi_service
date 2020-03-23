package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
	"weqi_service/models"
)

type Counts struct {
	NameField string
	Count     int64
}

func GetLatestID(field_name string) int64 {
	var c Counts
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", field_name}}).Decode(&c)
	if err != nil {
		fmt.Println("can't get ID")
		return 0
	}
	collection.UpdateOne(context.TODO(), bson.M{"name": field_name}, bson.M{"$set": bson.M{"count": c.Count + 1}})
	fmt.Printf("%s count: %d", field_name, c.Count)
	return c.Count + 1
}

func SmartPrint(i interface{}){
	var kv = make(map[string]interface{})
	vValue := reflect.ValueOf(i)
	vType :=reflect.TypeOf(i)
	for i:=0; i < vValue.NumField(); i++{
		kv[vType.Field(i).Name] = vValue.Field(i)
	}
	fmt.Println("获取到数据:")
	for k,v := range kv {
		fmt.Print(k)
		fmt.Print(":")
		fmt.Print(v)
		fmt.Println()
	}
}

func GenRandomDigitCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	var digitRunes = []rune("0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = digitRunes[rand.Intn(len(digitRunes))]
	}
	return string(b)
}

func GetIpAddress(c *gin.Context) string {
	realip := c.GetHeader("X-Real-Ip")
	if realip != "" {
		return realip
	}
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr != "" {
		idx := strings.Index(remoteAddr, ":")
		return remoteAddr[:idx]
	}
	ips := c.GetHeader("X-Forwarded-For")
	fmt.Println("X-Forwarded-For", ips)
	if ips != "" {
		iplist := strings.Split(ips, ",")
		return strings.TrimSpace(iplist[0])
	}
	return ""
}

type OrderCount struct {
	Count int64 `bson:"count"`
}

func GetTempOrderSN() string {
	// 000001 000002 000003
	// 先顺序生成数字 然后转成字符串，不足6位的用0补齐
	var coc OrderCount
	collection := models.Client.Collection("counters")
	err := collection.FindOne(context.TODO(), bson.D{{"name", "order"}}).Decode(&coc)
	if err != nil {
		fmt.Println("can't get OrderSN")
		return ""
	}
	sn := strconv.Itoa(int(coc.Count)+1)
	if len(sn) < 6 {
		fmt.Printf("len of sn: %d\n", len(sn))
		step := 6-len(sn)
		for i := 0; i < step; i++ {
			sn = "0" + sn
		}
	}
	current_date := time.Now().Format("20060102")
	sn = current_date + sn
	fmt.Println("Current OrderSN: ", sn)

	_, _  = collection.UpdateOne(context.TODO(), bson.M{"name": "order"}, bson.M{"$set": bson.M{"count": coc.Count + 1}})

	return sn
}
