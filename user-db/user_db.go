package user_db

import (
	"context"
	"errors"
	"fmt"
	"maas/loggers"
	meme_service "maas/meme-service"
	"os"

	auth_service "maas/auth-service"
	error_types "maas/error-types"
	"maas/models"
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

// Normally I would say you probably don't want both the same repo serving both users and auth
// to be the same, but since our auth service is just checking a plaintext string and a bool in
//  our db, I figured this works.
var _ user_service.UserRepository = &MongoDBUserRepository{}
var _ auth_service.AuthRepository = &MongoDBUserRepository{}
var _ meme_service.UserRepository = &MongoDBUserRepository{}

func NewMongoDBUserRepository(client *mongo.Client, ctx *context.Context) *MongoDBUserRepository {
	return &MongoDBUserRepository{client: client, ctx: ctx}
}

func (m *MongoDBUserRepository) NewUser(user models.User) (interface{}, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")

	insertResult, err := maas_users_collection.InsertOne(*m.ctx, user)
	if err != nil {
		return nil, err
	}
	return insertResult.InsertedID, nil
}

func (m *MongoDBUserRepository) UserByAuthHeader(auth string) (*models.User, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")

	filter := bson.D{{Key: "auth_key", Value: auth}}

	opts := options.FindOne()

	var user models.User

	err := maas_users_collection.FindOne(*m.ctx, filter, opts).Decode(&user)
	if err != nil {
		// I want to avoid using mongo-specific errors up the chain
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &error_types.UnableToLocateDocumentError{Err: err}
		}
		return nil, err
	}

	return &user, nil
}

func (m *MongoDBUserRepository) UpdateUser(id string, user *models.User) error {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")
	hexId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = maas_users_collection.ReplaceOne(
		*m.ctx,
		bson.M{"_id": hexId},
		user,
	)
	if err != nil {
		return err
	}
	return nil
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
	insertResult, err := maas_users.InsertMany(*m.ctx, models.DefaultUsers)
	if err != nil {
		return nil, err
	}

	// Return data to caller
	return insertResult.InsertedIDs, nil
}

func (m *MongoDBUserRepository) User(id string) (*models.User, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: objectId}}

	opts := options.FindOne()

	var user models.User

	err = maas_users_collection.FindOne(*m.ctx, filter, opts).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *MongoDBUserRepository) AllUsers() ([]models.User, error) {
	database := m.client.Database("maas")
	maas_users_collection := database.Collection("maas_users")

	cursor, err := maas_users_collection.Find(*m.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var users_output []models.User

	if err := cursor.All(*m.ctx, &users_output); err != nil {
		return nil, err
	}

	return users_output, nil
}

func (m *MongoDBUserRepository) Ping() error {
	if err := m.client.Database("admin").RunCommand(*m.ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return &error_types.MongoConnectionError{Err: err}
	}
	loggers.InfoLog.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}
