package models

type Order struct {
	OrderID int64 `json:"order_id" bson:"order_id"`
	OrderNO string `json:"order_no" bson:"order_no"`
	UserID int64 `json:"user_id" bson:"user_id"`
	InstanceID int64 `json:"instance_id" bson:"instance_id"`
	ModuleName string `json:"module_name" bson:"module_name"`
	Amount float64 `json:"amount" bson:"amount"`
	CreateAt int64 `json:"create_at" bson:"create_at"`
	ExpireAt int64 `json:"expire_at" bson:"expire_at"`
	Status int64 `json:"status" bson:"status"` // 0生成预付单 1收到微信预付单 2付款成功
}
