# AutoMarket

Taller "Implementación de AAA en sitio web". Aplicación web en Go con Gin.

## Requisitos
- Go 1.21 o superior

## Cómo correrlo
1. Clonar el repositorio y entrar a la carpeta:
   ```
   git clone <URL-del-repo>
   cd automarket
   ```
2. Descargar las dependencias:
   ```
   go mod download
   ```
3. Definir las credenciales del administrador (PowerShell):
   ```
   $env:ADMIN_EMAIL="admin@automarket.com"
   $env:ADMIN_PASSWORD="Admin12345"
   ```
   En Linux/Mac: `export ADMIN_EMAIL=... ADMIN_PASSWORD=...`
4. Ejecutar:
   ```
   go run .
   ```
5. Abrir http://localhost:8080/

## Uso
- Visitante: ve el catálogo y el contacto del vendedor.
- Vendedor: se registra, inicia sesión, publica un vehículo con imagen (queda pendiente) y lo marca como vendido.
- Administrador: aprueba o elimina publicaciones y elimina usuarios.

## Archivos generados
- `data.json`: datos de usuarios y vehículos.
- `audit.log`: registro de auditoría (accounting).
- `uploads/`: imágenes subidas.