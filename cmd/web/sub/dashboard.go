package sub

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bejaneps/go-git-webapp/internal/crud"
	"github.com/bejaneps/go-git-webapp/internal/util"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	c = &cache{}
)

// temp cache to save data between requests
type cache struct {
	*crud.GitCollection
}

func (e *env) handleDownload(w http.ResponseWriter, r *http.Request) {
	// filter just config files for json
	configColl := &crud.GitCollection{
		BaseURL:  c.BaseURL,
		BaseHash: c.BaseHash,
		BaseDir:  c.BaseDir,
	}
	for _, v := range c.Coll {
		if v.Config {
			configColl.Coll = append(configColl.Coll, v)
		}
	}

	// create a json file for serving
	f, err := os.Create(util.RandomString(10) + ".json")
	if err != nil {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", err)
		if err != nil {
			log.Error()
		}
		return
	}

	// create a new json file with config files
	err = json.NewEncoder(f).Encode(configColl)
	if err != nil {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", err)
		if err != nil {
			log.Error()
		}
		return
	}

	// create a download link for json file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name()))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, f.Name(), time.Now(), f)
}

// TODO: create links for each file and commit hash
func (e *env) handleData(w http.ResponseWriter, r *http.Request) {
	// parse values from url query
	url := r.URL.Query().Get("url")
	hash := r.URL.Query().Get("commit")
	dir := r.URL.Query().Get("dir")

	// get hash and file collections
	coll, err := crud.GetGitCollection(url, hash, dir)
	if err != nil {
		err := e.templateFiles.ExecuteTemplate(w, "error.html", err)
		if err != nil {
			log.Error(err)
		}
		return
	}

	// save collection to cache
	c.GitCollection = coll

	// execute collection on template
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", coll)
}

func (e *env) handleDashboard(w http.ResponseWriter, r *http.Request) {
	e.templateFiles.ExecuteTemplate(w, "dashboard.html", nil)
}
