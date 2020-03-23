package models

type Domain struct {
	ID int64 `json:"id" bson:"id"`
	InstanceID int64 `json:"instance_id" bson:"instance_id"`
	Value string `json:"value" bson:"value"`
}
