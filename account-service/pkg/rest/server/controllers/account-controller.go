package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/daos/clients/nosqls"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/models"
	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/services"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
)

type AccountController struct {
	accountService *services.AccountService
}

func NewAccountController() (*AccountController, error) {
	accountService, err := services.NewAccountService()
	if err != nil {
		return nil, err
	}
	return &AccountController{
		accountService: accountService,
	}, nil
}

func (accountController *AccountController) CreateAccount(context *gin.Context) {
	// validate input
	var input models.Account
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// trigger account creation
	account, err := accountController.accountService.CreateAccount(&input)
	if err != nil {
		log.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, accountCreated)
}

func (accountController *AccountController) ListAccounts(context *gin.Context) {
	// trigger all accounts fetching
	accounts, err := accountController.accountService.ListAccounts()
	if err != nil {
		log.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, accounts)
}

func (accountController *AccountController) FetchAccount(context *gin.Context) {
	// trigger account fetching
	account, err := accountController.accountService.GetAccount(context.Param("id"))
	if err != nil {
		log.Error(err)
		if errors.Is(err, nosqls.ErrNotExists) {
			context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, nosqls.ErrInvalidObjectID) {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	serviceName := os.Getenv("SERVICE_NAME")
	collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if len(serviceName) > 0 && len(collectorURL) > 0 {
		// get the current span by the request context
		currentSpan := trace.SpanFromContext(context.Request.Context())
		currentSpan.SetAttributes(attribute.String("account.id", account.ID))
	}

	context.JSON(http.StatusOK, account)
}

func (accountController *AccountController) UpdateAccount(context *gin.Context) {
	// validate input
	var input models.Account
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// trigger account update
	if _, err := accountController.accountService.UpdateAccount(context.Param("id"), &input); err != nil {
		log.Error(err)
		if errors.Is(err, nosqls.ErrNotExists) {
			context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, nosqls.ErrInvalidObjectID) {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusNoContent)
}
