package routine

import (
	"context"
	"multiroom/sucursal-service/internal/server/setup"
)

func Init(ctx context.Context) {
	deps := setup.GetDependencies()
	go UsoSalasActualizar(ctx, deps.Service.Sala, deps.Service.RabbitMQ)
}
