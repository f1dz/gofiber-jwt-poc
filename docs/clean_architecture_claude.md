# Go Fiber + GORM Clean Architecture Project## Penjelasan Clean Architecture (Updated)

## Struktur Direktori

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                    # Clean, hanya 30-40 baris!
├── internal/
│   ├── app/
│   │   └── app.go                     # Dependency container
│   ├── domain/
│   │   ├── entity/
│   │   │   └── user.go
│   │   └── repository/
│   │       └── user_repository.go
│   ├── usecase/
│   │   └── user_usecase.go
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   └── user_handler.go
│   │       ├── middleware/
│   │       │   └── auth_middleware.go
│   │       └── route/
│   │           ├── route.go           # Main route setup
│   │           ├── user_route.go      # User routes (modular!)
│   │           ├── product_route.go   # Product routes
│   │           └── order_route.go     # Order routes
│   └── repository/
│       └── postgres/
│           ├── model/
│           │   └── user_model.go
│           └── user_repository.go
├── pkg/
│   ├── database/
│   │   └── postgres.go
│   ├── response/
│   │   └── response.go
│   └── validator/
│       └── validator.go
├── config/
│   └── config.go
├── .env
├── go.mod
└── go.sum
```

## 1. go.mod

```go
module myapp

go 1.21

require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/joho/godotenv v1.5.1
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/go-playground/validator/v10 v10.16.0
)
```

## 2. .env

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=myapp_db
DB_SSLMODE=disable

APP_PORT=3000
APP_ENV=development
```

## 3. config/config.go

```go
package config

import (
    "fmt"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    Database DatabaseConfig
    App      AppConfig
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type AppConfig struct {
    Port string
    Env  string
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, fmt.Errorf("error loading .env file: %w", err)
    }

    config := &Config{
        Database: DatabaseConfig{
            Host:     os.Getenv("DB_HOST"),
            Port:     os.Getenv("DB_PORT"),
            User:     os.Getenv("DB_USER"),
            Password: os.Getenv("DB_PASSWORD"),
            DBName:   os.Getenv("DB_NAME"),
            SSLMode:  os.Getenv("DB_SSLMODE"),
        },
        App: AppConfig{
            Port: os.Getenv("APP_PORT"),
            Env:  os.Getenv("APP_ENV"),
        },
    }

    return config, nil
}
```

## 4. pkg/database/postgres.go

```go
package database

import (
    "fmt"
    "myapp/config"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })

    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    return db, nil
}
```

## 5. pkg/response/response.go

```go
package response

import "github.com/gofiber/fiber/v2"

type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func Success(c *fiber.Ctx, message string, data interface{}) error {
    return c.Status(fiber.StatusOK).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
    return c.Status(fiber.StatusCreated).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

func BadRequest(c *fiber.Ctx, message string, err error) error {
    errMsg := ""
    if err != nil {
        errMsg = err.Error()
    }
    return c.Status(fiber.StatusBadRequest).JSON(Response{
        Success: false,
        Message: message,
        Error:   errMsg,
    })
}

func InternalServerError(c *fiber.Ctx, message string, err error) error {
    errMsg := ""
    if err != nil {
        errMsg = err.Error()
    }
    return c.Status(fiber.StatusInternalServerError).JSON(Response{
        Success: false,
        Message: message,
        Error:   errMsg,
    })
}

func NotFound(c *fiber.Ctx, message string) error {
    return c.Status(fiber.StatusNotFound).JSON(Response{
        Success: false,
        Message: message,
    })
}
```

## 6. internal/domain/entity/user.go (Clean - Tanpa Dependency)

```go
package entity

import "time"

// User adalah domain entity yang clean, tanpa dependency eksternal
type User struct {
    ID        uint
    Name      string
    Email     string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

// Business logic methods bisa ditambahkan di sini
func (u *User) IsDeleted() bool {
    return u.DeletedAt != nil
}
```

## 6b. internal/repository/postgres/model/user_model.go (Database Model)

