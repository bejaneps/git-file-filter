package main

import (
	"io/ioutil"
	"log"

	"github.com/ghodss/yaml"
)

func main() {
	bs, err := ioutil.ReadFile("docker-compose.yml")
	if err != nil {
		log.Fatal(err)
	} else if len(bs) == 0 {
		log.Fatal("no bytes red from Dockerfile")
	}

	js, err := yaml.YAMLToJSON(bs)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(js))
}
