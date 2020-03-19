package main

import (
	"os"

	"github.com/bejaneps/go-git-webapp/cmd/web/sub"
	log "github.com/sirupsen/logrus"
)

func main() {
	//log.SetReportCaller(true) // log method names
	log.SetOutput(os.Stdout) // log into file
	if err := sub.Execute(); err != nil {
		log.Fatal(err)
	}
}
