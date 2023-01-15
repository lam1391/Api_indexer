package main

import (
	"net/http"

	apiServer "api_indexer/cmd/main/pkg/apiServer"
)

func main() {

	s := apiServer.CreateNewServer()
	//creation  of API middlewares
	s.MountHandlers()
	//creation of API methods and routes.
	s.ApiMethods()
	//run server in port 3033
	http.ListenAndServe(":3033", s.Router)

}
