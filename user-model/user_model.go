package user_model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserId          string             `bson:"user_id"`
	TokensRemaining int                `bson:"tokens_remaining"`
	AuthKey         string             `bson:"auth_key"`
	IsAdmin         bool               `bson:"is_admin"`
}
