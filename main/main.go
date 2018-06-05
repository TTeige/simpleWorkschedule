package main

import (
	"github.com/tteige/simpleWorkschedule/service"
	"github.com/tteige/simpleWorkschedule/models"
	"github.com/gorilla/sessions"
	"os"
)

func main() {

	db, err := models.OpenDatabase("tim", "workschedule", "something")
	if err != nil {
		panic(err)
	}

	store := sessions.NewCookieStore([]byte(os.Getenv("COOKIE_STORE_SECRET")))

	server := service.Server{
		DB:          db,
		CookieStore: store,
	}
	server.Serve()
}
