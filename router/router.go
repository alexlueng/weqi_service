package router

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
	"weqi_service/api"
	"weqi_service/middleware"
	"weqi_service/wxpay"

	"github.com/gin-contrib/zap"
	"go.uber.org/zap"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.Default()
	// 中间件, 顺序不能改

	// 允许跨域
	r.Use(middleware.Cors())

	//设置默认路由当访问一个错误网站时返回
	//r.NoRoute(api.NotFound)
	logger, _ := zap.NewProduction()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	//r.Use(middleware.CheckAuth())
	r.Use(sessions.Sessions("mysession", sessions.NewCookieStore([]byte("secret"))))

	//使用以下gin提供的Group函数为不同的API进行分组
	v1 := r.Group("/api/v1")
	{

		v1.POST("/register", api.Register)
		v1.POST("/sendcode", api.SendCode)
		v1.POST("/checkcode", api.CheckCode)

		v1.POST("/login", api.Login)
		v1.POST("/logout", api.Logout)

		v1.POST("/module/list", api.ModuleList)
		v1.POST("/module/detail", api.ModuleDetail)
		v1.POST("/module/create", api.CreateModule)
		v1.POST("/module_instance/detail", api.InstanceDetail)
		v1.POST("/module_instance/stop", api.StopInstance)
		v1.POST("/module_instance/start", api.StartInstance)
		v1.POST("/module_instance/delete", api.DeleteInstance)

		v1.POST("/user/info", api.UserInfo)
		v1.POST("/user/resetpassword", api.ResetUserPassword)
		v1.POST("/user/resetcode", api.SendResetCode)

		v1.POST("/wxpay/cheopenid", wxpay.GetOpenIDURL)
		v1.GET("/wxpay/getopenid", wxpay.Callback)
		v1.POST("/wxpay/getprepayid", wxpay.GetPrepayID)
		v1.POST("/wxpay/wxpaycallback", wxpay.WxpayCallback)
		v1.POST("/wxpay/CheckOrderStatus", wxpay.CheckOrderStatus)
		v1.POST("/wxpay/paidhistory", wxpay.UserPaidHistory)
	}

	return r
}
