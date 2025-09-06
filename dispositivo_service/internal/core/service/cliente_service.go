package service

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/port"
)

type ClienteService struct {
	clienteRepository port.ClienteRepository
}

func (c ClienteService) RegistrarCliente(ctx context.Context, request *domain.ClienteRequest) (*int, error) {
	return c.clienteRepository.RegistrarCliente(ctx, request)
}

func (c ClienteService) ModificarCliente(ctx context.Context, id *int, request *domain.ClienteRequest) error {
	return c.clienteRepository.ModificarCliente(ctx, id, request)
}

func (c ClienteService) ObtenerListaClientes(ctx context.Context, filtros map[string]string) (*[]domain.ClienteInfo, error) {
	return c.clienteRepository.ObtenerListaClientes(ctx, filtros)
}

func (c ClienteService) ObtenerClienteDetailById(ctx context.Context, id *int) (*domain.ClienteDetail, error) {
	return c.clienteRepository.ObtenerClienteDetailById(ctx, id)
}

func (c ClienteService) HabilitarCliente(ctx context.Context, id *int) error {
	return c.clienteRepository.HabilitarCliente(ctx, id)
}

func (c ClienteService) DeshabilitarCliente(ctx context.Context, id *int) error {
	return c.clienteRepository.DeshabilitarCliente(ctx, id)
}

func NewClienteService(clienteRepository port.ClienteRepository) *ClienteService {
	return &ClienteService{clienteRepository: clienteRepository}
}

var _ port.ClienteService = (*ClienteService)(nil)
