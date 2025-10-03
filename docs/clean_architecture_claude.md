# Go Fiber + GORM Clean Architecture Project

## Struktur Direktori

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
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
│   │           └── route.go
│   └── repository/
│       └── postgres/
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

## 6. internal/domain/entity/user.go

```go
package entity

import (
    "time"

    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Name      string         `gorm:"size:255;not null" json:"name"`
    Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## 7. internal/domain/repository/user_repository.go

```go
package repository

import (
    "context"
    "myapp/internal/domain/entity"
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

## 8. internal/repository/postgres/user_repository.go

```go
package postgres

import (
    "context"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"

    "gorm.io/gorm"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]entity.User, error) {
    var users []entity.User
    err := r.db.WithContext(ctx).Find(&users).Error
    return users, err
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}
```

## 9. internal/usecase/user_usecase.go

```go
package usecase

import (
    "context"
    "errors"
    "myapp/internal/domain/entity"
    "myapp/internal/domain/repository"

    "gorm.io/gorm"
)

type UserUsecase interface {
    Create(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id uint) (*entity.User, error)
    GetAll(ctx context.Context) ([]entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uint) error
}

type userUsecase struct {
    userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
    return &userUsecase{
        userRepo: userRepo,
    }
}

func (u *userUsecase) Create(ctx context.Context, user *entity.User) error {
    // Business logic validation
    if user.Name == "" || user.Email == "" {
        return errors.New("name and email are required")
    }

    // Check if email already exists
    _, err := u.userRepo.FindByEmail(ctx, user.Email)
    if err == nil {
        return errors.New("email already exists")
    }

    return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*entity.User, error) {
    user, err := u.userRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return user, nil
}

func (u *userUsecase) GetAll(ctx context.Context) ([]entity.User, error) {
    return u.userRepo.FindAll(ctx)
}

func (u *userUsecase) Update(ctx context.Context, user *entity.User) error {
    // Check if user exists
    _, err := u.userRepo.FindByID(ctx, user.ID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("user not found")
        }
        return err
    }

    return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
    // Check if user exists
    _, err := u.userRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("user not found")
        }
        return err
    }

    return u.userRepo.Delete(ctx, id)
}
```

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
    userUsecase usecase.UserUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
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

## 11. internal/delivery/http/route/route.go

```go
package route

import (
    "myapp/internal/delivery/http/handler"

    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler) {
    api := app.Group("/api/v1")

    // User routes
    users := api.Group("/users")
    users.Post("/", userHandler.Create)
    users.Get("/", userHandler.GetAll)
    users.Get("/:id", userHandler.GetByID)
    users.Put("/:id", userHandler.Update)
    users.Delete("/:id", userHandler.Delete)
}
```

## 12. cmd/api/main.go

```go
package main

import (
    "log"
    "myapp/config"
    "myapp/internal/delivery/http/handler"
    "myapp/internal/delivery/http/route"
    "myapp/internal/domain/entity"
    "myapp/internal/repository/postgres"
    "myapp/internal/usecase"
    "myapp/pkg/database"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Initialize database
    db, err := database.NewPostgresDB(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto migrate
    if err := db.AutoMigrate(&entity.User{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    // Initialize repositories
    userRepo := postgres.NewUserRepository(db)

    // Initialize use cases
    userUsecase := usecase.NewUserUsecase(userRepo)

    // Initialize handlers
    userHandler := handler.NewUserHandler(userUsecase)

    // Initialize Fiber app
    app := fiber.New(fiber.Config{
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

    // Middleware
    app.Use(recover.New())
    app.Use(logger.New())
    app.Use(cors.New())

    // Setup routes
    route.SetupRoutes(app, userHandler)

    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
        })
    })

    // Start server
    log.Printf("Server starting on port %s", cfg.App.Port)
    if err := app.Listen(":" + cfg.App.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
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

## Penjelasan Clean Architecture

**Layer-layer:**
- **Domain/Entity**: Business entities dan interface repository
- **Repository**: Implementasi database operations
- **Usecase**: Business logic layer
- **Delivery/Handler**: HTTP handlers dan routing

**Keuntungan:**
- Separation of concerns
- Testable
- Independent dari framework
- Mudah maintenance dan scalable