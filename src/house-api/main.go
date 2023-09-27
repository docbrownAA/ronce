package main

import (
	"os"
	"ronce/src/go/app"
	"ronce/src/go/log"
	"ronce/src/go/sql"

	"github.com/synthesio/zconfig"
)

type Service struct {
	DB     *sql.DB     `key:"postgres"`
	Logger *log.Logger `key:"logger" inject-as:"logger"`
}

func main() {
	s := new(Service)

	if err := zconfig.Configure(s); err != nil {
		log.New().Error("starting dependencies", "error", err)
		os.Exit(1)
	}
	defer app.Cleanup(s)
}
