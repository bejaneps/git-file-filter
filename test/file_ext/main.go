package main

import (
	"fmt"

	enry "github.com/src-d/enry/v2"
)

func main() {
	lang, safe := enry.GetLanguageByExtension("sras/dasd/rqwr/r-123/main.churs")
	fmt.Println(lang, safe)
}
