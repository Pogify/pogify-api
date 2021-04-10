module github.com/pogify/pogify-api/main

go 1.15

replace github.com/pogify/pogify-api/v2 => ../

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/joho/godotenv v1.3.0
	github.com/pogify/pogify-api v1.0.12
	github.com/pogify/pogify-api/v2 v2.0.0-beta.3
)