```go
package model

import (
    "myapp/internal/domain/entity"
    "time"

    "gorm.io/gorm"
)

// UserModel adalah representasi database dari User entity
// Model ini boleh menggunakan GORM tags karena ada di layer infrastructure
type UserModel struct {
    ID        uint           `gorm:"primarykey"`
    Name      string         `gorm:"size:255;not null"`
    Email     string         `gorm:"size:255;uniqueIndex;not null"`
    Password  string         `gorm:"size:255;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
    return "users"
}

// ToEntity converts database model to domain entity
func (m *UserModel) ToEntity() *entity.User {
    var deletedAt *time.Time
    if m.DeletedAt.Valid {
        deletedAt = &m.DeletedAt.Time
    }

    return &entity.User{
        ID:        m.ID,
        Name:      m.Name,
        Email:     m.Email,
        Password:  m.Password,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
        DeletedAt: deletedAt,
    }
}

// FromEntity converts domain entity to database model
func (m *UserModel) FromEntity(user *entity.User) {
    m.ID = user.ID
    m.Name = user.Name
    m.Email = user.Email
    m.Password = user.Password
    m.CreatedAt = user.CreatedAt
    m.UpdatedAt = user.UpdatedAt
    
    if user.DeletedAt != nil {
        m.DeletedAt = gorm.DeletedAt{
            Time:  *user.DeletedAt,
            Valid: true,
        }
    }
}
```

## 7. internal/domain/repository/user_repository.go

```go
package repository

import (
    "context"
    "errors"
    "myapp/internal/domain/entity"
)

// Domain errors yang bisa digunakan di seluruh aplikasi
var (
    ErrNotFound      = errors.New("record not found")
    ErrDuplicateKey  = errors.New("duplicate key")
    ErrInvalidInput  = errors.New("invalid input")
)

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    FindByID(ctx context.Context, id uint) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindAll(ctx context.Context) ([]entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uint) error
}
```

## 8. internal/repository/postgres/user_repository.go (Updated)

```go
package postgres

import (
    "context"
    "errors"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"
    "myapp/internal/repository/postgres/model"

    "gorm.io/gorm"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
    var userModel model.UserModel
    userModel.FromEntity(user)
    
    if err := r.db.WithContext(ctx).Create(&userModel).Error; err != nil {
        // Convert GORM errors ke domain errors
        return r.handleError(err)
    }
    
    // Update entity with generated ID and timestamps
    *user = *userModel.ToEntity()
    return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
    var userModel model.UserModel
    err := r.db.WithContext(ctx).First(&userModel, id).Error
    if err != nil {
        return nil, r.handleError(err)
    }
    return userModel.ToEntity(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var userModel model.UserModel
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel).Error
    if err != nil {
        return nil, r.handleError(err)
    }
    return userModel.ToEntity(), nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]entity.User, error) {
    var userModels []model.UserModel
    err := r.db.WithContext(ctx).Find(&userModels).Error
    if err != nil {
        return nil, r.handleError(err)
    }
    
    users := make([]entity.User, len(userModels))
    for i, um := range userModels {
        users[i] = *um.ToEntity()
    }
    
    return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    var userModel model.UserModel
    userModel.FromEntity(user)
    
    if err := r.db.WithContext(ctx).Save(&userModel).Error; err != nil {
        return r.handleError(err)
    }
    
    *user = *userModel.ToEntity()
    return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
    err := r.db.WithContext(ctx).Delete(&model.UserModel{}, id).Error
    return r.handleError(err)
}

// handleError converts GORM-specific errors to domain errors
// Ini adalah KUNCI: Repository bertanggung jawab untuk error translation
func (r *userRepository) handleError(err error) error {
    if err == nil {
        return nil
    }
    
    // Convert GORM errors ke domain errors
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return repository.ErrNotFound
    }
    
    // Bisa tambahkan handling untuk error lain
    // misalnya duplicate key, foreign key constraint, dll
    
    // Return original error jika tidak bisa di-convert
    return err
}
```

## 9. internal/usecase/user_usecase.go (Versi Pragmatis - Tanpa Interface)

```go
package usecase

import (
    "context"
    "errors"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"
)

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidInput      = errors.New("invalid input")
)

// Tidak perlu interface jika hanya ada 1 implementasi
// Langsung pakai struct saja
type UserUsecase struct {
    userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) *UserUsecase {
    return &UserUsecase{
        userRepo: userRepo,
    }
}

