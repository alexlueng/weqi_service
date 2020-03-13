package api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"reflect"
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