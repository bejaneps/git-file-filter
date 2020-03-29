package main

import (
	"os"

	"github.com/bejaneps/go-git-webapp/cmd/web/sub"
	log "github.com/sirupsen/logrus"
)

func main() {
	//log.SetReportCaller(true) // log method names
	log.SetOutput(os.Stdout) // log into file

	// start executing functions 1 by 1
	if err := sub.Execute(); err != nil {
		log.Fatal(err)
	}
}