func (u *UserUsecase) Create(ctx context.Context, user *entity.User) error {
    // Business logic validation
    if user.Name == "" || user.Email == "" {
        return ErrInvalidInput
    }

    // Check if email already exists
    _, err := u.userRepo.FindByEmail(ctx, user.Email)
    if err == nil {
        return ErrEmailAlreadyExists
    }
    if err != repository.ErrNotFound {
        return err
    }

    return u.userRepo.Create(ctx, user)
}

func (u *UserUsecase) GetByID(ctx context.Context, id uint) (*entity.User, error) {
    user, err := u.userRepo.FindByID(ctx, id)
    if err != nil {
        if err == repository.ErrNotFound {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return user, nil
}

func (u *UserUsecase) GetAll(ctx context.Context) ([]entity.User, error) {
    return u.userRepo.FindAll(ctx)
}

func (u *UserUsecase) Update(ctx context.Context, user *entity.User) error {
    if user.Name == "" || user.Email == "" {
        return ErrInvalidInput
    }

    _, err := u.userRepo.FindByID(ctx, user.ID)
    if err != nil {
        if err == repository.ErrNotFound {
            return ErrUserNotFound
        }
        return err
    }

    return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) Delete(ctx context.Context, id uint) error {
    _, err := u.userRepo.FindByID(ctx, id)
    if err != nil {
        if err == repository.ErrNotFound {
            return ErrUserNotFound
        }
        return err
    }

    return u.userRepo.Delete(ctx, id)
}
```

**Catatan:**
- Tidak ada interface `UserUsecase`, langsung pakai struct `*UserUsecase`
- Tetap bisa di-test dengan mock repository
- Lebih sederhana dan pragmatis
- Cocok untuk mayoritas use case

## 10. internal/delivery/http/handler/user_handler.go

```go
package handler

import (
    "myapp/internal/domain/entity"
    "myapp/internal/usecase"
    "myapp/pkg/response"
    "strconv"

    "github.com/gofiber/fiber/v2"
)

type UserHandler struct {
    userUsecase *usecase.UserUsecase // ← Pakai concrete struct
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
    return &UserHandler{
        userUsecase: userUsecase,
    }
}

type CreateUserRequest struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(c, "Invalid request body", err)
    }

    user := &entity.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password, // In production, hash this password!
    }

    // Pakai concrete struct, bukan interface
    if err := h.userUsecase.Create(c.Context(), user); err != nil {
        return response.BadRequest(c, "Failed to create user", err)
    }

    return response.Created(c, "User created successfully", user)
}

func (h *UserHandler) GetByID(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return response.BadRequest(c, "Invalid user ID", err)
    }

    user, err := h.userUsecase.GetByID(c.Context(), uint(id))
    if err != nil {
        return response.NotFound(c, err.Error())
    }

    return response.Success(c, "User retrieved successfully", user)
}

func (h *UserHandler) GetAll(c *fiber.Ctx) error {
    users, err := h.userUsecase.GetAll(c.Context())
    if err != nil {
        return response.InternalServerError(c, "Failed to retrieve users", err)
    }

    return response.Success(c, "Users retrieved successfully", users)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return response.BadRequest(c, "Invalid user ID", err)
    }

    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(c, "Invalid request body", err)
    }

    user := &entity.User{
        ID:       uint(id),
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
    }

    if err := h.userUsecase.Update(c.Context(), user); err != nil {
        return response.BadRequest(c, "Failed to update user", err)
    }

    return response.Success(c, "User updated successfully", user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return response.BadRequest(c, "Invalid user ID", err)
    }

    if err := h.userUsecase.Delete(c.Context(), uint(id)); err != nil {
        return response.BadRequest(c, "Failed to delete user", err)
    }

    return response.Success(c, "User deleted successfully", nil)
}
```

## 12c. internal/delivery/http/route/route.go (Main Router)

```go
package route

