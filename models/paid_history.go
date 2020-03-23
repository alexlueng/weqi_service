package models

type PaidHistory struct {
	ID         int64   `json:"id" bson:"id"`
	UserID     int64   `json:"user_id" bson:"user_id"`
	InstanceID int64   `json:"instance_id" bson:"instance_id"`
	Amount     float64 `json:"amount" bson:"amount"`
	PaidAt     int64   `json:"paid_at" bson:"paid_at"`
}
