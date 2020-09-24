module github.com/pogify/pogify-api

go 1.15

replace github.com/pogify/pogify-api/v1 => ./v1

require (
	github.com/pogify/pogify-api/v1
	github.com/alicebob/miniredis/v2 v2.13.3 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.1.1 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/lestrrat/go-jwx v0.0.0-20180221005942-b7d4802280ae // indirect
	github.com/lestrrat/go-pdebug v0.0.0-20180220043741-569c97477ae8 // indirect
	github.com/matoous/go-nanoid v1.4.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
)
