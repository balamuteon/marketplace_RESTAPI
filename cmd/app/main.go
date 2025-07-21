// cmd/app/main.go
package main

import (
	"marketplace/internal/app"
)

// @title       Marketplace API
// @version     1.0
// @description API для учебного проекта торговой площадки.
//
// @host      marketplace-restapi.onrender.com
// @BasePath  /api/v1
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Для доступа к защищенным эндпоинтам, укажите токен в формате "Bearer ваш_токен"
func main() {
	application := app.New()

	application.Run()
}
