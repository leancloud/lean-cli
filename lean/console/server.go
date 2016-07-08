package console

import (
	"log"
)

// Server is a struct for develoment console server
type Server struct {
	AppID     string
	AppKey    string
	MasterKey string
	Port      string
}

// Run the dev server
func (server *Server) Run() {
	addr := "localhost:" + server.Port
	log.Println("start developement console in " + addr)
}
