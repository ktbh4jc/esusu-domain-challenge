package user_db

import (
	"context"
	"fmt"
	"maas/loggers"
	"os"

	error_types "maas/error-types"
	user_model "maas/user-model"
	user_service "maas/user-service"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBUserRepository struct {
	client *mongo.Client
	ctx    *context.Context
}

var _ user_service.UserRepository = &MongoDBUserRepository{}

func NewMongoDBUserRepository(client *mongo.Client, ctx *context.Context) *MongoDBUserRepository {
	return &MongoDBUserRepository{client: client, ctx: ctx}
}

func (m *MongoDBUserRepository) ResetDb() ([]interface{}, error) {
	// ensure local environment
	if os.Getenv("ENV_NAME") != "local" {
		return nil, &error_types.BadEnvironmentError{
			Err: fmt.Errorf("tried to reset db in environment %s", os.Getenv("ENV_NAME")),
		}
	}

	// reset the db
	database := m.client.Database("maas")
	maas_users := database.Collection("maas_users")

	// Drop everything
	if err := maas_users.Drop(*m.ctx); err != nil {
		return nil, err
	}

	// Insert new data
	insertResult, err := maas_users.InsertMany(*m.ctx, user_model.DefaultUsers)
	if err != nil {
		return nil, err
	}

	// Return data to caller
	return insertResult.InsertedIDs, nil
}

func (m *MongoDBUserRepository) User(id string) (*user_model.User, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: objectId}}

	opts := options.FindOne()

	var user user_model.User

	err = maas_users_collection.FindOne(*m.ctx, filter, opts).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *MongoDBUserRepository) AllUsers() ([]user_model.User, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")

	cursor, err := maas_users_collection.Find(*m.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var users_output []user_model.User

	if err := cursor.All(*m.ctx, &users_output); err != nil {
		return nil, err
	}

	return users_output, nil
}

// Ping implements user_service.UserRepository.Ping
func (m *MongoDBUserRepository) Ping() error {
	if err := m.client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return &error_types.MongoConnectionError{Err: err}
	}
	loggers.InfoLog.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}
