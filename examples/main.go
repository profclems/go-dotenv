package main

import (
	"fmt"
	"log"

	"github.com/profclems/go-dotenv"
)

func main() {
	// SetConfigFile explicitly defines the path, name and extension of the config file.
	// .env - It will search for the .env file in the current directory
	dotenv.SetConfigFile(".env")

	// Find and read the config file
	err := dotenv.LoadConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	// dotenv.Get() returns an interface{}
	// to get the underlying type of the key,
	// we have to do the type assertion, we know the underlying value is string
	// if we type assert to other type it will throw an error
	value, ok := dotenv.Get("STRONGEST_AVENGER").(string)

	// If the type is a string then ok will be true
	// ok will make sure the program not break
	if !ok {
		log.Fatalf("Invalid type assertion")
	}

	fmt.Printf("%s = %s \n", "STRONGEST_AVENGER", value)

	// Alternatively, you can use any of the Get___ methods for a specific value type
	value = dotenv.GetString("STRONGEST_AVENGER")

	fmt.Printf("%s = %s \n", "STRONGEST_AVENGER", value)
}
