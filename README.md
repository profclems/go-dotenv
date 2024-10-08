# Dotenv [![Go Report Card](https://goreportcard.com/badge/github.com/profclems/go-dotenv)](https://goreportcard.com/report/github.com/profclems/go-dotenv) [![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/profclems/go-dotenv)](https://pkg.go.dev/mod/github.com/profclems/go-dotenv)

Dotenv is a minimal Go Library for reading and writing .env configuration files.

Dotenv reads config in the following order. Each item takes precedence over the item below it:

- env
- key-value config cache/store
- config

The config cache store is set on first read operation.

## Installation

```sh
go get -u github.com/profclems/go-dotenv
```

## Usage

Assuming you have a .env file in the current directory with the following values
```env
S3_BUCKET=yours3bucket
SECRET_KEY=yoursecretKey
PRIORITY_LEVEL=2
```

### Reading .env files

```go
package main

import (
    "log"
    
    "github.com/profclems/go-dotenv"
)

func main() {
  // .env - It will search for the .env file in the current directory and load it. 
  // You can explicitly set config file with dotenv.SetConfigFile("path/to/file.env")
  err := dotenv.LoadConfig()
  if err != nil {
    log.Fatalf("Error loading .env file: %v", err)
  }

  s3Bucket := dotenv.GetString("S3_BUCKET")
  secretKey := dotenv.GetString("SECRET_KEY")
  priorityLevel := dotenv.GetInt("PRIORITY_LEVEL")

  // now do something with s3 or whatever
}
```

### Writing .env files

```go
import (
	"fmt"
	"log"
	
	"github.com/profclems/go-dotenv"
)

func main() {
	// SetConfigFile explicitly defines the path, name and extension of the config file.
	dotenv.SetConfigFile("config/.env")
    err := dotenv.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dotenv.Set("STRONGEST_AVENGER", "Hulk")
	dotenv.Set("PLAYER_NAME", "Anon")

	err := dotenv.Save()
	if err != nil {
		log.Fatal(err)
	}

	value := dotenv.GetString("STRONGEST_AVENGER")
	fmt.Printf("%s = %s \n", "STRONGEST_AVENGER", value)

	value = dotenv.GetString("PLAYER_NAME")
	fmt.Printf("%s = %s \n", "PLAYER_NAME", value)
}

```

All the above examples use the global DotEnv instance. You can instantiate a new Dotenv instance:

```go
cfg := dotenv.New()
cfg.SetConfigFile("path/to/.env")
err := cfg.LoadConfig()
if err != nil {
	log.Fatalf("Error loading .env file: %v", err)
}

val := cfg.GetString("SOME_ENV")
```

### Getting Values From DotEnv
The following functions and methods exist to get a value depending the Type:

- `Get(key string) : any`
- `GetString(key string) : string`
- `GetBool(key string) : bool`
- `GetFloat64(key string) : float64`
- `GetInt(key string) : int`
- `GetIntSlice(key string) : []int`
- `GetStringSlice(key string) : []string`
- `GetTime(key string) : time.Time`
- `GetDuration(key string) : time.Duration`
- `isSet(key string) : bool`
- `LookUp(key string) : (any, bool)`
- `Set(key string, value any)`

## Contributing
Contributions are most welcome! It could be a new feature, bug fix, refactoring or even reporting an issue.

- Fork it
- Create your feature branch (git checkout -b my-new-feature)
- Commit your changes (git commit -am 'Added some feature')
- Push to the branch (git push origin my-new-feature)
- Create new Pull Request

## License
Copyright © [Clement Sam](http://twitter.com/clems_dev)

This package is open-sourced software licensed under the [MIT](LICENSE) license.
