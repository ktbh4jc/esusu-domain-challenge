package user_db_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	user_db "maas/user-db"
	user_model "maas/user-model"

	"github.com/stretchr/testify/assert"
	"github.com/strikesecurity/strikememongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

const (
	usersCollectionName = "maas_users"
)

var (
	usersCollection *mongo.Collection
	usersClient     *mongo.Client
	ctx             context.Context
	repository      *user_db.MongoDBUserRepository

	databaseName = "maas"
	mongoURI     = ""
	database     *mongo.Database
)

func TestMain(m *testing.M) {
	mongoServer, err := strikememongo.Start("4.2.1")
	if err != nil {
		log.Fatal(err)
	}

	mongoURI = fmt.Sprintf("%s/maas", mongoServer.URI())
	splitDatabaseName := strings.Split(mongoURI, "/")

	databaseName = splitDatabaseName[len(splitDatabaseName)-1]

	defer mongoServer.Stop()

	setup()
	m.Run()
}

func setup() {
	startApplication()
	createCollections()
	cleanup()
}

func startApplication() {
	var err error
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	usersClient, err = initDB(ctx)
	if err != nil {
		log.Fatal("error connecting to database", err)
	}
	repository = user_db.NewMongoDBUserRepository(usersClient, &ctx)

	err = usersClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("error connecting to database", err)
	}

	database = usersClient.Database(databaseName)
}

func initDB(ctx context.Context) (client *mongo.Client, err error) {
	uri := fmt.Sprintf("%s%s", mongoURI, "?retryWrites=false")

	opts := options.Client().ApplyURI(uri)
	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return
	}
	return
}

func createCollections() {
	err := database.CreateCollection(context.Background(), usersCollectionName)
	if err != nil {
		panic(fmt.Sprintf("error creating collection: %s", err.Error()))
	}
	usersCollection = database.Collection(usersCollectionName)
}

func cleanup() {
	usersCollection.DeleteMany(ctx, bson.M{})
}

func loadDefaultData() {
	usersCollection.InsertMany(ctx, user_model.DefaultUsers)
}

/*
	A quick note on testing: Working out an in-memory test database proved to be difficult with the
	timeline given. In order to submit within a reasonable timeframe, I have decided to only do a
	sample test.
*/
func TestAllUsersDebug_ReturnsAllUsers(t *testing.T) {
	loadDefaultData()
	usersActual, err := repository.AllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(usersActual))
}
