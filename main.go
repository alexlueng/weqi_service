package main

import (
	"weqi_service/conf"
	"weqi_service/router"
)

func main() {
	// 从配置文件读取配置
	conf.Init()
	// 装载路由
	r := router.InitRouter()
	r.Run(":3002")
}
