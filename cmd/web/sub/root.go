package sub

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var listenPort = ":" + os.Getenv("PORT")

type env struct {
	router *mux.Router

	templateFiles *template.Template
}

func newEnv() (*env, error) {
	var op = "cmd.newEnv"
	var err error

	e := &env{}

	// register routes and router
	e.router = mux.NewRouter()
	e.routes()

	// parse all templates
	e.templateFiles, err = template.ParseGlob(filepath.Join("templates", "*"))
	if err != nil {
		return nil, errors.WithMessagef(err, "(%s): ", op)
	}

	return e, nil
}

// Execute all functions 1 by 1
func Execute() (err error) {
	var op = "cmd.Execute"

	e, err := newEnv()
	if err != nil {
		err = errors.Wrapf(err, "(%s): initializing env", op)
		return
	}

	// check if port variable is set, if no set it to default value
	if len(listenPort) < 2 {
		listenPort = ":7000"
	}

	// setup a server
	var server = &http.Server{
		Addr:         listenPort,
		Handler:      e.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	// listen and serve connections
	errChan := make(chan error)
	go func(errChan chan<- error) {
		log.Infof("listening for incoming connections on: %s PORT", listenPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.Wrapf(err, "(%s): listen", op)
			return
		}
	}(errChan)

	// deal a CTRL + C signal
	log.Info("waiting for SIGINT signal")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// catch channels
	select {
	case <-errChan:
		err = <-errChan
		break
	case <-quit:
		// shutdown gracefully
		log.Info("shutting down server gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = server.Shutdown(ctx); err != nil {
			err = errors.Wrapf(err, "(%s): server shutdown", op)
			break
		}
		<-ctx.Done()
		break
	}

	return err
}
