package middleware

import (
	"github.com/gin-gonic/gin"
	//"weqi_service/serializer"
)


// CurrentUser 获取登录用户
func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// uid := session.Get("user_id")
		// if uid != nil {
		// 	user, err := models.GetUser(uid)
		// 	if err == nil {
		// 		c.Set("user", &user)
		// 	}
		// }
		c.Next()
	}
}

// AuthRequired 需要登录
//func AuthRequired() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		if user, _ := c.Get("user"); user != nil {
//			if _, ok := user.(*models.User); ok {
//				c.Next()
//				return
//			}
//		}
//
//		c.JSON(200, serializer.CheckLogin())
//		c.Abort()
//	}
//}
