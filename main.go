package main

import (
	"github.com/wa-candra/webservice-go/appmode"
)

func main() {
	app := App{}

	app.Init(appmode.Development)

	app.Run("localhost:3000")

}
