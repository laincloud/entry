// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	lainlet "github.com/laincloud/lainlet/client"
	"github.com/mijia/sweb/log"
	graceful "github.com/tylerb/graceful"

	"github.com/laincloud/entry/server/config"
	"github.com/laincloud/entry/server/gen/restapi/operations"
	"github.com/laincloud/entry/server/gen/restapi/operations/commands"
	"github.com/laincloud/entry/server/gen/restapi/operations/container"
	"github.com/laincloud/entry/server/gen/restapi/operations/ping"
	"github.com/laincloud/entry/server/gen/restapi/operations/sessions"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/handler"
	"github.com/laincloud/entry/server/models"
)

//go:generate swagger generate server --target ../server/gen --name  --spec ../swagger.yml

var (
	customOptions = struct {
		ConfigFile string `long:"config" description:"the configuration file"`
	}{}
)

func configureFlags(api *operations.EntryAPI) {
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			ShortDescription: "Custum Options",
			Options:          &customOptions,
		},
	}
}

func configureAPI(api *operations.EntryAPI) http.Handler {
	if customOptions.ConfigFile == "" {
		log.Fatal("config is required.")
	}

	c, err := config.NewConfig(customOptions.ConfigFile)
	if err != nil {
		log.Fatalf("config.NewConfig() failed, error: %s.", err)
	}

	db, err := gorm.Open("mysql", c.MySQL.DataSourceName())
	if err != nil {
		log.Fatalf("gorm.Open() failed, error: %s.", err)
	}

	swarmPort := os.Getenv("SWARM_PORT")

	var dockerClient *docker.Client
	for {
		if dockerClient, err = docker.NewClient(net.JoinHostPort("swarm.lain", swarmPort)); err != nil {
			log.Errorf("Initialize docker client error: %s", err.Error())
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	g := &global.Global{
		DB:           db,
		DockerClient: dockerClient,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		LAINDomain:    os.Getenv("LAIN_DOMAIN"),
		LAINLETClient: lainlet.New(net.JoinHostPort("lainlet.lain", os.Getenv("LAINLET_PORT"))),
		SSOURL:        c.SSOURL,
	}
	ctx, cancel := context.WithCancel(context.Background())

	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf
	api.Logger = log.Infof

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ContainerAttachContainerHandler = container.AttachContainerHandlerFunc(func(params container.AttachContainerParams) middleware.Responder {
		return handler.HandleWebsocket(ctx, models.SessionTypeAttach, handler.Attach, params.HTTPRequest, g)
	})
	api.ContainerEnterContainerHandler = container.EnterContainerHandlerFunc(func(params container.EnterContainerParams) middleware.Responder {
		return handler.HandleWebsocket(ctx, models.SessionTypeEnter, handler.Enter, params.HTTPRequest, g)
	})

	api.PingPingHandler = ping.PingHandlerFunc(handler.Ping)
	api.SessionsListSessionsHandler = sessions.ListSessionsHandlerFunc(func(params sessions.ListSessionsParams) middleware.Responder {
		return handler.ListSessions(params, db)
	})
	api.SessionsGetSessionHandler = sessions.GetSessionHandlerFunc(func(params sessions.GetSessionParams) middleware.Responder {
		return handler.GetSession(params, db)
	})
	api.CommandsListCommandsHandler = commands.ListCommandsHandlerFunc(func(params commands.ListCommandsParams) middleware.Responder {
		return handler.ListCommands(params, db)
	})

	api.ServerShutdown = func() {
		cancel()
		db.Close()
	}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
