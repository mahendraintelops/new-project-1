package daos

import (
	"context"
	"errors"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/daos/clients/nosqls"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccountDao struct {
	mongoDBClient *nosqls.MongoDBClient
	collection    *mongo.Collection
}

func NewAccountDao() (*AccountDao, error) {
	mongoDBClient, err := nosqls.InitMongoDB()
	if err != nil {
		log.Debugf("init mongoDB failed: %v", err)
		return nil, err
	}
	return &AccountDao{
		mongoDBClient: mongoDBClient,
		collection:    mongoDBClient.Database.Collection("accounts"),
	}, nil
}

func (accountDao *AccountDao) CreateAccount(account *models.Account) (*models.Account, error) {
	// create a document for given account
	insertOneResult, err := accountDao.collection.InsertOne(context.TODO(), account)
	if err != nil {
		log.Debugf("insert failed: %v", err)
		return nil, err
	}
	account.ID = insertOneResult.InsertedID.(primitive.ObjectID).Hex()

	log.Debugf("account created")
	return account, nil
}

func (accountDao *AccountDao) ListAccounts() ([]*models.Account, error) {
	filters := bson.D{}
	accounts, err := accountDao.collection.Find(context.TODO(), filters)
	if err != nil {
		return nil, err
	}
	var accountList []*models.Account
	for accounts.Next(context.TODO()) {
		var account *models.Account
		if err = accounts.Decode(&account); err != nil {
			log.Debugf("decode account failed: %v", err)
			return nil, err
		}
		accountList = append(accountList, account)
	}
	if accountList == nil {
		return []*models.Account{}, nil
	}

	log.Debugf("account listed")
	return accountList, nil
}

func (accountDao *AccountDao) GetAccount(id string) (*models.Account, error) {
	var account *models.Account
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &models.Account{}, nosqls.ErrInvalidObjectID
	}
	filter := bson.D{
		{Key: "_id", Value: objectID},
	}
	if err = accountDao.collection.FindOne(context.TODO(), filter).Decode(&account); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// This error means your query did not match any documents.
			return &models.Account{}, nosqls.ErrNotExists
		}
		log.Debugf("decode account failed: %v", err)
		return nil, err
	}

	log.Debugf("account retrieved")
	return account, nil
}

func (accountDao *AccountDao) UpdateAccount(id string, account *models.Account) (*models.Account, error) {
	if id != account.ID {
		log.Debugf("id(%s) and payload(%s) don't match", id, account.ID)
		return nil, errors.New("id and payload don't match")
	}

	existingAccount, err := accountDao.GetAccount(id)
	if err != nil {
		log.Debugf("get account failed: %v", err)
		return nil, err
	}

	// no account retrieved means no account exists to update
	if existingAccount == nil || existingAccount.ID == "" {
		return nil, nosqls.ErrNotExists
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nosqls.ErrInvalidObjectID
	}
	filter := bson.D{
		{Key: "_id", Value: objectID},
	}
	account.ID = ""
	update := bson.M{
		"$set": account,
	}

	updateResult, err := accountDao.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Debugf("update account failed: %v", err)
		return nil, err
	}
	if updateResult.ModifiedCount == 0 {
		return nil, nosqls.ErrUpdateFailed
	}

	log.Debugf("account updated")
	return account, nil
}