import (
    "myapp/internal/app"

    "github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up all application routes
func SetupRoutes(fiberApp *fiber.App, application *app.Application) {
    // API v1 group
    api := fiberApp.Group("/api/v1")

    // Setup domain-specific routes
    SetupUserRoutes(api, application.UserHandler)
    // SetupProductRoutes(api, application.ProductHandler)
    // SetupOrderRoutes(api, application.OrderHandler)
    // SetupAuthRoutes(api, application.AuthHandler)
    // SetupPaymentRoutes(api, application.PaymentHandler)
}
```

## 12d. internal/delivery/http/route/user_route.go (User Routes)

```go
package route

import (
    "myapp/internal/delivery/http/handler"

    "github.com/gofiber/fiber/v2"
)

// SetupUserRoutes sets up all user-related routes
func SetupUserRoutes(api fiber.Router, h *handler.UserHandler) {
    users := api.Group("/users")
    
    // Public routes
    users.Get("/", h.GetAll)           // GET /api/v1/users
    users.Get("/:id", h.GetByID)       // GET /api/v1/users/:id
    
    // Protected routes (add auth middleware later)
    users.Post("/", h.Create)          // POST /api/v1/users
    users.Put("/:id", h.Update)        // PUT /api/v1/users/:id
    users.Delete("/:id", h.Delete)     // DELETE /api/v1/users/:id
}
```

## 12e. internal/delivery/http/route/product_route.go (Contoh untuk Domain Lain)

```go
package route

import (
    "myapp/internal/delivery/http/handler"

    "github.com/gofiber/fiber/v2"
)

// SetupProductRoutes sets up all product-related routes
func SetupProductRoutes(api fiber.Router, h *handler.ProductHandler) {
    products := api.Group("/products")
    
    // Public routes
    products.Get("/", h.GetAll)              // GET /api/v1/products
    products.Get("/:id", h.GetByID)          // GET /api/v1/products/:id
    products.Get("/category/:id", h.GetByCategory) // GET /api/v1/products/category/:id
    products.Get("/search", h.Search)        // GET /api/v1/products/search?q=...
    
    // Protected routes (admin only)
    products.Post("/", h.Create)             // POST /api/v1/products
    products.Put("/:id", h.Update)           // PUT /api/v1/products/:id
    products.Delete("/:id", h.Delete)        // DELETE /api/v1/products/:id
    products.Post("/:id/images", h.UploadImage) // POST /api/v1/products/:id/images
}
```

## 12f. internal/delivery/http/route/order_route.go (Contoh untuk Domain Lain)

```go
package route

import (
    "myapp/internal/delivery/http/handler"

    "github.com/gofiber/fiber/v2"
)

// SetupOrderRoutes sets up all order-related routes
func SetupOrderRoutes(api fiber.Router, h *handler.OrderHandler) {
    orders := api.Group("/orders")
    
    // All routes protected (require authentication)
    orders.Get("/", h.GetMyOrders)           // GET /api/v1/orders
    orders.Get("/:id", h.GetByID)            // GET /api/v1/orders/:id
    orders.Post("/", h.Create)               // POST /api/v1/orders
    orders.Put("/:id/cancel", h.Cancel)      // PUT /api/v1/orders/:id/cancel
    orders.Get("/:id/invoice", h.GetInvoice) // GET /api/v1/orders/:id/invoice
    
    // Admin routes
    orders.Get("/admin/all", h.GetAllOrders) // GET /api/v1/orders/admin/all
    orders.Put("/:id/status", h.UpdateStatus) // PUT /api/v1/orders/:id/status
}
```

## 12. cmd/api/main.go (Clean dengan Dependency Container)

```go
package main

import (
    "log"
    "myapp/config"
    "myapp/internal/app"

    "github.com/gofiber/fiber/v2"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Initialize application dengan dependency container
    application, err := app.NewApplication(cfg)
    if err != nil {
        log.Fatal("Failed to initialize application:", err)
    }
    defer application.Close()

    // Initialize Fiber app
    fiberApp := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            return c.Status(code).JSON(fiber.Map{
                "success": false,
                "message": err.Error(),
            })
        },
    })

    // Setup application (middleware, routes, etc)
    application.SetupApp(fiberApp)

    // Start server
    log.Printf("Server starting on port %s", cfg.App.Port)
    if err := fiberApp.Listen(":" + cfg.App.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

## 12b. internal/app/app.go (Simplified - Delegate to Route Package)

```go
package app

import (
    "fmt"
    "myapp/config"
    "myapp/internal/delivery/http/handler"
    "myapp/internal/delivery/http/route"
    "myapp/internal/repository/postgres"
    "myapp/internal/repository/postgres/model"
    "myapp/internal/usecase"
    "myapp/pkg/database"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "gorm.io/gorm"
)

// Application holds all dependencies
type Application struct {
    Config *config.Config
    DB     *gorm.DB
    
    // Handlers
    UserHandler *handler.UserHandler
    // ProductHandler *handler.ProductHandler
    // OrderHandler   *handler.OrderHandler
}

// NewApplication creates and wires up all dependencies
func NewApplication(cfg *config.Config) (*Application, error) {
    // Initialize database
    db, err := database.NewPostgresDB(cfg.Database)
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    // Auto migrate
    if err := db.AutoMigrate(&model.UserModel{}); err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    // Initialize repositories
    userRepo := postgres.NewUserRepository(db)
    // productRepo := postgres.NewProductRepository(db)
    // orderRepo := postgres.NewOrderRepository(db)

    // Initialize use cases
    userUsecase := usecase.NewUserUsecase(userRepo)
    // productUsecase := usecase.NewProductUsecase(productRepo)
    // orderUsecase := usecase.NewOrderUsecase(orderRepo, productRepo)

    // Initialize handlers
    userHandler := handler.NewUserHandler(userUsecase)
    // productHandler := handler.NewProductHandler(productUsecase)
    // orderHandler := handler.NewOrderHandler(orderUsecase)

    return &Application{
        Config:      cfg,
        DB:          db,
        UserHandler: userHandler,
        // ProductHandler: productHandler,
        // OrderHandler:   orderHandler,
    }, nil
}

// SetupApp configures the Fiber application
func (app *Application) SetupApp(fiberApp *fiber.App) {
    // Middleware
    fiberApp.Use(recover.New())
    fiberApp.Use(logger.New())
    fiberApp.Use(cors.New())

    // Health check
    fiberApp.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
        })
    })

    // Setup routes - delegate to route package
    route.SetupRoutes(fiberApp, app)
}

// Close closes all resources
func (app *Application) Close() error {
    sqlDB, err := app.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

## 12c. internal/delivery/http/route/route.go (Updated)

```go
package route

import (
    "myapp/internal/app"

    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(fiberApp *fiber.App, handlers *app.Handlers) {
    api := fiberApp.Group("/api/v1")

    // User routes
    setupUserRoutes(api, handlers.User)
    
    // Product routes (contoh untuk scaling)
    // setupProductRoutes(api, handlers.Product)
    
    // Order routes
    // setupOrderRoutes(api, handlers.Order)
}

func setupUserRoutes(api fiber.Router, userHandler *handler.UserHandler) {
    users := api.Group("/users")
    users.Post("/", userHandler.Create)
    users.Get("/", userHandler.GetAll)
    users.Get("/:id", userHandler.GetByID)
    users.Put("/:id", userHandler.Update)
    users.Delete("/:id", userHandler.Delete)
}

// func setupProductRoutes(api fiber.Router, productHandler *handler.ProductHandler) {
//     products := api.Group("/products")
//     products.Post("/", productHandler.Create)
//     products.Get("/", productHandler.GetAll)
//     products.Get("/:id", productHandler.GetByID)
//     products.Put("/:id", productHandler.Update)
//     products.Delete("/:id", productHandler.Delete)
// }
```

## Cara Menjalankan

1. **Install dependencies:**
```bash
go mod download
```

2. **Setup database PostgreSQL**

3. **Buat file .env** sesuai dengan konfigurasi database Anda

4. **Jalankan aplikasi:**
```bash
go run cmd/api/main.go
```

5. **Test API:**
```bash
# Create user
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'

# Get all users
curl http://localhost:3000/api/v1/users

# Get user by ID
curl http://localhost:3000/api/v1/users/1

# Update user
curl -X PUT http://localhost:3000/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"newpass123"}'

# Delete user
curl -X DELETE http://localhost:3000/api/v1/users/1
```

## Solusi Import Cycle dengan Modular Routes

### **Arsitektur Baru (Clean & Modular):**

```
internal/
├── app/
│   └── app.go
│       └── Application struct (holds all handlers)
│
└── delivery/http/route/
    ├── route.go           ← Main router (orchestrator)
    ├── user_route.go      ← User routes (1 domain, 1 file)
    ├── product_route.go   ← Product routes
    └── order_route.go     ← Order routes
```

### **Dependency Flow (No Cycle!):**

```
app.go
  ↓ passes Application to
route.go (main)
  ↓ delegates to
user_route.go, product_route.go, order_route.go
  ↓ uses
handler (UserHandler, ProductHandler, etc)

✅ No circular dependency!
```

### **Keuntungan Struktur Ini:**

1. **✅ Modular**: Setiap domain punya file route sendiri
2. **✅ Scalable**: Tinggal tambah `xxx_route.go` untuk domain baru
3. **✅ Clean**: Setiap file fokus ke satu domain
4. **✅ Easy to navigate**: Developer langsung tahu dimana route untuk User
5. **✅ No import cycle**: route package hanya import handler, tidak import app

### **Pattern untuk Menambah Domain Baru:**

```go
// 1. Tambah handler di app.go
type Application struct {
    UserHandler    *handler.UserHandler
    ProductHandler *handler.ProductHandler
    PaymentHandler *handler.PaymentHandler  // ← New!
}

// 2. Buat file payment_route.go
package route

func SetupPaymentRoutes(api fiber.Router, h *handler.PaymentHandler) {
    payments := api.Group("/payments")
    payments.Post("/", h.ProcessPayment)
    payments.Get("/:id/status", h.GetStatus)
}

// 3. Register di route.go
func SetupRoutes(fiberApp *fiber.App, app *app.Application) {
    api := fiberApp.Group("/api/v1")
    
    SetupUserRoutes(api, app.UserHandler)
    SetupProductRoutes(api, app.ProductHandler)
    SetupPaymentRoutes(api, app.PaymentHandler) // ← Add here!
}
```

### **Contoh URL Structure:**

```
GET  /api/v1/users              → user_route.go
GET  /api/v1/users/:id          → user_route.go
POST /api/v1/users              → user_route.go

GET  /api/v1/products           → product_route.go
GET  /api/v1/products/:id       → product_route.go
GET  /api/v1/products/search    → product_route.go

GET  /api/v1/orders             → order_route.go
POST /api/v1/orders             → order_route.go
GET  /api/v1/orders/:id/invoice → order_route.go
```

### **Bonus: Route dengan Middleware per Domain**

```go
// user_route.go
func SetupUserRoutes(api fiber.Router, h *handler.UserHandler) {
    users := api.Group("/users")
    
    // Public routes
    users.Get("/", h.GetAll)
    users.Get("/:id", h.GetByID)
    
    // Protected routes with auth middleware
    protected := users.Group("", authMiddleware)
    protected.Post("/", h.Create)
    protected.Put("/:id", h.Update)
    protected.Delete("/:id", h.Delete)
}
```

### **File Structure Summary:**

| File | Purpose | Lines |
|------|---------|-------|
| `route.go` | Orchestrator, setup semua routes | ~15-20 |
| `user_route.go` | User domain routes | ~10-20 |
| `product_route.go` | Product domain routes | ~15-25 |
| `order_route.go` | Order domain routes | ~15-25 |

**Scalable sampai 50+ domains!** 🚀

### **Masalah:**
```go
// main.go akan jadi seperti ini jika aplikasi besar:
func main() {
    // 50+ baris config
    // 100+ baris repository initialization
    // 100+ baris usecase initialization  
    // 100+ baris handler initialization
    // Total: 300-500+ baris! 😱
}
```

### **Solusi 1: Dependency Container Pattern (Yang Saya Terapkan) ✅**

**Keuntungan:**
- ✅ `main.go` tetap clean (~30-40 baris)
- ✅ Semua dependency wiring di satu tempat (`internal/app/app.go`)
- ✅ Mudah di-maintain dan di-test
- ✅ Clear separation: setup vs runtime
- ✅ Easy to add new features (tinggal tambah di container)

**Struktur:**
```
main.go (30 baris)
    ↓ calls
app.Application (dependency container)
    ↓ initializes
    ├── Repositories
    ├── Usecases
    └── Handlers
```

**Contoh Scaling (Aplikasi Besar):**
```go
// internal/app/app.go
type Repositories struct {
    User     postgres.UserRepository
    Product  postgres.ProductRepository
    Order    postgres.OrderRepository
    Payment  postgres.PaymentRepository
    Shipping postgres.ShippingRepository
    // ... 50 repositories lainnya
}

type Usecases struct {
    User     *usecase.UserUsecase
    Product  *usecase.ProductUsecase
    Order    *usecase.OrderUsecase
    // ... 50 usecases lainnya
}

type Handlers struct {
    User    *handler.UserHandler
    Product *handler.ProductHandler
    Order   *handler.OrderHandler
    // ... 50 handlers lainnya
}

// main.go tetap 30 baris! 🎉
```

---

### **Solusi 2: Wire (Google's Dependency Injection)**

Untuk project yang sangat besar, bisa pakai [Wire](https://github.com/google/wire):

```go
// wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "myapp/internal/repository/postgres"
    "myapp/internal/usecase"
    "myapp/internal/delivery/http/handler"
)

func InitializeApp(cfg *config.Config) (*Application, error) {
    wire.Build(
        // Database
        database.NewPostgresDB,
        
        // Repositories
        postgres.NewUserRepository,
        postgres.NewProductRepository,
        
        // Usecases
        usecase.NewUserUsecase,
        usecase.NewProductUsecase,
        
        // Handlers
        handler.NewUserHandler,
        handler.NewProductHandler,
        
        // Application
        NewApplication,
    )
    return nil, nil
}
```

Wire akan **generate code otomatis** untuk dependency injection!

---

### **Solusi 3: Uber's Fx (Dependency Injection Framework)**

Untuk yang suka framework-based:

```go
package main

import (
    "go.uber.org/fx"
    "myapp/internal/repository/postgres"
    "myapp/internal/usecase"
    "myapp/internal/delivery/http/handler"
)

func main() {
    fx.New(
        // Provide dependencies
        fx.Provide(
            config.LoadConfig,
            database.NewPostgresDB,
            postgres.NewUserRepository,
            usecase.NewUserUsecase,
            handler.NewUserHandler,
            // ... dll
        ),
        
        // Invoke startup
        fx.Invoke(startServer),
    ).Run()
}
```

---

### **Perbandingan:**

| Solusi | Kompleksitas | Best For | Learning Curve |
|--------|--------------|----------|----------------|
| **Manual Container** | Simple | Small-Medium projects | Easy ⭐ |
| **Wire (Google)** | Medium | Large projects | Medium ⭐⭐ |
| **Fx (Uber)** | Complex | Enterprise projects | Hard ⭐⭐⭐ |

---

### **Rekomendasi:**

1. **Project Kecil-Menengah (1-20 domains)**: 
   - ✅ Manual Container Pattern (yang saya implementasikan)
   
2. **Project Besar (20-50 domains)**:
   - ✅ Wire (code generation)
   
3. **Project Enterprise (50+ domains)**:
   - ✅ Fx atau custom DI framework

---

### **Best Practices:**

```go
// ✅ GOOD: Grouping dependencies
type Repositories struct {
    User    UserRepository
    Product ProductRepository
}

type Usecases struct {
    User    *UserUsecase
    Product *ProductUsecase
}

// ❌ BAD: Flat structure
type Application struct {
    UserRepo    UserRepository
    ProductRepo ProductRepository
    UserUC      *UserUsecase
    ProductUC   *ProductUsecase
    // ... 100 fields! 😱
}
```

**Perubahan Penting - Usecase yang Benar-benar Clean:**

Sekarang struktur sudah mengikuti prinsip Clean Architecture dengan benar:

1. **Domain Entity (`internal/domain/entity/user.go`)**: 
   - **Benar-benar clean**, tidak ada dependency eksternal (GORM, JSON tags, dll)
   - Hanya berisi business logic murni
   - Tidak tahu tentang database atau framework apapun

2. **Domain Repository Interface (`internal/domain/repository/user_repository.go`)**:
   - Mendefinisikan **domain errors** (`ErrNotFound`, dll) yang framework-agnostic
   - Interface contract tanpa dependency eksternal
   - Tidak tahu tentang GORM

3. **Use Case (`internal/usecase/user_usecase.go`)**:
   - **TIDAK import GORM** atau library database lainnya
   - Hanya bekerja dengan domain entity dan repository interface
   - Menggunakan domain errors, bukan GORM errors
   - Pure business logic

4. **Repository Implementation (`internal/repository/postgres/user_repository.go`)**:
   - **Satu-satunya tempat yang boleh import GORM**
   - Bertanggung jawab untuk **error translation**: GORM errors → Domain errors
   - Method `handleError()` mengkonversi error
   - Bekerja dengan database model

5. **Database Model (`internal/repository/postgres/model/user_model.go`)**:
   - Boleh menggunakan GORM tags karena ada di **infrastructure layer**
   - Berisi converter methods: `ToEntity()` dan `FromEntity()`
   - Bertanggung jawab untuk mapping antara database dan domain

**Alur Error Handling yang Benar:**
```
GORM Error (gorm.ErrRecordNotFound)
    ↓ [Repository converts]
Domain Error (repository.ErrNotFound)
    ↓ [Usecase handles]
Usecase Error (usecase.ErrUserNotFound)
    ↓ [Handler converts]
HTTP Response (404 Not Found)
```

**Layer-layer:**
- **Domain/Entity**: Business entities (CLEAN, no external deps)
- **Domain/Repository Interface**: Contract + Domain Errors (CLEAN)
- **Usecase**: Business logic (CLEAN, no GORM/DB libs)
- **Repository Implementation**: Database operations + Error translation (boleh pakai GORM)
- **Database Model**: ORM mappings (boleh pakai GORM tags)
- **Delivery/Handler**: HTTP handlers dan routing

**Dependency Rule:**
```
┌─────────────────────────────────────────┐
│  Entities (Pure Go, no dependencies)   │ ← Innermost
├─────────────────────────────────────────┤
│  Use Cases (domain errors only)        │
├─────────────────────────────────────────┤
│  Interface Adapters (Repository Interface)│
├─────────────────────────────────────────┤
│  Frameworks & Drivers (GORM, Fiber)    │ ← Outermost
└─────────────────────────────────────────┘

Dependencies always point INWARD →
```

**Keuntungan Pendekatan Ini:**
- ✅ Domain entity benar-benar independen
- ✅ **Usecase tidak tahu tentang GORM atau database apapun**
- ✅ **Error handling yang clean dengan domain errors**
- ✅ Mudah testing (mock tanpa database)
- ✅ Bisa ganti ORM (GORM → SQLx → sqlc) tanpa ubah usecase
- ✅ Bisa ganti database (PostgreSQL → MySQL → MongoDB) tanpa ubah usecase
- ✅ Separation of concerns yang jelas
- ✅ Scalable dan maintainable

**Yang Boleh Import GORM:**
- ❌ Domain Entity → TIDAK
- ❌ Repository Interface → TIDAK
- ❌ Use Case → TIDAK
- ❌ Handler → TIDAK
- ✅ Repository Implementation → YA (hanya di sini!)
- ✅ Database Model → YA

**Prinsip Penting:**
> "Inner layers should not depend on outer layers"
> Usecase adalah inner layer, GORM adalah outer layer
> Jadi Usecase TIDAK BOLEH depend on GORM

**Trade-off:**
- Butuh mapping code (ToEntity/FromEntity)
- Sedikit lebih verbose
- Tapi lebih flexible dan testable!

## Pendekatan Alternatif (Pragmatic)

Jika ingin lebih pragmatis untuk project kecil-menengah, bisa pakai struct tags tapi dengan catatan:
- Gunakan tags yang framework-agnostic seperti `json`
- Atau terima bahwa ini "pragmatic clean architecture"
- Cocok untuk rapid development

Tapi untuk project besar atau yang butuh flexibility tinggi, **pisahkan entity dan model seperti di atas adalah pilihan terbaik**.