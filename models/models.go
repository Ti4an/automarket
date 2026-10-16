package models

type Role string

const (
	RoleSeller Role = "vendedor"
	RoleAdmin  Role = "administrador"
)

type VehicleStatus string

const (
	StatusPending   VehicleStatus = "pendiente"
	StatusPublished VehicleStatus = "publicada"
	StatusSold      VehicleStatus = "vendida"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     Role   `json:"role"`
	Password string `json:"password"`
}

type Vehicle struct {
	ID       string        `json:"id"`
	SellerID string        `json:"seller_id"`
	Brand    string        `json:"brand"`
	Model    string        `json:"model"`
	Year     int           `json:"year"`
	Price    int           `json:"price"`
	ImageURL string        `json:"image_url"`
	Status   VehicleStatus `json:"status"`
}