package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"weqi_service/models"
	"weqi_service/serializer"
)

func UserInfo(c *gin.Context) {
	userID := 1
	collection := models.Client.Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.D{{"user_id", userID}}).Decode(&user)
	if err != nil {
		fmt.Println("Can't find user: ", err)
		return
	}
	collection = models.Client.Collection("module_instance")
	var instances []models.ModuleInstance
	cur, err := collection.Find(context.TODO(), bson.D{{"user_id", userID}})
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
		instances = append(instances,res)
	}

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "Get module instance",
		Data: map[string]interface{}{
			"user" : user,
			"instances" : instances,
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