package main

import (
	"multiroom/sucursal-service/internal/server"
	"multiroom/sucursal-service/internal/server/setup"
)

func main() {
	// Inicializar contenedor de dependencias, variables de entorno y conexión a base de datos
	setup.Init()

	deps := setup.GetDependencies()

	// Inicializar el servidor HTTP
	httpServer := server.NewServer(deps.Handler)

	// Iniciar el servidor
	httpServer.Initialize()
}
