package main

import (
	"github.com/subutai-io/cdn/server/app"
)

func init() {
	app.InitFilters()
}

// main starts CDN server
func main() {
	app.ListenAndServe()
}
