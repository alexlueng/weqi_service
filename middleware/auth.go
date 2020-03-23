package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	"weqi_service/api"
	"weqi_service/auth"
	"weqi_service/serializer"

	//"weqi_service/serializer"
)

// 权限判断流程
// headers 中 Access-Token 为空，只能访问登录接口
// 不为空则进行解析
// 解析出错则踢出

// 其他函数里获取com_id,user_id,解析token即可

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		url := c.Request.URL.String()
		if url == "/api/v1/login" {
			c.Next()
			return
		}

		urlArr := strings.Split(url,"?")
		url = urlArr[0]

		token := c.GetHeader("Access-Token")
		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -2,
				Msg:  "token失效，请重新登录",
			})
			c.Abort()
			return
		}
		api.SmartPrint(claims)
		if claims.ExpiresAt == 0 {
			c.Abort()
			return
		}
		if time.Now().Unix() > claims.ExpiresAt {
			c.JSON(http.StatusOK, serializer.Response{
				Code: -2,
				Msg:  "token失效，请重新登录",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}