package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Log struct {
	Id        primitive.ObjectID `json:"id,omitempty"`
	Username  string             `json:"username,omitempty" validate:"required"`
	Status    string             `json:"status,omitempty" validate:"required"`
	State     string             `json:"state,omitempty" validate:"required"`
	City      string             `json:"city,omitempty" validate:"required"`
	Timestamp string             `json:"timestamp,omitempty" validate:"required"`
}
