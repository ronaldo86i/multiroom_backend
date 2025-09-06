package main

import (
	"context"
	"multiroom/sucursal-service/internal/postgresql/routine"
	"multiroom/sucursal-service/internal/server"
	"multiroom/sucursal-service/internal/server/setup"
)

func main() {
	setup.Init()
	deps := setup.GetDependencies()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Lanza la rutina en segundo plano
	routine.Init(ctx)

	// Inicializa servidor HTTP
	httpServer := server.NewServer(deps.Handler)
	httpServer.Initialize()
}
