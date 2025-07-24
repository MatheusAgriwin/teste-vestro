package dto

import "time"

// UserToIntegrate representa a resposta da sua API Agriwin, informando quem precisa ser integrado.
type UserToIntegrate struct {
	// UUID do usuário no seu sistema Agriwin.
	UserUUID string `json:"userUuid"`
	// Identificador do usuário no sistema Vestro (ex: matrícula do motorista).
	VestroIdentifier string `json:"vestroIdentifier"`
	// Data da última integração bem-sucedida para este usuário.
	LastIntegration time.Time `json:"lastIntegration"`
}

// IntegrationPayload é o DTO que agrupa todos os dados a serem enviados para a aplicação Grails.
type IntegrationPayload struct {
	UserUUID     string        `json:"userUuid"`
	FetchedAt    time.Time     `json:"fetchedAt"`
	Supplies     []Supply      `json:"supplies"`
	ProductSales []ProductSale `json:"productSales"`
	// Os outros campos de dados continuam aqui
	Products  []Product  `json:"products"`
	FuelTypes []FuelType `json:"fuelTypes"`
	Vehicles  []Vehicle  `json:"vehicles"`
	Drivers   []Driver   `json:"drivers"`
	Employees []Employee `json:"employees"`
}

// IsEmpty verifica se o payload contém algum dado para ser enviado.
func (p *IntegrationPayload) IsEmpty() bool {
	return len(p.Supplies) == 0 &&
		len(p.ProductSales) == 0 &&
		len(p.Products) == 0 &&
		len(p.FuelTypes) == 0 &&
		len(p.Vehicles) == 0 &&
		len(p.Drivers) == 0 &&
		len(p.Employees) == 0
}
