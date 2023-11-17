package nosqls

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"

	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"sync"
)

var (
	ErrDuplicate       = errors.New("document already exists")
	ErrNotExists       = errors.New("document not exists")
	ErrUpdateFailed    = errors.New("update failed")
	ErrDeleteFailed    = errors.New("delete failed")
	ErrInvalidObjectID = errors.New("objectID is invalid")
)

var (
	isMongoAtlas = os.Getenv("IS_MONGO_ATLAS")
	user         = os.Getenv("MONGO_DB_USER")
	password     = os.Getenv("MONGO_DB_PASSWORD")
	host         = os.Getenv("MONGO_DB_HOST")
	port         = os.Getenv("MONGO_DB_PORT")
	database     = os.Getenv("MONGO_DATABASE")
)

var o sync.Once

type MongoDBClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var mongoDBClientErr error
var mongoDBClient *MongoDBClient

func InitMongoDB() (*MongoDBClient, error) {
	o.Do(func() {
		var client *mongo.Client
		var dataSourceURI string
		if isMongoAtlas == "true" {
			dataSourceURI = fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority", user, password, host, database)
		} else {
			dataSourceURI = fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)
		}

		serviceName := os.Getenv("SERVICE_NAME")
		collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		if len(serviceName) > 0 && len(collectorURL) > 0 {
			// add opentel
			client, mongoDBClientErr = mongo.Connect(context.TODO(), options.Client().ApplyURI(dataSourceURI).SetMonitor(otelmongo.NewMonitor()))
		} else {
			client, mongoDBClientErr = mongo.Connect(context.TODO(), options.Client().ApplyURI(dataSourceURI))
		}

		if mongoDBClientErr != nil {
			log.Debugf("mongoDBClientErr: %v", mongoDBClientErr)
			return
		}

		mongoDBClientErr = client.Ping(context.TODO(), readpref.Primary())
		if mongoDBClientErr != nil {
			log.Debugf("mongoDBClientErr: %v", mongoDBClientErr)
			return
		}

		mongoDBClient = &MongoDBClient{
			Client:   client,
			Database: client.Database(database),
		}
	})

	return mongoDBClient, mongoDBClientErr
}
