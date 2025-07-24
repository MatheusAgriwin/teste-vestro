package ports

import (
	"context"
	"time"
	"vestro/internal/dto"
)

// UserProvider define o contrato para buscar os usuários que serão processados.
type UserProvider interface {
	GetUsersToIntegrate(ctx context.Context) ([]dto.UserToIntegrate, error)
}

// VestroAPIClient define o contrato para o cliente da API Vestro.
type VestroAPIClient interface {
	Authenticate(ctx context.Context) (string, error)
	GetSupplies(ctx context.Context, token string, since time.Time, userIdentifier string) ([]dto.Supply, error)
	GetProductSales(ctx context.Context, token string, since time.Time, userIdentifier string) ([]dto.ProductSale, error)
	GetProducts(ctx context.Context, token string) ([]dto.Product, error)
	GetFuelTypes(ctx context.Context, token string) ([]dto.FuelType, error)
	GetVehicles(ctx context.Context, token string) ([]dto.Vehicle, error)
	GetDrivers(ctx context.Context, token string) ([]dto.Driver, error)
	GetEmployees(ctx context.Context, token string) ([]dto.Employee, error)
}

// Notifier define o contrato para o serviço que notifica a aplicação final.
type Notifier interface {
	Send(ctx context.Context, payload dto.IntegrationPayload) error
}
