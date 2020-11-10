package main

import (
	"github.com/DaniilOr/goPages/pkg/server"
	"github.com/DaniilOr/gorest/pkg/middleware/logger"
	"github.com/DaniilOr/gorest/pkg/middleware/recoverer"
	"github.com/DaniilOr/gorest/pkg/remux"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
)
const defaultPort = "8888"
const defaultHost = "0.0.0.0"

func main() {
	os.Setenv("PORT", defaultPort)
	os.Setenv("HOST", defaultHost)
	port, _ := os.LookupEnv("PORT")
	host, _ := os.LookupEnv("HOST")
	log.Println(host)
	log.Println(port)
	if err := execute(net.JoinHostPort(host, port)); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
func execute(addr string)(err error) {
	service := server.NewService()
	myLogger := logger.Logger
	myRecoverer := recoverer.Recoverer
	if err := service.Mux.NewPlain(remux.GET,  "/api/pages", http.HandlerFunc(service.GetAll), myLogger, myRecoverer); err != nil {
		return err
	}
	getRegex := regexp.MustCompile(`^/api/pages/(?P<Id>\d+)$`)
	if err := service.Mux.NewRegex(remux.GET, http.HandlerFunc(service.GetSingle), getRegex, myLogger, myRecoverer); err != nil {
		return err
	}
	if err := service.Mux.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r*http.Request){
	w.WriteHeader(http.StatusNotFound)
	})); err != nil{
		return err
	}

	if err := service.Mux.NewRegex(remux.PUT, http.HandlerFunc(service.Change), getRegex, myLogger, myRecoverer); err != nil {
		return err
	}
	if err := service.Mux.NewPlain(remux.POST, "/api/pages",  http.HandlerFunc(service.Add),myLogger,myRecoverer); err != nil {
		return err
	}
	if err := service.Mux.NewRegex(remux.DELETE, http.HandlerFunc(service.Delete), getRegex, myLogger, myRecoverer); err != nil {
		return err
	}

	server := &http.Server{
		Addr: addr,
		Handler: service,
	}
	return server.ListenAndServe()
}