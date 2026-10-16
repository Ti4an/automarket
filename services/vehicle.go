package services

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"automarket/models"
	"automarket/storage"
)

var (
	ErrForbidden    = errors.New("no tiene permiso sobre este recurso")
	ErrInvalidState = errors.New("la publicación no está en un estado válido para esta operación")
)

type SellerContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type PublicVehicle struct {
	Vehicle models.Vehicle `json:"vehicle"`
	Seller  SellerContact  `json:"seller"`
}

// Límite de tamaño y extensiones permitidas para las imágenes.
const maxImageSize = 5 << 20 // 5 MB

var allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

// UserSummary es un usuario sin la contraseña (para el panel del admin).
type UserSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type VehicleService struct {
	store      *storage.Store
	uploadsDir string
}

func NewVehicleService(store *storage.Store, uploadsDir string) *VehicleService {
	return &VehicleService{store: store, uploadsDir: uploadsDir}
}

// HU-VIS-01: catálogo público
func (v *VehicleService) Catalog() []models.Vehicle {
	return v.store.ListVehiclesByStatus(models.StatusPublished)
}

// HU-VIS-01: un vehículo con el contacto del vendedor
func (v *VehicleService) GetPublic(id string) (PublicVehicle, error) {
	vehicle, err := v.store.FindVehicleByID(id)
	if err != nil {
		return PublicVehicle{}, err
	}
	if vehicle.Status != models.StatusPublished {
		return PublicVehicle{}, storage.ErrNotFound
	}
	seller, err := v.store.FindUserByID(vehicle.SellerID)
	if err != nil {
		return PublicVehicle{}, err
	}
	contact := SellerContact{Name: seller.Name, Email: seller.Email, Phone: seller.Phone}
	return PublicVehicle{Vehicle: vehicle, Seller: contact}, nil
}

// HU-VEN-01: publicar con imagen (queda pendiente de aprobación)
func (v *VehicleService) Publish(sellerID, brand, model string, year, price int, image *multipart.FileHeader) (models.Vehicle, error) {
	if strings.TrimSpace(brand) == "" || strings.TrimSpace(model) == "" || year <= 0 || price <= 0 {
		return models.Vehicle{}, ErrInvalidInput
	}
	imageURL, err := v.saveImage(image)
	if err != nil {
		return models.Vehicle{}, err
	}
	id, err := randomHex(8)
	if err != nil {
		return models.Vehicle{}, err
	}
	vehicle := models.Vehicle{
		ID: id, SellerID: sellerID, Brand: brand, Model: model,
		Year: year, Price: price, ImageURL: imageURL, Status: models.StatusPending,
	}
	if err := v.store.AddVehicle(vehicle); err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

// HU-VEN-01: reportar la venta
func (v *VehicleService) MarkSold(sellerID, vehicleID string) (models.Vehicle, error) {
	return v.changeStatus(vehicleID, models.StatusPublished, models.StatusSold, func(vehicle models.Vehicle) error {
		if vehicle.SellerID != sellerID {
			return ErrForbidden
		}
		return nil
	})
}

// HU-ADM-01: autorizar una publicación
func (v *VehicleService) Approve(vehicleID string) (models.Vehicle, error) {
	return v.changeStatus(vehicleID, models.StatusPending, models.StatusPublished, nil)
}

func (v *VehicleService) changeStatus(id string, from, to models.VehicleStatus, check func(models.Vehicle) error) (models.Vehicle, error) {
	vehicle, err := v.store.FindVehicleByID(id)
	if err != nil {
		return models.Vehicle{}, err
	}
	if check != nil {
		if err := check(vehicle); err != nil {
			return models.Vehicle{}, err
		}
	}
	if vehicle.Status != from {
		return models.Vehicle{}, ErrInvalidState
	}
	vehicle.Status = to
	if err := v.store.UpdateVehicle(vehicle); err != nil {
		return models.Vehicle{}, err
	}
	return vehicle, nil
}

// HU-ADM-01: eliminar una publicación
func (v *VehicleService) DeleteVehicle(id string) error {
	return v.store.DeleteVehicle(id)
}

// HU-ADM-01: eliminar un usuario (y sus publicaciones)
func (v *VehicleService) DeleteUser(id string) error {
	if _, err := v.store.FindUserByID(id); err != nil {
		return err
	}
	if err := v.store.DeleteVehiclesBySeller(id); err != nil {
		return err
	}
	return v.store.DeleteUser(id)
}

// saveImage valida la imagen y la guarda con un nombre aleatorio en uploadsDir.
func (v *VehicleService) saveImage(image *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(image.Filename))
	if !allowedImageExt[ext] || image.Size <= 0 || image.Size > maxImageSize {
		return "", ErrInvalidInput
	}
	name, err := randomHex(16)
	if err != nil {
		return "", err
	}
	src, err := image.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(filepath.Join(v.uploadsDir, name+ext))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return "/uploads/" + name + ext, nil
}

// HU-VEN-01: el vendedor ve sus propias publicaciones
func (v *VehicleService) ListBySeller(sellerID string) []models.Vehicle {
	return v.store.ListVehicles(func(vehicle models.Vehicle) bool { return vehicle.SellerID == sellerID })
}

// HU-ADM-01: el admin ve todas las publicaciones
func (v *VehicleService) ListAll() []models.Vehicle {
	return v.store.ListVehicles(func(models.Vehicle) bool { return true })
}

// HU-ADM-01: el admin ve los vendedores (sin contraseña)
func (v *VehicleService) ListSellers() []UserSummary {
	result := []UserSummary{}
	for _, u := range v.store.ListUsersByRole(models.RoleSeller) {
		result = append(result, UserSummary{ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone})
	}
	return result
}