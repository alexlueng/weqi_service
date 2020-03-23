package conf

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"weqi_service/cache/redis"
	"weqi_service/models"
	"weqi_service/util"
)

var RestrictIP map[string]map[string]interface{}
var RestrictLogin map[string]map[string]interface{}
// Init 初始化配置项
func Init() {
	// 从本地读取环境变量
	// 使用默认配置路径 .env 读取不到配置文件
	// 所以在这里先获取到当前项目目录，拼接出 .env 的绝对路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load(dir + "/.env")
	godotenv.Load() // .env

	// 设置日志级别
	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	fmt.Println("LOG_LEVEL :", os.Getenv("LOG_LEVEL"))

	// 连接数据库
	models.Database(os.Getenv("URI"), os.Getenv("DBNAME"))
	redis.RedisClient(os.Getenv("REDIS_ADDR"))

	RestrictIP = make(map[string]map[string]interface{})
	RestrictLogin = make(map[string]map[string]interface{})
}
