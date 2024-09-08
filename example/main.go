package main

import (
	"fmt"
	"log"

	"github.com/profclems/go-dotenv"
)

func main() {
	err := dotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bucket := dotenv.GetString("S3_BUCKET")
	secret := dotenv.GetString("SECRET_KEY")
	pLevel := dotenv.GetInt("PRIORITY_LEVEL")

	fmt.Println("S3_BUCKET:", bucket)
	fmt.Println("SECRET_KEY:", secret)
	fmt.Println("PRIORITY_LEVEL:", pLevel)
}
