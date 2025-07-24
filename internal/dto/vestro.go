package dto

// VestroResponseWrapper é a estrutura padrão de resposta da API.
type VestroResponseWrapper struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Count   int         `json":count"`
}

// AuthResponse é a resposta da rota de autenticação.
type AuthResponse struct {
	Session string `json:"session"`
	Access  string `json:"access"` // Este é o Bearer Token
}

// Supply é a estrutura de um registro de abastecimento.
type Supply struct {
	ID                 int    `json:"id"`
	Fuel               string `json:"fuel"`
	Date               string `json:"date"` // "yyyy-mm-ddThh-mm-ssZ"
	Volume             string `json:"volume"`
	Plate              string `json:"plate"`
	Mileage            string `json:"mileage"`
	Company            string `json:"company"`
	Employee           string `json:"employee"`
	Driver             string `json:"driver"`
	EmployeeEnrollment string `json:"employeeEnrollment"`
	DriverEnrollment   string `json:"driverEnrollment"`
}

// ProductSale representa uma venda de produto consolidado.
type ProductSale struct {
	ID                 int    `json:"id"`
	SerialNumber       string `json:"serialNumber"`
	Date               string `json:"date"`
	Name               string `json:"name"`
	Amount             string `json:"amount"`
	Driver             string `json:"driver"`
	DriverEnrollment   string `json:"driverEnrollment"`
	Plate              string `json:"plate"`
	Company            string `json:"company"`
	Employee           string `json:"employee"`
	EmployeeEnrollment string `json:"employeeEnrollment"`
}

// Product representa um produto.
type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// FuelType representa um tipo de combustível.
type FuelType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Vehicle representa um veículo.
type Vehicle struct {
	ID       int    `json:"id"`
	Plate    string `json:"plate"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Company  string `json:"companyName"`
	IsActive bool   `json:"active"`
}

// Driver representa um motorista.
type Driver struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Enrollment string `json:"enrollment"`
	IsActive   bool   `json:"active"`
}

// Employee representa um funcionário.
type Employee struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Enrollment string `json:"enrollment"`
	IsActive   bool   `json:"active"`
}
