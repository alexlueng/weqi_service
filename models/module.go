package models

// 模块表
type Module struct {
	ModuleID  int64   `json:"id" bson:"id"`
	Name      string  `json:"name" bson:"name"`
	Price     float64 `json:"price" bson:"price"`
	Developer string  `json:"developer" bson:"developer"`
}

// 模块实例表，记录用户与模块之间的对应关系
type ModuleInstance struct {
	InstanceID int64    `json:"id" bson:"id"`
	ModuleID   int64    `json:"module_id" bson:"module_id"`
	ModuleName string   `json:"module_name" bson:"module_name"` // 模块名，这个是管理员设置的
	UserID     int64    `json:"user_id" bson:"user_id"`
	Admin      string   `json:"admin" bson:"admin"` // 超级管理员
	Telephone string 	`json:"telephone" bson:"telephone"` // 超级管理员电话
	ComID      int64    `json:"com_id" bson:"com_id"`
	Company    string   `json:"company_name" bson:"company_name"`
	Domains    []string `json:"domains" bson:"domains"`
	CreateAt   int64    `json:"create_at" bson:"create_at"`       // 创建时间
	LastPaidAt int64    `json:"last_paid_at" bson:"last_paid_at"` // 上次支付时间
	ExpireAt   int64    `json:"expire_at" bson:"expire_at"`       // 过期时间

	Using bool `json:"using" bson:"using"`   // 是否正在使用
	IsTry bool `json:"is_try" bson:"is_try"` // 是否试用
}
