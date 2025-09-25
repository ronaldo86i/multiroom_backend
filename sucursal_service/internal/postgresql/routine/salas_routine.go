package routine

import (
	"context"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"runtime"
	"time"
	"weak"

	amqp "github.com/rabbitmq/amqp091-go"
)

func UsoSalasActualizar(ctx context.Context, salaService port.SalaService, rabbitMQService port.RabbitMQService) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Ejecuta inmediatamente al iniciar
	runOnce(ctx, salaService, rabbitMQService)
	for {
		select {
		case <-ctx.Done():
			log.Println("UsoSalasActualizar detenido por cancelación del contexto")
			return
		case <-ticker.C:
			runOnce(ctx, salaService, rabbitMQService)
		}
	}
}

func runOnce(ctx context.Context, salaService port.SalaService, rabbitMQService port.RabbitMQService) {
	salasIds, err := salaService.ActualizarUsoSalas(ctx)
	if err != nil {
		log.Println("Error al actualizar uso de salas:", err)
		return
	}
	weakSalasIds := weak.Make(salasIds)
	ids := weakSalasIds.Value()
	if ids == nil || len(*ids) == 0 {
		return
	}

	log.Printf("Salas finalizadas: %v\n", *salasIds)

	salas, err := salaService.ObtenerListaSalasDetailByIds(ctx, *salasIds)
	if err != nil {
		log.Println("Error al obtener detalle de salas:", err)
		return
	}

	weakSalas := weak.Make(salas)
	s := weakSalas.Value()
	if s == nil || len(*s) == 0 {
		return
	}
	for _, sala := range *s {
		// Publica en canal individual
		if err := rabbitMQService.Publish(fmt.Sprintf("salas_%d", sala.Id), sala, amqp.Table{
			amqp.QueueMaxLenArg:   int32(1),
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Printf("Error al publicar sala_%d: %s", sala.Id, err.Error())
		}

		// Publica en canal general
		if err := rabbitMQService.Publish("salas", sala, amqp.Table{
			amqp.QueueMaxLenArg:   int32(1),
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Printf("Error al publicar en 'salas': %s", err.Error())
		}

		// Publica en el canal dispositivo_usuario_%s
		channel := fmt.Sprintf("dispositivo_%d_usuario_%d", sala.Dispositivo.Id, sala.Dispositivo.Usuario.Id)
		if err := rabbitMQService.Publish(channel, sala, amqp.Table{
			amqp.QueueMaxLenArg:   int32(1),
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {

			log.Printf("Error al publicar en %s: %s", channel, err.Error())
		}
	}
	runtime.GC()
}
