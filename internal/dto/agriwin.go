package dto

import "time"

// UserToIntegrate representa a resposta da sua API Agriwin,
// informando qual produtor precisa ser integrado.
type UserToIntegrate struct {
	ProdutorID int       `json:"produtor_id"`
	Login      string    `json:"login"`
	Senha      string    `json:"senha"`
	Data       time.Time `json:"data"`
}

// IntegrationPayload é o DTO que agrupa todos os dados
// a serem enviados de volta para a aplicação Agriwin.
type IntegrationPayload struct {
	ProdutorID   int           `json:"produtor_id"`
	FetchedAt    time.Time     `json:"fetchedAt"`
	Supplies     []Supply      `json:"supplies"`
	ProductSales []ProductSale `json:"productSales"`
	Products     []Product     `json:"products"`
	FuelTypes    []FuelType    `json:"fuelTypes"`
	Vehicles     []Vehicle     `json:"vehicles"`
	Drivers      []Driver      `json:"drivers"`
	Employees    []Employee    `json:"employees"`
}

// IsEmpty verifica se o payload contém algum dado transacional para ser enviado.
func (p *IntegrationPayload) IsEmpty() bool {
	return len(p.Supplies) == 0 && len(p.ProductSales) == 0
}
