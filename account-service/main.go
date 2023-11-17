package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/mahendraintelops/new-project-1/account-service/config"
	restcontrollers "github.com/mahendraintelops/new-project-1/account-service/pkg/rest/server/controllers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sinhashubham95/go-actuator"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"

	"github.com/mahendraintelops/new-project-1/account-service/pkg/rest/client"
)

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE")
)

func main() {

	// rest server configuration
	router := gin.Default()
	var restTraceProvider *sdktrace.TracerProvider
	if len(serviceName) > 0 && len(collectorURL) > 0 {
		// add opentel
		restTraceProvider = config.InitRestTracer(serviceName, collectorURL, insecure)
		router.Use(otelgin.Middleware(serviceName))
	}
	defer func() {
		if restTraceProvider != nil {
			if err := restTraceProvider.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}
	}()
	// add actuator
	addActuator(router)
	// add prometheus
	addPrometheus(router)

	accountController, err := restcontrollers.NewAccountController()
	if err != nil {
		log.Errorf("error occurred: %v", err)
		os.Exit(1)
	}

	v1 := router.Group("/v1")
	{

		v1.POST("/accounts", accountController.CreateAccount)

		v1.GET("/accounts", accountController.ListAccounts)

		v1.GET("/accounts/:id", accountController.FetchAccount)

		v1.PUT("/accounts/:id", accountController.UpdateAccount)

	}

	Port := ":4300"
	log.Println("Server started")
	if err = router.Run(Port); err != nil {
		log.Errorf("error occurred: %v", err)
		os.Exit(1)
	}

	// this will not be called as the control won't reach here.
	// call external services here if the HasRestClients value is true
	// (that means this repo is a client to external service(s)
	var err0 error

	bNodeEe, err0 := client.ExecuteNodeEe()
	if err0 != nil {
		log.Printf("error occurred: %v", err0)
		return
	}
	log.Printf("response received: %s", string(bNodeEe))

}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func addPrometheus(router *gin.Engine) {
	router.GET("/metrics", prometheusHandler())
}

func addActuator(router *gin.Engine) {
	actuatorHandler := actuator.GetActuatorHandler(&actuator.Config{Endpoints: []int{
		actuator.Env,
		actuator.Info,
		actuator.Metrics,
		actuator.Ping,
		// actuator.Shutdown,
		actuator.ThreadDump,
	},
		Env:     "dev",
		Name:    "account-service",
		Port:    4300,
		Version: "0.0.1",
	})
	ginActuatorHandler := func(ctx *gin.Context) {
		actuatorHandler(ctx.Writer, ctx.Request)
	}
	router.GET("/actuator/*endpoint", ginActuatorHandler)
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
