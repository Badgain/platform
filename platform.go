package platform

import (
	"fmt"

	"github.com/Badgain/platform/config"
	"github.com/Badgain/platform/database/models"
	"github.com/Badgain/platform/database/pool"
	"github.com/Badgain/platform/logger"
	"github.com/Badgain/platform/server"
	"github.com/Badgain/platform/server/controlls"
	"github.com/Badgain/platform/services"

	"github.com/gorilla/mux"
)

type Platform struct {
	Logger                   *logger.Logger
	Server                   *server.Server
	ConnectionPool           *pool.DbConnectionsPool
	Services                 *services.ServicesList
	Config                   *config.ServiceConfigFile
	loggerIsSettuped         bool
	serverIsSettuped         bool
	connectionPoolIsSettuped bool
}

func NewPlatform(path string, conn models.DbConnection) (*Platform, error) {
	platform := &Platform{
		Logger:                   &logger.Logger{},
		serverIsSettuped:         false,
		loggerIsSettuped:         false,
		connectionPoolIsSettuped: false,
	}

	platform.ConnectionPool = &pool.DbConnectionsPool{}

	err := config.ReadConfig()
	if err != nil {
		return nil, err
	}

	loggerInitWaiter := make(chan bool)
	DbConnectionPoolWaiter := make(chan error)

	go logger.InitLogger(loggerInitWaiter, path+"/logs", platform.Logger, false)
	go pool.InitConnectionPool(conn, platform.ConnectionPool, DbConnectionPoolWaiter)
	<-loggerInitWaiter
	fmt.Println("Init Logger complete")
	platform.loggerIsSettuped = true
	DbPoolError := <-DbConnectionPoolWaiter
	if DbPoolError != nil {
		close(DbConnectionPoolWaiter)
		panic(err.Error())
	}

	fmt.Println("Init DCP complete")
	platform.connectionPoolIsSettuped = true
	close(DbConnectionPoolWaiter)
	close(loggerInitWaiter)

	fmt.Println("Loading linked services...")
	dbc := platform.ConnectionPool.GetFree()
	platform.Services, err = services.LoadServices(dbc)
	if err != nil {
		panic(err.Error())
	}
	platform.ConnectionPool.Free(dbc)
	fmt.Println("Linked services is loaded")
	return platform, nil
}

func (p *Platform) SetupServer(ctrls []controlls.RouteController, middlewares ...mux.MiddlewareFunc) error {
	srv, err := server.SetupServer(ctrls, middlewares...)
	if err != nil {
		return err
	}
	p.Server = srv
	p.serverIsSettuped = true
	return nil
}

func (p *Platform) Run() error {
	var err error
	if err = p.Server.StartServer(); err != nil {
		p.Logger.Terminate()
		return err
	}

	p.Logger.DeleteEmptyLogs()
	p.Logger.Terminate()
	p.ConnectionPool.Shutdown()
	return nil
}
