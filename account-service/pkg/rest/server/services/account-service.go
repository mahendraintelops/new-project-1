package services

import (
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/daos"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/models"
)

type AccountService struct {
	accountDao *daos.AccountDao
}

func NewAccountService() (*AccountService, error) {
	accountDao, err := daos.NewAccountDao()
	if err != nil {
		return nil, err
	}
	return &AccountService{
		accountDao: accountDao,
	}, nil
}

func (accountService *AccountService) CreateAccount(account *models.Account) (*models.Account, error) {
	return accountService.accountDao.CreateAccount(account)
}

func (accountService *AccountService) ListAccounts() ([]*models.Account, error) {
	return accountService.accountDao.ListAccounts()
}

func (accountService *AccountService) GetAccount(id string) (*models.Account, error) {
	return accountService.accountDao.GetAccount(id)
}

func (accountService *AccountService) UpdateAccount(id string, account *models.Account) (*models.Account, error) {
	return accountService.accountDao.UpdateAccount(id, account)
}
