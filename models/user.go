package models

type User struct {
	UserID int64 `json:"user_id" bson:"user_id"`
	Username string `json:"username" bson:"username"`
	ComID int64 `json:"com_id" bson:"com_id"`
	Password string `json:"password" bson:"password"`
	Level int64 `json:"level" bson:"level"` // 客户等级
	Telephone string `json:"telephone" bson:"telephone"`
	CreateAt int64 `json:"create_at" bson:"create_at"`

	// Email Birthday region 等信息
}
