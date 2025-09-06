package routine

import (
	"context"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/port"
	"time"
)

func UsoSalasActualizar(ctx context.Context, salaService port.SalaService, rabbitMQService port.RabbitMQService) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Ejecuta inmediatamente al iniciar
	runOnce(ctx, salaService, rabbitMQService)
	for {
		select {
		case <-ctx.Done():
			log.Println("⏹ UsoSalasActualizar detenido por cancelación del contexto")
			return
		case <-ticker.C:
			runOnce(ctx, salaService, rabbitMQService)
		}
	}
}

func runOnce(ctx context.Context, salaService port.SalaService, rabbitMQService port.RabbitMQService) {
	salasIds, err := salaService.ActualizarUsoSalas(ctx)
	if err != nil {
		log.Println("❌ Error al actualizar uso de salas:", err)
		return
	}

	if salasIds == nil || len(*salasIds) == 0 {
		log.Println("ℹ️ Ninguna sala finalizada en este ciclo")
		return
	}

	log.Printf("✅ Salas finalizadas: %v\n", *salasIds)

	salas, err := salaService.ObtenerListaSalasDetailByIds(ctx, *salasIds)
	if err != nil {
		log.Println("❌ Error al obtener detalle de salas:", err)
		return
	}

	for _, sala := range *salas {
		// Publica en canal individual
		if err := rabbitMQService.Publish(fmt.Sprintf("salas_%d", sala.Id), sala); err != nil {
			log.Printf("❌ Error al publicar sala_%d: %s", sala.Id, err.Error())
		}
		// Publica en canal general
		if err := rabbitMQService.Publish("salas", sala); err != nil {
			log.Printf("❌ Error al publicar en 'salas': %s", err.Error())
		}
	}
}
