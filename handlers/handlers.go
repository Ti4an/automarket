package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"automarket/middleware"
	"automarket/models"
	"automarket/services"
	"automarket/storage"
)

type Handlers struct {
	auth     *services.AuthService
	vehicles *services.VehicleService
}

func New(auth *services.AuthService, vehicles *services.VehicleService) *Handlers {
	return &Handlers{auth: auth, vehicles: vehicles}
}

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, services.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	case errors.Is(err, services.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, storage.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, services.ErrEmailTaken), errors.Is(err, services.ErrInvalidState):
		status = http.StatusConflict
	}
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "error interno del servidor"
	}
	c.JSON(status, gin.H{"error": message})
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// publishRequest llega como multipart/form-data (campos + archivo "image").
type publishRequest struct {
	Brand string `form:"brand"`
	Model string `form:"model"`
	Year  int    `form:"year"`
	Price int    `form:"price"`
}

// POST /register
func (h *Handlers) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, services.ErrInvalidInput)
		return
	}
	user, err := h.auth.CreateUser(req.Name, req.Email, req.Phone, req.Password, models.RoleSeller)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role})
}

// POST /login
func (h *Handlers) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, services.ErrInvalidInput)
		return
	}
	token, user, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "role": user.Role, "name": user.Name})
}

// GET /vehicles
func (h *Handlers) Catalog(c *gin.Context) {
	c.JSON(http.StatusOK, h.vehicles.Catalog())
}

// GET /vehicles/:id
func (h *Handlers) GetVehicle(c *gin.Context) {
	result, err := h.vehicles.GetPublic(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /vehicles (vendedor, multipart/form-data)
func (h *Handlers) Publish(c *gin.Context) {
	var req publishRequest
	if err := c.ShouldBind(&req); err != nil {
		respondError(c, services.ErrInvalidInput)
		return
	}
	image, err := c.FormFile("image")
	if err != nil {
		respondError(c, services.ErrInvalidInput)
		return
	}
	seller := middleware.CurrentUser(c)
	vehicle, err := h.vehicles.Publish(seller.ID, req.Brand, req.Model, req.Year, req.Price, image)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}

// GET /my/vehicles (vendedor)
func (h *Handlers) MyVehicles(c *gin.Context) {
	c.JSON(http.StatusOK, h.vehicles.ListBySeller(middleware.CurrentUser(c).ID))
}

// GET /admin/vehicles (administrador)
func (h *Handlers) AllVehicles(c *gin.Context) {
	c.JSON(http.StatusOK, h.vehicles.ListAll())
}

// GET /admin/users (administrador)
func (h *Handlers) Sellers(c *gin.Context) {
	c.JSON(http.StatusOK, h.vehicles.ListSellers())
}

// PATCH /vehicles/:id/sold (vendedor)
func (h *Handlers) MarkSold(c *gin.Context) {
	seller := middleware.CurrentUser(c)
	vehicle, err := h.vehicles.MarkSold(seller.ID, c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

// PATCH /vehicles/:id/approve (administrador)
func (h *Handlers) Approve(c *gin.Context) {
	vehicle, err := h.vehicles.Approve(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

// DELETE /vehicles/:id (administrador)
func (h *Handlers) DeleteVehicle(c *gin.Context) {
	if err := h.vehicles.DeleteVehicle(c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DELETE /users/:id (administrador)
func (h *Handlers) DeleteUser(c *gin.Context) {
	if err := h.vehicles.DeleteUser(c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}