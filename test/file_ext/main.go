package main

import (
	"fmt"

	enry "github.com/src-d/enry/v2"
)

func main() {
	lang, safe := enry.GetLanguageByExtension("Dockerfile")
	fmt.Println(lang, safe)
}
