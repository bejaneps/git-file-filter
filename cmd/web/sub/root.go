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

	"github.com/bejaneps/go-git-webapp/internal/crud"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var listenPort = ":" + os.Getenv("PORT")

type env struct {
	router *mux.Router

	gitCollectionFiles   *crud.GitCollection
	gitCollectionConfigs *crud.GitCollection

	templateCache map[string]*template.Template
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file pat
		// and assign it to the name variable.
		name := filepath.Base(page)

		// Parse the page template file in to a template set.
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'layout' templates to the
		// template set (in our case, it's just the 'base' layout at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'partial' templates to the
		// template set (in our case, it's just the 'footer' partial at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}

	// Return the map.
	return cache, nil
}

func newEnv() (*env, error) {
	var op = "cmd.newEnv"
	var err error

	e := &env{}

	// register routes and router
	e.router = mux.NewRouter()
	e.routes()

	// initialize a new template cache...
	e.templateCache, err = newTemplateCache("./ui/html/")
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
		listenPort = ":4000"
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
