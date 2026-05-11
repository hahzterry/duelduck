package main

import (
	"dd-prediction-api/app"
	"log"

	"github.com/joho/godotenv"
)

// @title						Duel Duck Prediction API
// @version					1.0
// @description				This is a swagger specification for a Duel Duck back-end.
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Enter the auth token
func main() {
	app.Build().Run()
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("no .env file found")
	}
}
