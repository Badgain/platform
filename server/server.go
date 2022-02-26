package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/Badgain/platform/config"
	"github.com/Badgain/platform/server/controlls"
	"github.com/gorilla/mux"
)

type Server struct {
	*http.Server
}

func SetRoutes(rt *mux.Router, ctrl controlls.RouteController, middlewares ...mux.MiddlewareFunc) {
	ctrl.Link(rt, middlewares...)
}

func SetupServer(controllers []controlls.RouteController, middlewares ...mux.MiddlewareFunc) (*Server, error) {
	r := mux.NewRouter()

	for _, mdw := range middlewares {
		r.Use(mdw)
	}

	for _, control := range controllers {
		SetRoutes(r, control, middlewares...)
	}

	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Host, config.GlobalConfig.Port)

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}
	return &Server{&server}, nil
}

func (s *Server) StartServer() error {
	fmt.Println("Server is starting...")
	var err error
	go func() {
		if err = s.StartListening(); err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()

	shut := make(chan os.Signal)
	signal.Notify(shut, os.Interrupt)
	sig := <-shut
	fmt.Println("Server shutdown. Reason: " + sig.String())
	if err = s.Shutdown(context.Background()); err != nil {
		panic(err.Error())
	}
	return nil
}

func CheckLogger(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("SUCCESS"))
}

func (s *Server) StartListening() error {

	fmt.Println(fmt.Sprintf("Server is running on %s:%d", config.GlobalConfig.Host, config.GlobalConfig.Port))
	return s.ListenAndServe()
}
