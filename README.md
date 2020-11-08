# Dotenv
Dotenv is a minimal Go Library for reading and writing .env configuration files. 
It uses [renameio](https://github.com/google/renameio) to perform atomic write operations making sure _applications 
never see unexpected file content (a half-written file, or a 0-byte file)_.

Dotenv reads config in the following order. Each item takes precedence over the item below it:

- env
- key-value config cache/store
- config

The config cache store is set on first read operation.

## Installation

```sh
go get -u github.com/joho/godotenv
```

## Usage

Assuming you have a .env file in the current directory with the following values
```sh
S3_BUCKET=yours3bucket
SECRET_KEY=yoursecretKey
PRIORITY_LEVEL=2
```

### Reading .env files

```
package main

import (
    "log"
    "os"
    
    "github.com/profclems/go-dotenv"
)

func main() {
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

### Getting Values From DotEnv
The following functions and methods exist to get a value depending the Type:

- Get(key string) : interface{}
- GetBool(key string) : bool
- GetFloat64(key string) : float64
- GetInt(key string) : int
- GetIntSlice(key string) : []int
- GetString(key string) : string
- GetStringMap(key string) : map[string]interface{}
- GetStringMapString(key string) : map[string]string
- GetStringSlice(key string) : []string
- GetTime(key string) : time.Time
- GetDuration(key string) : time.Duration

## Contributing
Contributions are most welcome! It could be a new feature, bug fix, refactoring or even reporting an issue.

- Fork it
- Create your feature branch (git checkout -b my-new-feature)
- Commit your changes (git commit -am 'Added some feature')
- Push to the branch (git push origin my-new-feature)
- Create new Pull Request

## License
Copyright © [Clement Sam](http://twitter.com/clems_dev)

glab is open-sourced software licensed under the [MIT](LICENSE) license.
