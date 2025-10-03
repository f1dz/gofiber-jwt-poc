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
│   │   │   └── user.go              # Clean entity, no external dependencies
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
│           ├── model/
│           │   └── user_model.go    # Database model with GORM tags
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

## 9. internal/usecase/user_usecase.go (Clean - Tanpa GORM)

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
        return ErrInvalidInput
    }

    // Check if email already exists
    // Repository harus return error yang jelas (bukan GORM error)
    _, err := u.userRepo.FindByEmail(ctx, user.Email)
    if err == nil {
        return ErrEmailAlreadyExists
    }
    // Jika error bukan "not found", berarti error lain
    if err != repository.ErrNotFound {
        return err
    }

    return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*entity.User, error) {
    user, err := u.userRepo.FindByID(ctx, id)
    if err != nil {
        // Usecase tidak tahu tentang GORM errors
        // Repository bertanggung jawab convert GORM error ke domain error
        if err == repository.ErrNotFound {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return user, nil
}

func (u *userUsecase) GetAll(ctx context.Context) ([]entity.User, error) {
    return u.userRepo.FindAll(ctx)
}

func (u *userUsecase) Update(ctx context.Context, user *entity.User) error {
    // Business logic: validate input
    if user.Name == "" || user.Email == "" {
        return ErrInvalidInput
    }

    // Check if user exists
    _, err := u.userRepo.FindByID(ctx, user.ID)
    if err != nil {
        if err == repository.ErrNotFound {
            return ErrUserNotFound
        }
        return err
    }

    return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
    // Check if user exists
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
    "myapp/internal/repository/postgres"
    "myapp/internal/repository/postgres/model"
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

    // Auto migrate - gunakan database model, bukan entity
    if err := db.AutoMigrate(&model.UserModel{}); err != nil {
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