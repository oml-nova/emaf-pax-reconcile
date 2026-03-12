# pl-api-reports-command-service — Complete Architecture Reference

> A full implementation guide covering every pattern in this repository.
> Use this as a blueprint when creating a new Go service with the same architecture.

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Bootstrap & Entry Point](#2-bootstrap--entry-point)
3. [Configuration Loading](#3-configuration-loading)
4. [Logger (superlog)](#4-logger-superlog)
5. [HTTP Server & Middleware](#5-http-server--middleware)
6. [Route Registration](#6-route-registration)
7. [Handler Pattern](#7-handler-pattern)
8. [Service / Business Logic Pattern](#8-service--business-logic-pattern)
9. [Repository / Entity Pattern (GORM)](#9-repository--entity-pattern-gorm)
10. [Database — PostgreSQL](#10-database--postgresql)
11. [Database — MongoDB](#11-database--mongodb)
12. [Raw SQL Queries Pattern](#12-raw-sql-queries-pattern)
13. [Advisory Locks](#13-advisory-locks)
14. [Kafka — Producer](#14-kafka--producer)
15. [Kafka — Consumer](#15-kafka--consumer)
16. [SQS Consumer & Producer](#16-sqs-consumer--producer)
17. [Consumer Wiring (consumers.go)](#17-consumer-wiring-consumersgo)
18. [Kafka & SQS Message Types](#18-kafka--sqs-message-types)
19. [Event Audit Trail (EventStatuses)](#19-event-audit-trail-eventstatuses)
20. [Docker](#20-docker)
21. [VSCode launch.json](#21-vscode-launchjson)
22. [Environment Variables Reference](#22-environment-variables-reference)
23. [Key Dependencies (go.mod)](#23-key-dependencies-gomod)
24. [Concurrency Patterns](#24-concurrency-patterns)
25. [Error Handling Patterns](#25-error-handling-patterns)
26. [Full Interfaces Reference](#26-full-interfaces-reference)
27. [RPC Clients (Internal Service-to-Service Calls)](#27-rpc-clients-internal-service-to-service-calls)
28. [S3 Client & Uploader](#28-s3-client--uploader)
29. [FTP/SFTP Client (R365 Partner Exports)](#29-ftpsftp-client-r365-partner-exports)
30. [Firebase Config](#30-firebase-config)
31. [Vision AI Config](#31-vision-ai-config)
32. [Shared Helpers](#32-shared-helpers)

---

## 1. Project Structure

```
.
├── cmd/
│   └── main.go                          # Entry point — bootstraps everything
├── internal/
│   ├── server.go                        # Gin engine setup, middleware, route wiring
│   ├── consumers.go                     # Kafka & SQS consumer initialisation
│   ├── config/
│   │   ├── environment.go               # Config-service client + env hydration
│   │   ├── databases/
│   │   │   ├── postgres.go              # GORM PostgreSQL (standard + IAM)
│   │   │   └── mongodb.go               # MongoDB client (primary + rewards DB)
│   │   ├── kafka/
│   │   │   ├── kafka_config.go          # Auth helpers + broker resolution
│   │   │   ├── producer.go              # Singleton sync producer
│   │   │   ├── consumer.go              # Consumer-group service
│   │   │   └── types.go                 # All Kafka message structs + event constants
│   │   ├── sqs/
│   │   │   ├── sqs.go                   # SQS service (send, receive, delete, dispatch)
│   │   │   └── types.go                 # SQS event-type constants
│   │   ├── rpc/                         # RPC client setup
│   │   ├── s3/                          # S3 client
│   │   ├── ftp/                         # FTP/SFTP client
│   │   ├── firebase.go                  # Firebase app init
│   │   └── visionai.go                  # Vision AI config
│   ├── handlers/
│   │   ├── etl/                         # HTTP + consumer handlers for ETL domain
│   │   │   ├── api.orders.go
│   │   │   ├── api.customers.go
│   │   │   ├── api.restaurants.go
│   │   │   ├── api.transactions.go
│   │   │   ├── consumer.orders.go
│   │   │   ├── consumer.customers.go
│   │   │   ├── consumer.giftcard.go
│   │   │   ├── consumer.restaurants.go
│   │   │   ├── consumer.transactions.go
│   │   │   ├── consumer.visionai.go
│   │   │   └── consumer.voice_ordering.go
│   │   ├── exports/                     # Partner CSV/JSON export handlers
│   │   ├── resiliency/                  # Debugging / repair endpoints
│   │   └── visionai/                    # Vision AI ingestion
│   ├── routes/
│   │   ├── health.go
│   │   ├── orders.go
│   │   ├── internal.go
│   │   ├── resiliency.go
│   │   ├── partner_exports.go
│   │   └── visionai.go
│   ├── repositories/
│   │   ├── automigrate.go               # GORM AutoMigrate runner
│   │   ├── advisory_lock_helper.go      # pg_advisory_xact_lock wrapper
│   │   ├── entities/                    # 40+ GORM model structs
│   │   └── queries/                     # Raw .sql files, loaded at runtime
│   │       ├── orders/
│   │       ├── partners/
│   │       └── resiliency/
│   └── services/
│       ├── etl/
│       │   ├── orders/                  # Core order ETL (extract → validate → transform → save)
│       │   ├── customers/
│       │   ├── restaurants/
│       │   ├── giftcards/
│       │   └── visionai/
│       ├── exports/                     # Partner report queries + CSV generation
│       ├── notifications/               # Firebase push notifications
│       └── resiliency/                  # Validation, comparison, repair
├── pkg/
│   └── superlog/
│       ├── superlog.go                  # Logger singleton + Gin middleware
│       └── shippers.go                  # ConsoleShipper, FileShipper, HTTPShipper
├── test/
│   └── integration/
│       └── order_sync/
│           └── order_sync_test.go
├── docs/                                # Swagger generated files
├── Dockerfile
├── go.mod
└── go.sum
```

---

## 2. Bootstrap & Entry Point

**File:** `cmd/main.go`

The startup sequence is strictly ordered. Each step must succeed before the next begins.

```go
package main

import (
    "fmt"
    "os"
    app "reports-command-service/internal"
    "reports-command-service/internal/config"
    database "reports-command-service/internal/config/databases"
    "reports-command-service/internal/repositories"
    logger "reports-command-service/pkg/superlog"
    "strings"
)

func main() {
    // STEP 1 — Hydrate env vars from remote config service
    err := config.LoadEnvironment()

    // STEP 2 — Init logger BEFORE checking err so panics are logged
    isRemoteLoggerEnabled := strings.ToLower(os.Getenv("REMOTE_LOG")) == "true"
    if isRemoteLoggerEnabled {
        remoteLoggerURL := os.Getenv("LOGGER_HOST")
        logger.NewLogger(
            &logger.ConsoleShipper{},
            &logger.HTTPShipper{URL: remoteLoggerURL},
        )
    } else {
        logger.NewLogger(
            &logger.ConsoleShipper{},
            &logger.FileShipper{FilePath: "app.log"},
        )
    }
    fmt.Println("Remote Logger Enabled: ", isRemoteLoggerEnabled)

    if err != nil {
        panic(err.Error())
    }

    // STEP 3 — Connect databases
    err = database.InitPGDB()
    if err != nil {
        panic(err.Error())
    }
    err = database.InitMongoDB()
    if err != nil {
        panic(err.Error())
    }

    // STEP 4 — Run GORM migrations
    repositories.AutoMigrate()

    // STEP 5 — Build HTTP router
    r := app.RegisterRoutes()

    // STEP 6 — Start message consumers in background goroutines
    go func() {
        app.InitSQSConsumer()
        logger.Log().Info("SQS Consumer Initialized", nil)
    }()
    go app.InitKafkaConsumer()

    logger.Log().Info("Kafka and SQS Consumers Initialized", nil)

    // STEP 7 — Block on HTTP server (runs on :8080 by default via gin)
    r.StartServer()
}
```

**Key decisions:**
- Logger is initialised before the config error is checked so that panic stack traces always reach shippers.
- Consumers run in goroutines — the main goroutine blocks on `r.StartServer()` forever.
- `InitKafkaConsumer()` itself blocks internally (calls `kafkaService.Start(ctx)` which waits for the `ready` channel), so it must be in a goroutine.

---

## 3. Configuration Loading

**File:** `internal/config/environment.go`

### Philosophy

All non-trivial configuration lives in a centralised **Config Service** (an internal HTTP service). At startup, this service is called to get a flat or nested JSON blob, which is then recursively flattened into `os.Setenv` calls. The application thereafter reads everything via `os.Getenv`.

For `test` or `local` environments the config-service call is skipped entirely.

### Full Code

```go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "unicode"

    "github.com/go-resty/resty/v2"
)

const (
    configServiceBaseURL = "http://config-service"
    serviceID            = "report-command-service"
    serviceRole          = "admin"
)

type authResponse struct {
    Token string `json:"token"`
}

// sanitizeMultilineJSONValue replaces bare newlines/carriage-returns inside
// a specific JSON string value so that json.Unmarshal does not choke.
func sanitizeMultilineJSONValue(body, key string) string {
    needle := fmt.Sprintf("\"%s\": \"", key)
    start := strings.Index(body, needle)
    if start == -1 {
        return body
    }
    valueStart := start + len(needle)
    i := valueStart
    for i < len(body) {
        if body[i] == '"' {
            if i == valueStart || body[i-1] != '\\' {
                break
            }
        }
        i++
    }
    if i >= len(body) {
        return body
    }
    value := body[valueStart:i]
    value = strings.ReplaceAll(value, "\r", "\\r")
    value = strings.ReplaceAll(value, "\n", "\\n")
    return body[:valueStart] + value + body[i:]
}

func getConfigDetailsFromConfigService() (map[string]interface{}, error) {
    filePath := os.Getenv("CONFIG_PATH")
    if filePath == "" {
        return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
    }
    filePath += fmt.Sprintf("/%s/config.json", serviceID)
    client := resty.New()

    // --- Auth ---
    var authResp authResponse
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(map[string]string{"serviceId": serviceID, "role": serviceRole}).
        SetResult(&authResp).
        Post(fmt.Sprintf("%s/auth", configServiceBaseURL))
    if err != nil {
        return nil, fmt.Errorf("failed to authenticate with config service: %w", err)
    }
    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("authentication failed with status code: %d", resp.StatusCode())
    }
    if authResp.Token == "" {
        return nil, fmt.Errorf("no token received from config service")
    }

    // --- Fetch config ---
    resp, err = client.R().
        SetHeader("Authorization", fmt.Sprintf("Bearer %s", authResp.Token)).
        Get(fmt.Sprintf("%s/secure/services/%s/configs/%s", configServiceBaseURL, serviceID, filePath))
    if err != nil {
        return nil, fmt.Errorf("failed to fetch config from config service: %w", err)
    }
    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("failed to fetch config with status code: %d", resp.StatusCode())
    }

    body := resp.Body()
    bodyStr := string(body)

    // Handle double-encoded JSON (body is a quoted JSON string)
    if strings.HasPrefix(bodyStr, "\"") && strings.HasSuffix(bodyStr, "\"") {
        var unquoted string
        if err := json.Unmarshal(body, &unquoted); err == nil {
            body = []byte(unquoted)
            bodyStr = unquoted
        }
    }

    // Fix malformed KAFKA_HOST array-in-string
    if strings.Contains(bodyStr, "\"KAFKA_HOST\": [\"[\"") {
        bodyStr = strings.ReplaceAll(bodyStr, "\"KAFKA_HOST\": [\"[\"", "\"KAFKA_HOST\": [\"")
        bodyStr = strings.ReplaceAll(bodyStr, "\"]\"]", "\"]")
        body = []byte(bodyStr)
    }

    // Sanitize multiline PEM private key (Firebase)
    bodyStr = sanitizeMultilineJSONValue(bodyStr, "FIREBASE_PRIVATE_KEY")
    body = []byte(bodyStr)

    var configResp map[string]interface{}
    if err := json.Unmarshal(body, &configResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config response: %w, body: %s", err, string(body))
    }
    return configResp, nil
}

// setEnvFromConfig recursively walks the config map and calls os.Setenv.
// Arrays are JSON-encoded back to string (e.g. KAFKA_HOST = '["broker:9094"]').
func setEnvFromConfig(config map[string]interface{}, prefix string) error {
    for key, value := range config {
        envKey := key
        if prefix != "" {
            envKey = fmt.Sprintf("%s_%s", prefix, key)
        }
        envKey = toEnvVarName(envKey)

        switch v := value.(type) {
        case map[string]interface{}:
            if err := setEnvFromConfig(v, envKey); err != nil {
                return err
            }
        case []interface{}:
            jsonBytes, err := json.Marshal(v)
            if err != nil {
                return fmt.Errorf("failed to marshal array for key %s: %w", envKey, err)
            }
            if err := os.Setenv(envKey, string(jsonBytes)); err != nil {
                return fmt.Errorf("failed to set environment variable %s: %w", envKey, err)
            }
        default:
            valueStr := fmt.Sprintf("%v", v)
            if err := os.Setenv(envKey, valueStr); err != nil {
                return fmt.Errorf("failed to set environment variable %s: %w", envKey, err)
            }
        }
    }
    return nil
}

// toEnvVarName converts camelCase, kebab-case, dot-notation → UPPER_SNAKE_CASE
func toEnvVarName(s string) string {
    s = strings.ReplaceAll(s, ".", "_")
    s = strings.ReplaceAll(s, "-", "_")
    s = strings.ReplaceAll(s, " ", "_")

    var result strings.Builder
    for i, r := range s {
        if i > 0 && unicode.IsUpper(r) {
            prev := rune(s[i-1])
            if unicode.IsLower(prev) || unicode.IsDigit(prev) {
                result.WriteRune('_')
            }
        }
        result.WriteRune(unicode.ToUpper(r))
    }
    return result.String()
}

func flattenConfig(config map[string]interface{}) error {
    return setEnvFromConfig(config, "")
}

// LoadEnvironment is the public entry point called in main().
func LoadEnvironment() error {
    env := os.Getenv("NODE_ENV")
    if env == "" {
        return fmt.Errorf("NODE_ENV environment variable is not set")
    }

    // Skip config service in local/test environments
    if env == "test" || env == "local" {
        fmt.Printf("Skipping config service for environment: %s\n", env)
        if os.Getenv("AWS_REGION") == "" {
            return fmt.Errorf("AWS_REGION is not set")
        }
        return nil
    }

    config, err := getConfigDetailsFromConfigService()
    if err != nil {
        return fmt.Errorf("failed to get config from config service: %w", err)
    }
    if config == nil {
        return fmt.Errorf("no config found for config service path: %s", os.Getenv("CONFIG_PATH"))
    }
    if err := flattenConfig(config); err != nil {
        return fmt.Errorf("failed to set environment variables from config: %w", err)
    }
    if os.Getenv("AWS_REGION") == "" {
        return fmt.Errorf("AWS_REGION is not set")
    }
    return nil
}
```

### Flow Diagram

```
main()
  └─ LoadEnvironment()
       ├─ NODE_ENV == "test"|"local" → skip, validate AWS_REGION
       └─ else
            ├─ POST http://config-service/auth  { serviceId, role }
            │    → JWT token
            ├─ GET  http://config-service/secure/services/{id}/configs/{path}
            │    → JSON blob (possibly double-encoded)
            ├─ sanitize: double-encoding, KAFKA_HOST malform, PEM newlines
            ├─ json.Unmarshal → map[string]interface{}
            └─ setEnvFromConfig (recursive)
                 ├─ nested map  → recurse with prefix
                 ├─ []interface{} → json.Marshal → os.Setenv("KEY", `[...]`)
                 └─ scalar → fmt.Sprintf → os.Setenv("KEY", "value")
```

---

## 4. Logger (superlog)

**Package:** `pkg/superlog/`

### Design: Singleton + Pluggable Shippers

The logger is a singleton initialised once via `sync.Once`. It holds a slice of `LogShipper` implementations. Every log call fires all shippers **in goroutines** (fire-and-forget), so logging is never on the hot path.

### superlog.go — Full Code

```go
package superlog

import (
    "bytes"
    "fmt"
    "github.com/gin-gonic/gin"
    "io"
    "sync"
    "time"
)

type LogLevel int

const (
    DEBUG LogLevel = iota // 0
    INFO                  // 1
    WARN                  // 2
    ERROR                 // 3
)

type LogEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    Level       LogLevel               `json:"level"`
    Message     string                 `json:"message"`
    Correlation string                 `json:"correlation,omitempty"`
    Data        map[string]interface{} `json:"data,omitempty"`
}

type ILogger interface {
    Debug(message string, data map[string]interface{}) *LogEvent
    Info(message string, data map[string]interface{}) *LogEvent
    Warn(message string, data map[string]interface{}) *LogEvent
    Error(message string, data map[string]interface{}) *LogEvent
}

type LogShipper interface {
    Ship(event LogEvent) error
}

type Logger struct {
    shippers []LogShipper
}

var (
    instance *Logger
    once     sync.Once
)

// Log returns the singleton. Call after NewLogger().
func Log() *Logger { return instance }

// NewLogger creates the singleton. Subsequent calls are no-ops (sync.Once).
func NewLogger(shippers ...LogShipper) *Logger {
    once.Do(func() {
        instance = &Logger{shippers: shippers}
    })
    return instance
}

func (l *Logger) log(level LogLevel, message string, data map[string]interface{}) *LogEvent {
    event := LogEvent{
        Timestamp: time.Now(),
        Level:     level,
        Message:   message,
        Data:      data,
    }
    for _, shipper := range l.shippers {
        go func() {
            if err := shipper.Ship(event); err != nil {
                fmt.Printf("Error shipping log event: %v\n", err)
            }
        }()
    }
    return &event
}

func (l *Logger) Debug(message string, data map[string]interface{}) *LogEvent {
    return l.log(DEBUG, message, data)
}
func (l *Logger) Info(message string, data map[string]interface{}) *LogEvent {
    return l.log(INFO, message, data)
}
func (l *Logger) Warn(message string, data map[string]interface{}) *LogEvent {
    return l.log(WARN, message, data)
}
func (l *Logger) Error(message string, data map[string]interface{}) *LogEvent {
    return l.log(ERROR, message, data)
}

// Correlate chains a correlation ID onto a LogEvent (fluent API).
func (e *LogEvent) Correlate(correlationID string) *LogEvent {
    e.Correlation = correlationID
    return e
}

// GinMiddleware logs every HTTP request after it completes.
func GinMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var bodyBytes []byte
        if c.Request.Body != nil {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
        }
        // Restore body so handlers can read it too
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

        start := time.Now()
        c.Next()
        duration := time.Since(start)

        var bodyInfo interface{}
        if len(bodyBytes) > 0 {
            bodyInfo = string(bodyBytes)
        } else {
            bodyInfo = "empty body"
        }

        Log().Info("API: "+c.Request.URL.Path, map[string]interface{}{
            "method":    c.Request.Method,
            "path":      c.Request.URL.Path,
            "status":    c.Writer.Status(),
            "duration":  duration.String(),
            "clientIP":  c.ClientIP(),
            "userAgent": c.Request.UserAgent(),
            "body":      bodyInfo,
            "header":    c.Request.Header,
        })
    }
}
```

### shippers.go — Full Code

```go
package superlog

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

// ConsoleShipper — JSON line to stdout
type ConsoleShipper struct{}

func (s *ConsoleShipper) Ship(event LogEvent) error {
    jsonData, err := json.Marshal(event)
    if err != nil {
        return err
    }
    fmt.Println(string(jsonData))
    return nil
}

// FileShipper — appends JSON lines to a file
type FileShipper struct {
    FilePath string
}

func (s *FileShipper) Ship(event LogEvent) error {
    jsonData, err := json.Marshal(event)
    if err != nil {
        return err
    }
    file, err := os.OpenFile(s.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer func(file *os.File) {
        if err := file.Close(); err != nil {
            fmt.Println("Error closing file")
        }
    }(file)
    _, err = file.Write(append(jsonData, '\n'))
    return err
}

// HTTPShipper — POSTs JSON to a remote endpoint
type HTTPShipper struct {
    URL string
}

func (s *HTTPShipper) Ship(event LogEvent) error {
    jsonData, err := json.Marshal(event)
    if err != nil {
        return err
    }
    _, err = http.Post(s.URL, "application/json", bytes.NewBuffer(jsonData))
    return err
}
```

### Usage in application code

```go
// Import
import superlog "reports-command-service/pkg/superlog"

// Simple
superlog.Log().Info("Order processed", map[string]interface{}{
    "orderRefId": refId,
    "retryCount": count,
})

// With correlation ID
superlog.Log().Error("Sync failed", map[string]interface{}{
    "error": err.Error(),
}).Correlate(requestId)

// nil data is allowed
superlog.Log().Info("Kafka Consumer Initialized", nil)
```

---

## 5. HTTP Server & Middleware

**File:** `internal/server.go`

### GinEngine wrapper

The Gin engine is wrapped in a custom struct so it can implement the `Server` interface and allow method chaining.

```go
package app

import (
    "log"
    "reports-command-service/docs"
    routes "reports-command-service/internal/routes"
    "reports-command-service/pkg/superlog"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    swaggerfiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

type GinEngine struct {
    *gin.Engine
}

type Server interface {
    addRoutes() Server
    addMiddleware() Server
    StartServer()
}

func setupGin() GinEngine {
    r := gin.Default()

    config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    config.AllowHeaders = []string{"*"}
    config.ExposeHeaders = []string{"Content-Length"}
    config.AllowCredentials = true

    r.Use(cors.New(config))
    return GinEngine{Engine: r}
}

func (r GinEngine) addMiddleware() Server {
    r.Use(superlog.GinMiddleware()) // request/response logging
    r.Use(gin.Recovery())           // panic recovery → 500
    return r
}

func (r GinEngine) addRoutes() Server {
    routes.InitHealthRoutes(r.Engine)
    routes.InitOrderETLRoutes(r.Engine)
    routes.InitResiliencyRoutes(r.Engine)
    routes.InitInternalSyncRoutes(r.Engine)
    routes.InitPartnerExportRoutes(r.Engine)
    routes.DataIngestionRoutes(r.Engine)
    docs.SwaggerInfo.BasePath = "/api"
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
    return r
}

func RegisterRoutes() GinEngine {
    r := setupGin()
    r.RedirectFixedPath = false
    r.RedirectTrailingSlash = false
    r.addMiddleware()
    r.addRoutes()
    return r
}

func (r GinEngine) StartServer() {
    err := r.Run() // defaults to :8080, override with PORT env var
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

### Middleware stack (in order)

| # | Middleware | Purpose |
|---|-----------|---------|
| 1 | `cors.New(config)` | Added in `setupGin()` — applied first, before logging |
| 2 | `superlog.GinMiddleware()` | Reads + restores body, logs after `c.Next()` |
| 3 | `gin.Recovery()` | Converts panics → HTTP 500 |

---

## 6. Route Registration

Each domain has its own `routes/*.go` file. All are registered on the raw `*gin.Engine`.

### routes/orders.go
```go
func InitOrderETLRoutes(a *gin.Engine) {
    sync := a.Group("/sync")
    sync.GET("/:refId", etl.SyncOrder)
}
```

### routes/resiliency.go
```go
func InitResiliencyRoutes(a *gin.Engine) {
    res := a.Group("/internal/resiliency")
    res.GET("/slow-queries", resiliency.SlowQueries)
    res.POST("/kafka/publish", resiliency.PublishToKafka)
    res.POST("/kafka/consume", resiliency.ConsumeMsg)
    res.POST("/execute-query", resiliency.ExecuteSelectQuery)

    internal := a.Group("/internal/orders")
    internal.GET("/sync-status/:refId", resiliency.GetOrderSyncStatus)
    internal.POST("/missing-orders", resiliency.GetMissingOrdersAsync)
    internal.GET("/missing-orders-alert", resiliency.GetMissingOrdersAsyncAlert)
    internal.POST("/sync-missing", etl.SyncMissingOrders)
    internal.GET("/failed-orders", resiliency.GetFailedOrders)
    internal.GET("/success-orders", resiliency.GetSuccessOrders)
    internal.GET("/compare-order/:refId", resiliency.CompareOrder)
    internal.POST("/process-orders-async", resiliency.ProcessOrdersByDateRangeAsync)
    internal.POST("/process-orders", resiliency.ProcessOrdersByDateRange)
    internal.GET("/process-order-async/:refId", resiliency.ProcessSingleOrderAsync)
    internal.POST("/sales-mismatch", resiliency.GetSalesMismatch)
    internal.GET("/sync-transaction/:transactionRefId", resiliency.SyncTransaction)
    internal.POST("/sync-transaction-status", resiliency.SyncSettledTransactionStatus)

    customers := a.Group("/internal/customers")
    customers.GET("/sync-status/:userRefId", resiliency.GetCustomerSyncStatus)
}
```

### routes/internal.go
```go
func InitInternalSyncRoutes(a *gin.Engine) {
    sync := a.Group("/sync/internal")
    sync.GET("/restaurant-date-time-settings/:refId", etl.SyncRestaurantDateTimeSettings)
    sync.GET("/customer-details/:refId", etl.SyncCustomerDetails)
    sync.GET("/order-transaction-status/:refId", etl.SyncOrderTransactionStatus)
}
```

### routes/partner_exports.go
```go
func InitPartnerExportRoutes(a *gin.Engine) {
    partners := a.Group("/exports/partners")
    partners.GET("/r365/sync", exports.R365Sync)           // parallel aggregate
    partners.POST("/sales-detail", exports.GetSalesDetail)
    partners.POST("/menu-items", exports.GetMenuItems)
    partners.POST("/payments", exports.GetPayments)
    partners.POST("/tenders", exports.GetTenders)
    partners.POST("/tax-details", exports.GetTaxDetails)
    partners.POST("/surcharges", exports.GetSurcharges)
    partners.POST("/discounts", exports.GetDiscounts)
    partners.POST("/refunds", exports.GetRefunds)
    partners.POST("/voids", exports.GetVoids)
    partners.POST("/modifiers", exports.GetModifiers)
    partners.POST("/tills", exports.GetTills)
    partners.POST("/employees", exports.GetEmployees)
    partners.POST("/tax-definitions", exports.GetTaxDefinitions)
    // CSV → S3 variants
    partners.POST("/tills/csv", exports.ExportTillsToS3)
    partners.POST("/sales-detail/csv", exports.ExportSalesDetailToS3)
    // ... (one /csv variant per report type)
}
```

---

## 7. Handler Pattern

Handlers are thin. They parse the request, delegate to the service layer, and return JSON. No business logic lives here.

### HTTP Handler

```go
// internal/handlers/etl/api.orders.go
package etl

import (
    "net/http"
    etlorders "reports-command-service/internal/services/etl/orders"
    "github.com/gin-gonic/gin"
)

func SyncOrder(c *gin.Context) {
    refId := c.Param("refId")
    if refId == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "refId is required"})
        return
    }
    err := etlorders.ProcessOrder(refId, 0, "API for manual sync")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "order synced successfully"})
}
```

### Kafka Consumer Handler

```go
// internal/handlers/etl/consumer.orders.go
package etl

func HandleNewOrdersEvents(data []byte) error {
    var rawEvent kafka.KafkaOrderSyncMessage
    if err := json.Unmarshal(data, &rawEvent); err != nil {
        return fmt.Errorf("failed to unmarshal raw event data: %w", err)
    }
    if rawEvent.Type == kafka.KafkaEventOrderCreated {
        if rawEvent.Data.OrderRefID != "" {
            superlog.Log().Info("Handling New Order Event", map[string]interface{}{
                "orderRefID": rawEvent.Data.OrderRefID,
                "RequestId":  rawEvent.RequestId,
            })
            return etlorders.ProcessOrder(rawEvent.Data.OrderRefID, rawEvent.RetryCount, "Kafka sync for New Orders")
        }
    }
    return nil
}

func HandleUpdateEvents(data []byte) error {
    var rawEvent kafka.KafkaOrderSyncMessage
    if err := json.Unmarshal(data, &rawEvent); err != nil {
        return fmt.Errorf("failed to unmarshal raw event data: %w", err)
    }
    if rawEvent.Type == kafka.KafkaEventOrderUpdated {
        if rawEvent.Data.OrderRefID != "" {
            return etlorders.ProcessOrder(rawEvent.Data.OrderRefID, rawEvent.RetryCount, "Kafka sync for Updated Orders")
        }
    }
    return nil
}

// SQS retry handler
func HandleOrderRetryMessage(data []byte) error {
    return etlorders.HandleRetryMessage(data)
}
```

### Parallel batch handler (with semaphore)

```go
// inside api.orders.go — SyncMissingOrders
const (
    maxOrdersPerRequest = 100
    operationTimeout    = 5 * time.Minute
    maxConcurrency      = 10
)

semaphore := make(chan struct{}, maxConcurrency)
var wg sync.WaitGroup
var syncedCount, failedCount int32

for _, refId := range missingOrders {
    wg.Add(1)
    go func(orderRefId string) {
        defer wg.Done()
        select {
        case semaphore <- struct{}{}:
            defer func() { <-semaphore }()
        case <-ctx.Done():
            atomic.AddInt32(&failedCount, 1)
            return
        }
        if err := etlorders.ProcessOrder(orderRefId, 0, "API sync missing orders"); err != nil {
            atomic.AddInt32(&failedCount, 1)
        } else {
            atomic.AddInt32(&syncedCount, 1)
        }
    }(refId)
}
wg.Wait()
```

---

## 8. Service / Business Logic Pattern

Services are the only layer that may call repositories and external APIs. The canonical example is the ETL order pipeline.

### Context struct pattern

Rather than passing a dozen arguments through functions, build a "context" struct that accumulates state as the pipeline progresses:

```go
// Conceptual structure (internal/services/etl/orders/)

type SyncContext struct {
    SourceOrder   *types.SourceOrder       // raw data from MongoDB
    KDSEventLogs  *types.KDSEventLogsResponse
    Txn           *gorm.DB                 // active DB transaction
    Identifier    string                   // orderRefId
}

func NewSyncContext(identifier string) *SyncContext {
    return &SyncContext{Identifier: identifier}
}
```

### ProcessOrder — pipeline

```go
func ProcessOrder(identifier string, currentRetryCount int, syncSource string) error {
    syncContext := NewSyncContext(identifier)

    // 1. Audit trail — record that sync started
    entities.CreateOrderEventStatus(identifier, entities.SyncStarted, entities.Order, syncSource)

    // 2. Extract — fetch from MongoDB (parallel with other extractions if any)
    var extractErr error
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        extractErr = syncContext.Extract()
    }()
    wg.Wait()
    if extractErr != nil {
        entities.CreateOrderEventStatus(identifier, entities.FetchingOrderDetailsFailed, entities.Order, extractErr.Error())
        return extractErr
    }

    // 3. Validate
    if err := syncContext.Validate(); err != nil {
        entities.CreateOrderEventStatus(identifier, entities.ValidationFailed, entities.Order, err.Error())
        return err
    }

    // 4. Sort arrays (for deterministic advisory lock ordering)
    syncContext.SortArrayItems()

    // 5. Begin DB transaction
    db := database.GetPGClient()
    tx := db.Begin()
    syncContext.Txn = tx

    // 6. Acquire advisory lock (prevents concurrent processing of same order)
    repositories.AcquireAdvisoryLock(tx, identifier, map[string]interface{}{"orderRefId": identifier})

    // 7. Transform & persist
    if err := syncContext.TransformAndSave(); err != nil {
        tx.Rollback()
        entities.CreateOrderEventStatus(identifier, entities.SyncFailed, entities.Order, err.Error())
        // Send to SQS for retry if retries not exhausted
        produceRetryMessage(identifier, &currentRetryCount)
        return err
    }

    tx.Commit()

    // 8. Post-creation hooks (fire-and-forget notifications, etc.)
    HandlePostOrderCreation(identifier)

    entities.CreateOrderEventStatus(identifier, entities.SyncCompleted, entities.Order, "")
    return nil
}
```

---

## 9. Repository / Entity Pattern (GORM)

All database models (entities) live in `internal/repositories/entities/`. Each entity:

1. Defines the GORM struct with tags.
2. Implements `TableName() string` to prevent GORM's plural-snake convention from surprising you.
3. Provides persistence methods on the receiver — both with and without a transaction argument.

### Restaurants entity — full example

```go
package entities

import (
    "errors"
    database "reports-command-service/internal/config/databases"
    "time"
    "gorm.io/gorm"
)

type RestaurantType string

const (
    RestaurantTypeRestaurant RestaurantType = "Restaurant"
    RestaurantTypeFairStand  RestaurantType = "FairStand"
)

type Restaurants struct {
    ID                uint                        `gorm:"column:id;primaryKey;autoIncrement"`
    RestaurantName    string                      `gorm:"column:restaurant_name;not null;type:varchar"`
    AliasName         *string                     `gorm:"column:alias_name;type:varchar"`
    Address           string                      `gorm:"column:address;not null;type:varchar"`
    MobileNumber      string                      `gorm:"column:mobileNumber;not null;type:varchar"`
    RefID             string                      `gorm:"column:ref_id;uniqueIndex;not null;type:varchar"`
    BusinessRefID     string                      `gorm:"column:business_ref_id;not null;type:varchar"`
    BusinessID        uint                        `gorm:"column:business_id"`
    DateTimeSettingID *uint                       `gorm:"column:date_time_setting_id"`
    Business          Businesses                  `gorm:"foreignKey:BusinessID;constraint:OnDelete:CASCADE"`
    DateTimeSetting   *RestaurantDateTimeSettings `gorm:"foreignKey:DateTimeSettingID"`
    StartTime         *time.Time                  `gorm:"column:start_time"`
    EndTime           *time.Time                  `gorm:"column:end_time"`
    ListingType       RestaurantType              `gorm:"column:listing_type;not null;default:'Restaurant'"`
    CreatedAt         time.Time                   `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt         time.Time                   `gorm:"column:updated_at;autoUpdateTime"`
}

func (*Restaurants) TableName() string { return "restaurants" }

// --- No-transaction methods (use sparingly outside of ETL pipeline) ---

func (r *Restaurants) SaveRestaurant() error {
    return database.GetPGClient().Create(r).Error
}

func GetRestaurantByRefID(refID string) (*Restaurants, error) {
    var restaurant *Restaurants
    if err := database.GetPGClient().Where("ref_id = ?", refID).First(&restaurant).Error; err != nil {
        return nil, err
    }
    return restaurant, nil
}

func (r *Restaurants) CreateOrUpdateRestaurant() error {
    var existing Restaurants
    if err := database.GetPGClient().Where("ref_id = ?", r.RefID).First(&existing).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return database.GetPGClient().Create(r).Error
        }
        return err
    }
    r.ID = existing.ID
    return database.GetPGClient().Updates(r).Error
}

// --- Transaction-aware methods (used inside ETL pipeline) ---

func GetRestaurantByRefIDWithTransaction(tx *gorm.DB, refID string) (*Restaurants, error) {
    var restaurant *Restaurants
    if err := tx.Where("ref_id = ?", refID).First(&restaurant).Error; err != nil {
        return nil, err
    }
    return restaurant, nil
}

func (r *Restaurants) CreateOrUpdateRestaurantWithTransaction(tx *gorm.DB) error {
    var existing Restaurants
    err := tx.Where("ref_id = ?", r.RefID).First(&existing).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return tx.Create(r).Error
        }
        return err
    }
    r.ID = existing.ID
    return tx.Updates(r).Error
}

// CreateRestaurantWithTransaction — only inserts; if already exists sets ID and returns nil
func (r *Restaurants) CreateRestaurantWithTransaction(tx *gorm.DB) error {
    var existing Restaurants
    err := tx.Where("ref_id = ?", r.RefID).First(&existing).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return tx.Create(r).Error
        }
        return err
    }
    r.ID = existing.ID
    return nil
}
```

### Entity naming convention

| Pattern | Rule |
|---------|------|
| Struct name | PascalCase, usually plural: `OrderSales`, `Restaurants` |
| `TableName()` | Explicit lowercase snake: `"order_sales"`, `"restaurants"` |
| Primary key | `uint`, `gorm:"primaryKey;autoIncrement"` |
| Nullable fields | `*Type` (pointer) |
| Soft timestamps | `autoCreateTime`, `autoUpdateTime` tags |
| Foreign keys | `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE"` |
| Unique index | `gorm:"uniqueIndex"` on the ref_id column |

### GORM AutoMigrate

```go
// internal/repositories/automigrate.go
func AutoMigrate() {
    var models = []interface{}{
        // Uncomment to run migration:
        // &entities.Businesses{},
        // &entities.Restaurants{},
        // &entities.OrderSales{},
        // ...
    }
    db := database.GetPGClient()
    if err := db.AutoMigrate(models...); err != nil {
        panic(err)
    }
}
```

> **Convention:** In production, migrations are managed manually (or via a migration tool). AutoMigrate is kept here as a convenience for development but all models are commented out to prevent accidental schema changes.

---

## 10. Database — PostgreSQL

**File:** `internal/config/databases/postgres.go`

Two connection methods: standard password auth, and IAM (AWS RDS).

```go
package database

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/rds/auth"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "log"
    "os"
)

var PgDB *gorm.DB

// InitPGDB — standard DSN authentication
func InitPGDB() error {
    dsn := "host=" + os.Getenv("POSTGRES_HOST") +
        " user=" + os.Getenv("POSTGRES_USER") +
        " password=" + os.Getenv("POSTGRES_PASSWORD") +
        " dbname=" + os.Getenv("POSTGRES_DB") +
        " port=" + os.Getenv("POSTGRES_PORT") +
        " sslmode="  // sslmode intentionally blank (handled by network layer / VPC)
    var err error
    PgDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    return nil
}

// InitPGDBIAM — AWS RDS IAM authentication (generates a short-lived token)
func InitPGDBIAM() error {
    host := os.Getenv("POSTGRES_HOST")
    user := os.Getenv("POSTGRES_USER")
    port := 5432
    dbname := os.Getenv("POSTGRES_DB")
    awsRegion := os.Getenv("AWS_REGION")

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
    if err != nil {
        return fmt.Errorf("failed to load configuration, %w", err)
    }
    if _, err = cfg.Credentials.Retrieve(context.TODO()); err != nil {
        log.Fatalf("Unable to retrieve credentials: %v", err)
    }

    authToken, err := auth.BuildAuthToken(
        context.TODO(),
        fmt.Sprintf("%s:%d", host, port),
        awsRegion,
        user,
        cfg.Credentials,
    )
    if err != nil {
        log.Fatalf("Failed to create authentication token: %v", err)
    }

    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
        host, port, user, authToken, dbname,
    )
    PgDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    return nil
}

func GetPGClient() *gorm.DB { return PgDB }
```

### Access pattern

All entities call `database.GetPGClient()` directly. The global `PgDB` pointer is set once at startup. There is no connection pool management beyond GORM's default.

---

## 11. Database — MongoDB

**File:** `internal/config/databases/mongodb.go`

```go
package database

import (
    "context"
    "fmt"
    "os"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    Client         *mongo.Client
    MongoDB        *mongo.Database     // primary DB (MONGODB_DB)
    RewardsMongoDB *mongo.Database     // secondary DB (MONGODB_REWARDS_DB) — same client
)

func InitMongoDB() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cred := &MongoCredentials{
        Username: os.Getenv("MONGODB_USERNAME"),
        Password: os.Getenv("MONGODB_PASSWORD"),
        Host:     os.Getenv("MONGODB_URI"),
        Database: os.Getenv("MONGODB_DB"),
    }

    uri := fmt.Sprintf("mongodb+srv://%s:%s@%s", cred.Username, cred.Password, cred.Host)
    clientOptions := options.Client().ApplyURI(uri)

    var err error
    Client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        return fmt.Errorf("failed to connect to MongoDB: %w", err)
    }

    // Verify connection
    if err = Client.Ping(ctx, nil); err != nil {
        return fmt.Errorf("failed to ping MongoDB: %w", err)
    }

    MongoDB = Client.Database(cred.Database)
    RewardsMongoDB = Client.Database(os.Getenv("MONGODB_REWARDS_DB"))
    return nil
}

func GetMongoDB() *mongo.Database         { return MongoDB }
func GetRewardsMongoDB() *mongo.Database  { return RewardsMongoDB }
func GetMongoDBByName(dbName string) *mongo.Database {
    if Client == nil { return nil }
    return Client.Database(dbName)
}
```

### Accessing a collection

```go
// Example: fetch an order from MongoDB
coll := database.GetMongoDB().Collection("restaurant_orders")
filter := bson.M{"refId": orderRefId}
var result bson.M
err := coll.FindOne(ctx, filter).Decode(&result)
```

---

## 12. Raw SQL Queries Pattern

Some queries are too complex for GORM's ORM layer. These are stored as `.sql` files and loaded at runtime.

### File layout

```
internal/repositories/queries/
├── orders/
│   ├── failed_orders.sql
│   ├── order_sync_status.sql
│   ├── retry_counts.sql
│   ├── only_failed_orders.sql
│   └── success_orders.sql
├── partners/
│   ├── sales_detail.sql
│   ├── menu_items.sql
│   ├── payments.sql
│   └── ...
└── resiliency/
    ├── advisory_lock.sql
    ├── generate_lock_key_hash.sql
    └── slow_queries.sql
```

### Loading pattern (sync.Once)

```go
// internal/repositories/advisory_lock_helper.go
var (
    generateLockKeyHashQuery string
    advisoryLockQuery        string
    sqlQueriesOnce           sync.Once
    sqlQueriesLoadErr        error
)

func loadSQLQueries() error {
    sqlQueriesOnce.Do(func() {
        dir := getSQLDir() // uses runtime.Caller(0) to get absolute path
        b, err := os.ReadFile(filepath.Join(dir, "generate_lock_key_hash.sql"))
        if err != nil {
            sqlQueriesLoadErr = err
            return
        }
        generateLockKeyHashQuery = string(b)

        b, err = os.ReadFile(filepath.Join(dir, "advisory_lock.sql"))
        if err != nil {
            sqlQueriesLoadErr = err
            return
        }
        advisoryLockQuery = string(b)
    })
    return sqlQueriesLoadErr
}

// getSQLDir resolves the path relative to the source file at compile time
func getSQLDir() string {
    _, filename, _, _ := runtime.Caller(0)
    return filepath.Join(filepath.Dir(filename), "../repositories/queries/resiliency")
}
```

> **Important for Docker:** SQL files must be copied into the container image.
> See the `COPY --from=builder /app/internal/repositories/queries /app/internal/repositories/queries` line in the Dockerfile.

---

## 13. Advisory Locks

PostgreSQL advisory locks prevent two concurrent goroutines from processing the same order simultaneously — a common problem when both Kafka and the HTTP API can trigger the same order sync.

### SQL files

`resiliency/generate_lock_key_hash.sql`:
```sql
SELECT (('x'||substr(md5(@lock_identifier),1,16))::bit(64)::bigint) AS lock_key
```

`resiliency/advisory_lock.sql`:
```sql
SELECT 1 FROM pg_advisory_xact_lock(@lock_key)
```

### Go wrapper

```go
func AcquireAdvisoryLock(tx *gorm.DB, lockIdentifier string, logContext map[string]interface{}) error {
    if err := loadSQLQueries(); err != nil {
        return fmt.Errorf("failed to load SQL queries: %w", err)
    }

    // 1. Hash the string identifier to a 64-bit int (PostgreSQL requirement)
    var lockKey int64
    keyQuery := strings.Replace(generateLockKeyHashQuery, "@lock_identifier", "?", 1)
    if err := tx.Raw(keyQuery, lockIdentifier).Scan(&lockKey).Error; err != nil {
        return fmt.Errorf("failed to compute advisory lock key: %w", err)
    }

    // 2. Acquire the transaction-scoped advisory lock (blocks until acquired)
    lockQuery := strings.Replace(advisoryLockQuery, "@lock_key", "?", 1)
    var locked int
    if err := tx.Raw(lockQuery, lockKey).Scan(&locked).Error; err != nil {
        return fmt.Errorf("failed to acquire advisory lock: %w", err)
    }

    return nil
}
```

### Usage

```go
tx := database.GetPGClient().Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// Lock is released automatically when the transaction commits or rolls back
repositories.AcquireAdvisoryLock(tx, orderRefId, logCtx)

// ... all reads and writes inside this tx are now serialised per orderRefId
tx.Commit()
```

---

## 14. Kafka — Producer

**File:** `internal/config/kafka/producer.go`

```go
package kafka

import (
    "encoding/json"
    "fmt"
    "os"
    "reports-command-service/pkg/superlog"
    "strings"
    "sync"
    "time"
    "github.com/IBM/sarama"
)

type Producer struct {
    producer sarama.SyncProducer
    topic    string
}

var (
    kafkaProducerInstance *Producer
    producerOnce          sync.Once
)

// GetKafkaProducer returns a singleton SyncProducer for the given topic + auth type.
func GetKafkaProducer(topic string, authType string) (*Producer, error) {
    var producerErr error
    producerOnce.Do(func() {
        config := sarama.NewConfig()
        config.Producer.Return.Successes = true
        config.Producer.Return.Errors = true
        config.Producer.Retry.Max = 3
        config.Producer.RequiredAcks = sarama.WaitForAll // strongest durability

        switch strings.ToLower(authType) {
        case "sasl":
            producerErr = ConfigureSASLPlain(config)
        case "iam":
            producerErr = ConfigureIAM(config)
        default:
            producerErr = ConfigureSASLPlain(config) // SASL is the default
        }
        if producerErr != nil { return }

        producer, err := sarama.NewSyncProducer(GetBrokers(), config)
        if err != nil {
            producerErr = fmt.Errorf("failed to create producer: %w", err)
            return
        }
        kafkaProducerInstance = &Producer{producer: producer, topic: topic}
    })
    if producerErr != nil { return nil, producerErr }
    return kafkaProducerInstance, nil
}

// ProduceMessage sends a JSON-encoded value to the producer's topic.
func (kp *Producer) ProduceMessage(key string, value interface{}) error {
    if kp == nil {
        return fmt.Errorf("producer is nil")
    }
    valueBytes, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("error marshalling message: %w", err)
    }
    msg := &sarama.ProducerMessage{
        Topic:     kp.topic,
        Key:       sarama.StringEncoder(key),
        Value:     sarama.ByteEncoder(valueBytes),
        Timestamp: time.Now(),
    }
    _, _, err = kp.producer.SendMessage(msg)
    return err
}

// PublishKafkaMessage is the public generic helper used by application code.
func PublishKafkaMessage[T any](key string, messageType string, messageData T) error {
    authType := os.Getenv("KAFKA_AUTH_TYPE")
    topic := os.Getenv("KAFKA_NEW_ORDER_EVENT_TOPIC")

    producer, err := GetKafkaProducer(topic, authType)
    if err != nil {
        superlog.Log().Error("Error getting Kafka producer", map[string]interface{}{"error": err.Error()})
        return fmt.Errorf("failed to get producer: %w", err)
    }

    kafkaMsg := struct {
        Type string `json:"type"`
        Data T      `json:"data"`
    }{Type: messageType, Data: messageData}

    if err = producer.ProduceMessage(key, kafkaMsg); err != nil {
        superlog.Log().Error("Error producing message to Kafka", map[string]interface{}{
            "error": err.Error(), "msg": kafkaMsg,
        })
        return fmt.Errorf("error producing message to Kafka: %w on topic %s", err, topic)
    }
    return nil
}

func (kp *Producer) Close() error {
    if kp.producer != nil {
        return kp.producer.Close()
    }
    return nil
}
```

---

## 15. Kafka — Consumer

**Files:** `internal/config/kafka/consumer.go`, `kafka_config.go`

### Auth helpers

```go
// kafka_config.go

type MSKAccessTokenProvider struct {
    region string
}

func (m *MSKAccessTokenProvider) Token() (*sarama.AccessToken, error) {
    token, _, err := signer.GenerateAuthToken(context.TODO(), m.region)
    return &sarama.AccessToken{Token: token}, err
}

// ConfigureSASLPlain — username/password (default)
func ConfigureSASLPlain(config *sarama.Config) error {
    username := os.Getenv("KAFKA_SASL_USERNAME")
    password := os.Getenv("KAFKA_SASL_PASSWORD")
    if username == "" || password == "" {
        return errors.New("KAFKA_SASL_USERNAME and KAFKA_SASL_PASSWORD must be set")
    }
    config.Net.SASL.Enable = true
    config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
    config.Net.SASL.User = username
    config.Net.SASL.Password = password
    config.Net.TLS.Enable = (os.Getenv("KAFKA_AUTH_TYPE") == "IAM")
    return nil
}

// ConfigureIAM — AWS MSK IAM (short-lived OAuth tokens)
func ConfigureIAM(config *sarama.Config) error {
    config.Net.SASL.Enable = true
    config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
    config.Net.SASL.TokenProvider = &MSKAccessTokenProvider{
        region: os.Getenv("KAFKA_REGION"),
    }
    config.Net.TLS.Enable = true
    return nil
}

// GetBrokers — parses KAFKA_HOST JSON array
func GetBrokers() []string {
    kafkaHost := os.Getenv("KAFKA_HOST")
    if kafkaHost == "" {
        kafkaHost = `["13.200.177.157:9094"]` // fallback for local dev
    }
    var brokers []string
    if err := json.Unmarshal([]byte(kafkaHost), &brokers); err != nil {
        panic("Invalid KAFKA_HOST format")
    }
    return brokers
}
```

### Consumer group service

```go
// consumer.go

type EventHandler func([]byte) error

type Service struct {
    consumer sarama.ConsumerGroup
    handlers map[string]EventHandler
    topics   []string
    wg       sync.WaitGroup
    ready    chan bool
    ctx      context.Context
    cancel   context.CancelFunc
}

// Consumer implements sarama.ConsumerGroupHandler
type Consumer struct {
    ready    chan bool
    handlers map[string]EventHandler
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
    close(c.ready) // signals that the consumer is ready
    return nil
}
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for {
        select {
        case message, ok := <-claim.Messages():
            if !ok { return nil }
            if handler, exists := c.handlers[message.Topic]; exists {
                if err := handler(message.Value); err != nil {
                    log.Printf("Error processing message from topic %s: %v", message.Topic, err)
                }
            }
            session.MarkMessage(message, "") // commit offset
        case <-session.Context().Done():
            return nil
        }
    }
}

var (
    instance *Service
    once     sync.Once
)

func GetConsumerInstance() *Service {
    once.Do(func() {
        ctx, cancel := context.WithCancel(context.Background())
        instance = &Service{
            handlers: make(map[string]EventHandler),
            topics:   []string{},
            ready:    make(chan bool),
            ctx:      ctx,
            cancel:   cancel,
        }
    })
    return instance
}

func (ks *Service) AddTopic(topic string)                              { ks.topics = append(ks.topics, topic) }
func (ks *Service) RegisterHandler(topic string, handler EventHandler) { ks.handlers[topic] = handler }

func (ks *Service) InitializeWithAuth(groupID string, authType string) error {
    if len(ks.topics) == 0 {
        return errors.New("no topics added")
    }
    config := sarama.NewConfig()
    config.Version = sarama.V2_0_0_0
    config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
    config.Consumer.Offsets.Initial = sarama.OffsetNewest // start from latest on new group

    switch strings.ToLower(authType) {
    case "sasl":
        if err := ConfigureSASLPlain(config); err != nil { return err }
    case "iam":
        if err := ConfigureIAM(config); err != nil { return err }
    default:
        if err := ConfigureSASLPlain(config); err != nil { return err }
    }

    consumer, err := sarama.NewConsumerGroup(GetBrokers(), groupID, config)
    if err != nil {
        return fmt.Errorf("error creating consumer group: %w", err)
    }
    ks.consumer = consumer
    return nil
}

func (ks *Service) Start(ctx context.Context) {
    ks.wg.Add(1)
    go func() {
        defer ks.wg.Done()
        for {
            consumer := &Consumer{ready: make(chan bool), handlers: ks.handlers}
            if err := ks.consumer.Consume(ctx, ks.topics, consumer); err != nil {
                log.Printf("Error from consumer: %v", err)
            }
            if ctx.Err() != nil { return }
            consumer.ready = make(chan bool) // reset for next rebalance
        }
    }()
    <-ks.ready // block until first rebalance complete
    log.Println("Kafka consumer has started")
}

func (ks *Service) Shutdown(ctx context.Context) error {
    ks.cancel()
    if err := ks.consumer.Close(); err != nil {
        return fmt.Errorf("error closing consumer: %w", err)
    }
    done := make(chan struct{})
    go func() { ks.wg.Wait(); close(done) }()
    select {
    case <-ctx.Done(): return ctx.Err()
    case <-done:       return nil
    }
}
```

---

## 16. SQS Consumer & Producer

**File:** `internal/config/sqs/sqs.go`

SQS is used as a **retry queue**. When an order sync fails, the message is sent to SQS with a delay. The SQS consumer processes it later.

```go
package sqs

import (
    "encoding/json"
    "fmt"
    "os"
    "reports-command-service/pkg/superlog"
    "sync"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)

type MessageHandler func(message []byte) error

type Service struct {
    svc      *sqs.SQS
    queueURL string
    handlers map[string]MessageHandler
}

var (
    sqsServiceInstance *Service
    once               sync.Once
)

func GetSQSService() *Service {
    once.Do(func() {
        sess, err := session.NewSession(&aws.Config{
            Region: aws.String(os.Getenv("SQS_REGION")),
        })
        if err != nil {
            fmt.Printf("Error creating AWS session: %v\n", err)
            return
        }
        sqsServiceInstance = &Service{
            svc:      sqs.New(sess),
            queueURL: os.Getenv("SQS_RETRY_QUEUE_URL"),
            handlers: make(map[string]MessageHandler),
        }
    })
    return sqsServiceInstance
}

func (s *Service) RegisterHandler(messageType string, handler MessageHandler) {
    s.handlers[messageType] = handler
}

// SendMessage enqueues a message with a delay (for retry back-off).
func (s *Service) SendMessage(messageType string, message interface{}, delaySeconds int64) error {
    body, err := json.Marshal(struct {
        Type    string      `json:"type"`
        Payload interface{} `json:"payload"`
    }{Type: messageType, Payload: message})
    if err != nil {
        return fmt.Errorf("error marshalling message: %v", err)
    }
    _, err = s.svc.SendMessage(&sqs.SendMessageInput{
        DelaySeconds: aws.Int64(delaySeconds),
        MessageBody:  aws.String(string(body)),
        QueueUrl:     &s.queueURL,
    })
    return err
}

// StartProcessing is a blocking long-poll loop.
func (s *Service) StartProcessing() {
    for {
        message, err := s.ReceiveMessage()
        if err != nil { fmt.Printf("Error receiving: %v\n", err); continue }
        if message == nil { continue }

        var wrapper struct {
            Type    string          `json:"type"`
            Payload json.RawMessage `json:"payload"`
        }
        if err = json.Unmarshal([]byte(*message.Body), &wrapper); err != nil {
            superlog.Log().Error("Error unmarshalling SQS message", map[string]interface{}{"error": err})
            continue
        }

        handler, ok := s.handlers[wrapper.Type]
        if !ok {
            superlog.Log().Error("No handler for SQS message type", map[string]interface{}{"type": wrapper.Type})
            continue
        }

        if err = handler(wrapper.Payload); err != nil {
            superlog.Log().Error("Handler failed", map[string]interface{}{"error": err, "type": wrapper.Type})
            continue
        }

        // Only delete on success
        s.DeleteMessage(message.ReceiptHandle)
    }
}

func (s *Service) ReceiveMessage() (*sqs.Message, error) {
    result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:            &s.queueURL,
        MaxNumberOfMessages: aws.Int64(1),
        WaitTimeSeconds:     aws.Int64(20), // long polling
    })
    if err != nil { return nil, err }
    if len(result.Messages) > 0 { return result.Messages[0], nil }
    return nil, nil
}

func (s *Service) DeleteMessage(receiptHandle *string) error {
    _, err := s.svc.DeleteMessage(&sqs.DeleteMessageInput{
        QueueUrl:      &s.queueURL,
        ReceiptHandle: receiptHandle,
    })
    return err
}
```

**SQS event type constants:**
```go
// internal/config/sqs/types.go
type ConsumerEventTypes string

const (
    ConsumerEventOrderSyncRetry ConsumerEventTypes = "ORDER_SYNC_RETRY"
)
```

---

## 17. Consumer Wiring (consumers.go)

**File:** `internal/consumers.go`

This is the single place where all Kafka topics are registered with their handlers before the consumer is started.

```go
package app

import (
    "context"
    "os"
    "reports-command-service/internal/config/kafka"
    "reports-command-service/internal/config/sqs"
    "reports-command-service/internal/handlers/etl"
    "reports-command-service/pkg/superlog"
    "time"
)

func InitKafkaConsumer() {
    kafkaService := kafka.GetConsumerInstance()

    // Register every topic → handler pair
    topics := []struct {
        envKey  string
        handler kafka.EventHandler
    }{
        {"KAFKA_NEW_ORDER_EVENT_TOPIC",       etl.HandleNewOrdersEvents},
        {"KAFKA_GIFT_CARD_PURCHASE_UPDATE",   etl.HandleGiftCardEvents},
        {"KAFKA_UPDATES_EVENTS",              etl.HandleUpdateEvents},
        {"KAFKA_TEST_TOPIC",                  etl.HandleTestEvents},
        {"KAFKA_RESTAURANT_SETTINGS_UPDATES", etl.HandleRestaurantSettingsUpdates},
        {"KAFKA_CUSTOMER_UPDATES",            etl.HandleCustomerDetailsUpdates},
        {"KAFKA_TXN_UPDATE_EVENTS",           etl.HandleOrderTransactionUpdates},
        {"KAFKA_VOICE_ORDERING_EVENTS_TOPIC", etl.HandleVoiceOrderingEvents},
        {"KAFKA_VISION_AI_EVENTS_TOPIC",      etl.HandleVisionAICameraEvents},
    }
    for _, t := range topics {
        topic := os.Getenv(t.envKey)
        kafkaService.AddTopic(topic)
        go kafkaService.RegisterHandler(topic, t.handler)
    }

    // Initialise the consumer group connection
    if err := kafkaService.InitializeWithAuth(
        os.Getenv("CONSUMER_GROUP"),
        os.Getenv("KAFKA_AUTH_TYPE"),
    ); err != nil {
        panic(err)
    }
    superlog.Log().Info("Kafka Consumer Initialized", nil)

    // Start consuming (blocks until ready, then returns when ctx cancelled)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    kafkaService.Start(ctx)

    // Graceful shutdown with 30-second timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    if err := kafkaService.Shutdown(shutdownCtx); err != nil {
        panic(err)
    }
}

func InitSQSConsumer() {
    sqsService := sqs.GetSQSService()
    sqsService.RegisterHandler(string(sqs.ConsumerEventOrderSyncRetry), etl.HandleOrderRetryMessage)
    go sqsService.StartProcessing()
}
```

---

## 18. Kafka & SQS Message Types

**File:** `internal/config/kafka/types.go`

```go
package kafka

// Event type constants
const (
    KafkaEventOrderCreated        = "ORDER_CREATED"
    KafkaEventOrderUpdated        = "ORDER_UPDATED"
    KafkaReportOrderCreated       = "REPORT_ORDER_CREATED"
    KafkaRestaurantTimingsUpdated = "DATE_TIME_UPDATED"
    KafkaEventVisionAICameraData  = "VISION_AI_CAMERA_DATA"
)

// --- Inbound message structs ---

type OrderData struct {
    OrderRefID string `json:"orderRefId"`
}

type KafkaOrderSyncMessage struct {
    Type       string    `json:"type"`
    Data       OrderData `json:"data"`
    RequestId  string    `json:"requestId"`
    RetryCount int       `json:"retryCount"`
}

type RestaurantData struct {
    RestaurantRefID string `json:"restaurantRefId"`
}

type KafkaRestaurantTimeSettingsSyncMessage struct {
    Type string         `json:"type"`
    Data RestaurantData `json:"data"`
}

type CustomerData struct {
    UserRefID string `json:"userRefId"`
}

type KafkaCustomerSyncMessage struct {
    Type string       `json:"type"`
    Data CustomerData `json:"data"`
}

type GiftCardData struct {
    RefID string `json:"refId"`
}

type KafkaGiftCardSyncMessage struct {
    Type      string       `json:"type"`
    Data      GiftCardData `json:"data"`
    RequestId string       `json:"requestId"`
}

// VisionAI
type VisionAIMinimalData struct {
    ObjectID  string `json:"object_id"`
    EventType string `json:"event_type"`
    CameraID  string `json:"camera_id"`
}

type KafkaVisionAIMessage struct {
    Type      string              `json:"type"`
    Data      VisionAIMinimalData `json:"data"`
    RequestId string              `json:"requestId"`
}

// --- Retry message (deserialized from SQS payload) ---

type RetryMessage struct {
    Id   string `json:"id"`
    Body struct {
        Type      string `json:"type"`
        RequestId string `json:"requestId"`
        Data      struct {
            OrderRefId string `json:"orderRefId"`
        } `json:"data"`
        RetryCount int `json:"retryCount"`
    } `json:"body"`
}

// --- Outbound message (used by PublishKafkaMessage) ---

type KafkaReportOrderCreatedMessage struct {
    Type      string    `json:"type"`
    RequestId string    `json:"requestId"`
    Data      OrderData `json:"data"`
}
```

---

## 19. Event Audit Trail (EventStatuses)

Every order sync transition is recorded to the `event_statuses` table. This enables the `/internal/orders/sync-status/:refId` endpoint.

```go
// internal/repositories/entities/EventStatuses.go

type SyncStatus string
type EventType string

const (
    SyncStarted                SyncStatus = "SYNC_STARTED"
    FetchingOrderDetails       SyncStatus = "FETCHING_ORDER_DETAILS"
    FetchingOrderDetailsFailed SyncStatus = "FETCHING_ORDER_DETAILS_FAILED"
    OrderDetailsFetched        SyncStatus = "ORDER_DETAILS_FETCHED"
    SyncInProgress             SyncStatus = "SYNC_IN_PROGRESS"
    SyncCompleted              SyncStatus = "SYNC_COMPLETED"
    SyncFailed                 SyncStatus = "SYNC_FAILED"
    MaxRetriesReached          SyncStatus = "MAX_RETRIES_REACHED"
    SentForRetry               SyncStatus = "SENT_FOR_RETRY"
    ValidationFailed           SyncStatus = "VALIDATION_FAILED"
)

const (
    Order      EventType = "ORDER"
    OrderRetry EventType = "ORDER_RETRY"
    Customer   EventType = "CUSTOMER"
)

type EventStatuses struct {
    ID               uint       `gorm:"column:id;primaryKey;autoIncrement"`
    EventIdentifier  string     `gorm:"column:event_identifier;not null"`
    EventStatus      SyncStatus `gorm:"column:event_status;not null"`
    EventType        EventType  `gorm:"column:event_type;not null"`
    EventDescription string     `gorm:"column:event_description;not null"`
    Extras           string     `gorm:"column:extras;type:text"`
    CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (*EventStatuses) TableName() string { return "event_statuses" }

// CreateOrderEventStatus is the public helper called throughout the service layer.
func CreateOrderEventStatus(identifier string, status SyncStatus, eventType EventType, desc string) error {
    return database.GetPGClient().Create(&EventStatuses{
        EventIdentifier:  identifier,
        EventStatus:      status,
        EventType:        eventType,
        EventDescription: desc,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }).Error
}
```

---

## 20. Docker

**File:** `Dockerfile`

Multi-stage build: minimal final image, non-root user.

```dockerfile
# ---- Builder stage ----
FROM golang:1.24.7-alpine AS builder
WORKDIR /app

# Download dependencies separately for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build from cmd/ directory, statically linked for Alpine
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# ---- Runtime stage ----
FROM alpine:latest

# Create non-root user (UID 5555)
RUN addgroup -g 5555 plateron && adduser -D -u 5555 -G plateron plateron

WORKDIR /app

# Copy binary
COPY --from=builder /app/main .

# Copy SQL query files (loaded at runtime by advisory_lock_helper.go etc.)
COPY --from=builder /app/internal/repositories/queries /app/internal/repositories/queries

# Copy .env.* files (for local/test environments)
COPY .env.* .

EXPOSE 8080

# Set ownership then drop privileges
RUN chown -R plateron:plateron /app
USER plateron

CMD ["./main"]
```

### Key design decisions

| Decision | Reason |
|----------|--------|
| `CGO_ENABLED=0` | Pure-Go binary, compatible with Alpine's musl libc |
| `-a -installsuffix cgo` | Forces rebuild of all packages to ensure CGO is truly disabled |
| Separate `go mod download` layer | Docker cache hit on dependency layer if only source changed |
| Non-root user `plateron` (UID 5555) | Security: container cannot write to host even if escaped |
| SQL files copied explicitly | They are read from disk at runtime via `os.ReadFile` |
| `.env.*` copied | Allows local/test env vars to be loaded if needed |

---

## 21. VSCode launch.json

**File:** `.vscode/launch.json`

Provides one-click debug configurations for QA environments directly from VSCode.

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "QA02 Environment",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/main.go",
      "env": {
        "NODE_ENV": "qa02",
        "AWS_REGION": "us-east-1",
        "LOAD_CREDS_FROM_FILE": "false"
      },
      "cwd": "${workspaceFolder}",
      "args": [],
      "console": "integratedTerminal"
    },
    {
      "name": "QA03 Environment",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/main.go",
      "env": {
        "NODE_ENV": "qa03",
        "AWS_REGION": "us-east-1",
        "LOAD_CREDS_FROM_FILE": "false"
      },
      "cwd": "${workspaceFolder}",
      "args": [],
      "console": "integratedTerminal"
    }
  ]
}
```

### How it works

- `NODE_ENV=qa02` tells `LoadEnvironment()` to call the config service, which fetches the QA02 config blob and hydrates all other env vars.
- `AWS_REGION=us-east-1` is required even before config is loaded (validated in `LoadEnvironment`).
- `LOAD_CREDS_FROM_FILE=false` tells the AWS SDK to use the default credential chain (EC2 instance profile, `~/.aws/credentials`, env vars, etc.) rather than a local file.
- `cwd: "${workspaceFolder}"` ensures relative file paths (e.g., SQL files, `app.log`) resolve correctly.
- `console: "integratedTerminal"` keeps stdin/stdout visible in the VSCode terminal panel.

### Adding a new environment

```json
{
  "name": "Staging Environment",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/cmd/main.go",
  "env": {
    "NODE_ENV": "staging",
    "AWS_REGION": "ap-south-1",
    "LOAD_CREDS_FROM_FILE": "false"
  },
  "cwd": "${workspaceFolder}",
  "args": [],
  "console": "integratedTerminal"
}
```

---

## 22. Environment Variables Reference

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `NODE_ENV` | YES | — | Controls config loading: `test`/`local` skip config service |
| `AWS_REGION` | YES | — | Validated at startup; needed for IAM auth (RDS, MSK, SQS) |
| `CONFIG_PATH` | In non-local | — | Base path for config service (e.g. `/v1`) |
| `LOAD_CREDS_FROM_FILE` | NO | `false` | If `true` loads AWS creds from file instead of default chain |
| **PostgreSQL** | | | |
| `POSTGRES_HOST` | YES | — | DB hostname |
| `POSTGRES_USER` | YES | — | DB username |
| `POSTGRES_PASSWORD` | YES | — | DB password (empty for IAM) |
| `POSTGRES_DB` | YES | — | Database name |
| `POSTGRES_PORT` | YES | — | Usually `5432` |
| **MongoDB** | | | |
| `MONGODB_URI` | YES | — | Hostname part of connection string |
| `MONGODB_USERNAME` | YES | — | Atlas username |
| `MONGODB_PASSWORD` | YES | — | Atlas password |
| `MONGODB_DB` | YES | — | Primary database name |
| `MONGODB_REWARDS_DB` | YES | — | Rewards database name (same cluster) |
| **Kafka** | | | |
| `KAFKA_HOST` | YES | `["13.200.177.157:9094"]` | JSON array of broker addresses |
| `KAFKA_AUTH_TYPE` | NO | `sasl` | `sasl` or `iam` |
| `KAFKA_SASL_USERNAME` | SASL only | — | SASL plain username |
| `KAFKA_SASL_PASSWORD` | SASL only | — | SASL plain password |
| `KAFKA_REGION` | IAM only | — | AWS region for MSK IAM |
| `CONSUMER_GROUP` | YES | — | Consumer group ID |
| `KAFKA_NEW_ORDER_EVENT_TOPIC` | YES | — | New order events |
| `KAFKA_UPDATES_EVENTS` | YES | — | Order update events |
| `KAFKA_GIFT_CARD_PURCHASE_UPDATE` | YES | — | Gift card events |
| `KAFKA_TEST_TOPIC` | YES | — | Test topic |
| `KAFKA_RESTAURANT_SETTINGS_UPDATES` | YES | — | Restaurant settings events |
| `KAFKA_CUSTOMER_UPDATES` | YES | — | Customer update events |
| `KAFKA_TXN_UPDATE_EVENTS` | YES | — | Transaction update events |
| `KAFKA_VOICE_ORDERING_EVENTS_TOPIC` | YES | — | Voice ordering events |
| `KAFKA_VISION_AI_EVENTS_TOPIC` | YES | — | Vision AI camera events |
| **SQS** | | | |
| `SQS_REGION` | YES | — | AWS region for SQS |
| `SQS_RETRY_QUEUE_URL` | YES | — | Full URL of the retry queue |
| **Logging** | | | |
| `REMOTE_LOG` | NO | `false` | `true` → HTTPShipper; `false` → FileShipper |
| `LOGGER_HOST` | If REMOTE_LOG=true | — | HTTP endpoint for remote log shipping |
| **Firebase** | | | |
| `FIREBASE_DB_URL` | YES | — | Realtime database URL |
| `FIREBASE_PROJECT_ID` | YES | — | GCP project ID |
| `FIREBASE_PRIVATE_KEY` | YES | — | PEM private key (multi-line, sanitized at load time) |

---

## 23. Key Dependencies (go.mod)

```
module reports-command-service

go 1.23.0
toolchain go1.24.3
```

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/gin-gonic/gin` | v1.10.0 | HTTP framework |
| `github.com/gin-contrib/cors` | v1.7.2 | CORS middleware |
| `gorm.io/gorm` | v1.25.12 | ORM |
| `gorm.io/driver/postgres` | v1.5.9 | GORM PostgreSQL driver |
| `go.mongodb.org/mongo-driver` | v1.17.1 | MongoDB driver |
| `github.com/IBM/sarama` | v1.43.3 | Kafka client (producer + consumer group) |
| `github.com/aws/aws-msk-iam-sasl-signer-go` | v1.0.0 | MSK IAM token generator |
| `github.com/aws/aws-sdk-go` | v1.55.5 | AWS SDK v1 (SQS) |
| `github.com/aws/aws-sdk-go-v2/config` | v1.28.5 | AWS SDK v2 config |
| `github.com/aws/aws-sdk-go-v2/feature/rds/auth` | v1.4.24 | RDS IAM auth tokens |
| `github.com/go-resty/resty/v2` | v2.15.3 | HTTP client (config service calls) |
| `firebase.google.com/go/v4` | v4.18.0 | Firebase SDK |
| `github.com/google/uuid` | v1.6.0 | UUID generation |
| `github.com/joho/godotenv` | v1.5.1 | `.env` file loader |
| `github.com/pkg/sftp` | v1.13.10 | SFTP client |
| `github.com/swaggo/gin-swagger` | v1.6.0 | Swagger UI endpoint |
| `github.com/swaggo/swag` | v1.16.3 | Swagger annotation scanner |
| `github.com/stretchr/testify` | v1.10.0 | Test assertions |

---

## 24. Concurrency Patterns

### Pattern 1 — Singleton with sync.Once

Used for: Logger, Kafka Producer, Kafka Consumer, SQS Service.

```go
var (
    instance *Service
    once     sync.Once
)

func GetService() *Service {
    once.Do(func() {
        instance = &Service{ /* init */ }
    })
    return instance
}
```

### Pattern 2 — Parallel extraction with WaitGroup

Used for: fetching source data from multiple stores concurrently.

```go
var (
    extractErr error
    wg         sync.WaitGroup
)
wg.Add(1)
go func() {
    defer wg.Done()
    extractErr = syncContext.Extract()
}()
wg.Wait()
if extractErr != nil { return extractErr }
```

### Pattern 3 — Bounded parallelism with semaphore channel

Used for: `SyncMissingOrders` — process up to N orders concurrently.

```go
semaphore := make(chan struct{}, maxConcurrency) // e.g. 10
var wg sync.WaitGroup
var count int32

for _, id := range ids {
    wg.Add(1)
    go func(id string) {
        defer wg.Done()
        semaphore <- struct{}{}        // acquire
        defer func() { <-semaphore }() // release
        if err := process(id); err != nil {
            atomic.AddInt32(&count, 1)
        }
    }(id)
}
wg.Wait()
```

### Pattern 4 — Context timeout for long-running operations

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:    // completed within timeout
case <-ctx.Done(): // timed out
    log.Warn("operation timed out")
}
```

### Pattern 5 — Fire-and-forget goroutine (logger shippers)

```go
for _, shipper := range l.shippers {
    go func() {
        shipper.Ship(event) // non-blocking, errors printed to stderr
    }()
}
```

### Pattern 6 — Advisory lock for DB-level serialisation

```go
tx := db.Begin()
repositories.AcquireAdvisoryLock(tx, orderRefId, logCtx)
// all work inside tx is now serialised per orderRefId across all pods
tx.Commit()
```

---

## 25. Error Handling Patterns

### Wrap errors with context

```go
// service layer
if err := someOp(); err != nil {
    return fmt.Errorf("context description: %w", err)
}
```

### HTTP handler errors

```go
func SomeHandler(c *gin.Context) {
    if err := service.DoWork(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
```

### Transaction rollback with deferred recover

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        superlog.Log().Error("Panic during transaction", map[string]interface{}{"panic": r})
    }
}()

if err := doWork(tx); err != nil {
    tx.Rollback()
    superlog.Log().Error("Transaction failed", map[string]interface{}{"error": err.Error()})
    return err
}
tx.Commit()
```

### Retry via SQS

When an order sync fails and retries are not exhausted:

```go
const maxRetries = 3

func produceRetryMessage(identifier string, currentRetryCount *int) error {
    if *currentRetryCount >= maxRetries {
        entities.CreateOrderEventStatus(identifier, entities.MaxRetriesReached, entities.Order, "")
        return fmt.Errorf("max retries reached for order: %s", identifier)
    }
    *currentRetryCount++
    sqsService := sqs.GetSQSService()
    return sqsService.SendMessage(
        string(sqs.ConsumerEventOrderSyncRetry),
        map[string]interface{}{"orderRefId": identifier, "retryCount": *currentRetryCount},
        30, // delaySeconds
    )
}
```

### Log and continue (consumer errors)

In Kafka/SQS consumers, errors are logged and the message is skipped (not re-queued automatically) to avoid blocking the consumer group:

```go
if err := handler(message.Value); err != nil {
    log.Printf("Error processing message from topic %s: %v", message.Topic, err)
    // MarkMessage is still called — offset advances
}
session.MarkMessage(message, "")
```

---

## 26. Full Interfaces Reference

```go
// --- Logger ---
type ILogger interface {
    Debug(message string, data map[string]interface{}) *LogEvent
    Info(message string, data map[string]interface{}) *LogEvent
    Warn(message string, data map[string]interface{}) *LogEvent
    Error(message string, data map[string]interface{}) *LogEvent
}

type LogShipper interface {
    Ship(event LogEvent) error
}

// --- HTTP Server ---
type Server interface {
    addRoutes() Server
    addMiddleware() Server
    StartServer()
}

// --- Kafka ---
// Handler function type (not interface, but the canonical function signature)
type EventHandler func([]byte) error

// sarama.ConsumerGroupHandler (implemented by internal Consumer struct)
// Setup(sarama.ConsumerGroupSession) error
// Cleanup(sarama.ConsumerGroupSession) error
// ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error

// --- SQS ---
type MessageHandler func(message []byte) error

// --- AWS MSK IAM token provider ---
// sarama.AccessTokenProvider (implemented by MSKAccessTokenProvider)
// Token() (*sarama.AccessToken, error)
```

---

## 27. RPC Clients (Internal Service-to-Service Calls)

**Directory:** `internal/config/rpc/`

Every internal microservice that this app calls has its own resty client factory. Each returns a fresh `*resty.Client` with the base URL preconfigured from env vars.

```go
// internal/config/rpc/order_source.go
package rpc

import (
    "github.com/go-resty/resty/v2"
    "os"
)

func NewOrderSource() *resty.Client {
    client := resty.New()
    client.SetBaseURL(os.Getenv("UNIFIED_SERVICE_URL"))
    client.SetHeader("Accept", "application/json")
    return client
}
```

```go
// internal/config/rpc/customer_service.go
func CustomerRPCService() *resty.Client {
    client := resty.New()
    client.SetBaseURL(os.Getenv("CUSTOMER_SERVICE_URL"))
    client.SetHeader("Accept", "application/json")
    return client
}
```

```go
// internal/config/rpc/kds_service.go
func NewKDSService() *resty.Client {
    client := resty.New()
    client.SetBaseURL(os.Getenv("KDS_SERVICE_URL"))
    client.SetHeader("Accept", "application/json")
    return client
}
```

```go
// internal/config/rpc/rewards_service.go
func RewardsRPCService() *resty.Client {
    baseURL := os.Getenv("REWARDS_SERVICE_URL")
    if baseURL == "" {
        panic("REWARDS_SERVICE_URL environment variable is not set")
    }
    client := resty.New()
    client.SetBaseURL(baseURL)
    client.SetHeader("Accept", "application/json")
    return client
}
```

**Additional env vars needed:**

| Variable | Purpose |
|----------|---------|
| `UNIFIED_SERVICE_URL` | Base URL for the unified/order-source service |
| `CUSTOMER_SERVICE_URL` | Base URL for the customer service |
| `KDS_SERVICE_URL` | Base URL for the KDS (Kitchen Display System) service |
| `REWARDS_SERVICE_URL` | Base URL for the rewards service (panics if unset) |

---

## 28. S3 Client & Uploader

**Directory:** `internal/config/s3/`

### S3 Config & Client Singleton

```go
// internal/config/s3/config.go
package s3

import (
    "fmt"
    "os"
    "sync"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

var (
    s3Client *s3.S3
    once     sync.Once
    initErr  error
)

type S3Config struct {
    BucketName string
    Region     string
    BasePath   string
}

// Singleton pattern — same as Logger, Kafka, SQS
func GetS3Client() (*s3.S3, error) {
    once.Do(func() {
        region := os.Getenv("AWS_REGION")
        if region == "" {
            initErr = fmt.Errorf("AWS_REGION environment variable is not set")
            return
        }
        sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
        if err != nil {
            initErr = fmt.Errorf("failed to create AWS session: %w", err)
            return
        }
        s3Client = s3.New(sess)
    })
    if initErr != nil { return nil, initErr }
    return s3Client, nil
}

func GetS3Config() (*S3Config, error) {
    bucketName := os.Getenv("S3_BUCKET_NAME")
    if bucketName == "" {
        return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
    }
    region := os.Getenv("AWS_REGION")
    if region == "" {
        return nil, fmt.Errorf("AWS_REGION environment variable is not set")
    }
    basePath := os.Getenv("S3_BASE_PATH")
    if basePath == "" {
        basePath = "reports/partner-exports" // default
    }
    return &S3Config{BucketName: bucketName, Region: region, BasePath: basePath}, nil
}
```

### S3 Uploader

```go
// internal/config/s3/uploader.go

type UploadOptions struct {
    BucketName           string
    Key                  string
    ContentType          string
    Metadata             map[string]*string
    ACL                  string // "private", "public-read"
    StorageClass         string // "STANDARD", "STANDARD_IA"
    ServerSideEncryption string // "AES256"
}

type UploadResult struct {
    BucketName string
    Key        string
    Location   string     // https://{bucket}.s3.amazonaws.com/{key}
    VersionID  *string
    UploadedAt time.Time
}

// UploadFile — generic S3 upload with sensible defaults
func UploadFile(data []byte, options UploadOptions) (*UploadResult, error) {
    client, err := GetS3Client()
    if err != nil { return nil, err }

    // Defaults
    if options.ContentType == "" { options.ContentType = "application/octet-stream" }
    if options.ACL == "" { options.ACL = "private" }
    if options.StorageClass == "" { options.StorageClass = "STANDARD" }
    if options.ServerSideEncryption == "" { options.ServerSideEncryption = "AES256" }

    input := &s3.PutObjectInput{
        Bucket:               aws.String(options.BucketName),
        Key:                  aws.String(options.Key),
        Body:                 bytes.NewReader(data),
        ContentType:          aws.String(options.ContentType),
        ACL:                  aws.String(options.ACL),
        StorageClass:         aws.String(options.StorageClass),
        ServerSideEncryption: aws.String(options.ServerSideEncryption),
        Metadata:             options.Metadata,
    }

    output, err := client.PutObject(input)
    if err != nil { return nil, fmt.Errorf("failed to upload to S3: %w", err) }

    return &UploadResult{
        BucketName: options.BucketName,
        Key:        options.Key,
        Location:   fmt.Sprintf("https://%s.s3.amazonaws.com/%s", options.BucketName, options.Key),
        VersionID:  output.VersionId,
        UploadedAt: time.Now().UTC(),
    }, nil
}

// UploadCSV — convenience wrapper with text/csv content type
func UploadCSV(data []byte, bucketName, basePath, key string, metadata map[string]*string) (*UploadResult, error) {
    return UploadFile(data, UploadOptions{
        BucketName:  bucketName,
        Key:         key,
        ContentType: "text/csv",
        Metadata:    metadata,
        ACL:         "private",
    })
}

// GenerateS3Key — R365-compliant timestamped key
// Format: {basePath}/{exportType}/{filename}_{locationIdentifier}_{YYYYMMDD}.csv
func GenerateS3Key(basePath, restaurantRefId, exportType, filename string) string {
    date := time.Now().UTC().Format("20060102")
    fileName := fmt.Sprintf("%s_%s_%s.csv", filename, restaurantRefId, date)
    return filepath.Join(basePath, exportType, fileName)
}

// DeleteFile, FileExists also available (HEAD / DELETE operations)
```

**Additional env vars:**

| Variable | Default | Purpose |
|----------|---------|---------|
| `S3_BUCKET_NAME` | — | Target S3 bucket for partner exports |
| `S3_BASE_PATH` | `reports/partner-exports` | Key prefix for uploaded files |

---

## 29. FTP/SFTP Client (R365 Partner Exports)

**Directory:** `internal/config/ftp/`

An alternative upload target for partner exports (Restaurant365 uses SFTP). Follows the same singleton pattern.

### Config & Connection

```go
// internal/config/ftp/config.go

type Config struct {
    Host              string
    Port              int
    Username          string
    Password          string
    Protocol          string        // only "sftp" currently supported
    Timeout           time.Duration
    MaxRetries        int
    RetryDelay        time.Duration
    KeepAliveInterval time.Duration
}

type Client struct {
    config     *Config
    sshClient  *ssh.Client
    sftpClient *sftp.Client
    mu         sync.Mutex
}

// Singleton
var (
    instance *Client
    once     sync.Once
)

func LoadConfig() (*Config, error) {
    // Reads from env: R365_FTP_HOST, R365_FTP_PORT (default 22),
    // R365_FTP_USERNAME, R365_FTP_PASSWORD, R365_FTP_PROTOCOL (default "sftp"),
    // R365_FTP_TIMEOUT (default 30s), R365_FTP_MAX_RETRIES (default 3),
    // R365_FTP_RETRY_DELAY (default 5s), R365_FTP_KEEPALIVE_INTERVAL (default 15s)
    // ...
}

func GetClient() (*Client, error) {
    once.Do(func() {
        config, _ := LoadConfig()
        instance = &Client{config: config}
        instance.Connect()
    })
    return instance, nil
}
```

**Key features:**
- SSH-based SFTP connection with password auth
- **Keep-alive goroutine**: sends `keepalive@openssh.com` requests at intervals
- **Auto-reconnect**: on keep-alive failure, retries `MaxRetries` times with `RetryDelay`
- Thread-safe via `sync.Mutex`

### Uploader

```go
// internal/config/ftp/uploader.go

type FileUploader struct {
    client *Client
}

func NewFileUploader() (*FileUploader, error) {
    client, err := GetClient()
    if err != nil { return nil, err }
    return &FileUploader{client: client}, nil
}

// UploadR365File uploads with R365 naming: {fileType}_{locationId}_{YYYYMMDD}.ext
func (u *FileUploader) UploadR365File(localFilePath, fileType, locationIdentifier string, businessDate time.Time) error {
    ext := filepath.Ext(localFilePath)
    dateStr := businessDate.Format("20060102")
    remoteFileName := fmt.Sprintf("%s_%s_%s%s", fileType, locationIdentifier, dateStr, ext)
    remotePath := os.Getenv("R365_FTP_REMOTE_PATH") // default "/"
    return u.UploadFile(UploadFileOptions{
        LocalFilePath:  localFilePath,
        RemotePath:     remotePath,
        RemoteFileName: remoteFileName,
        Overwrite:      false,
    })
}

// Named convenience methods for each R365 file type
func (u *FileUploader) UploadSalesDetail(path, loc string, date time.Time) error { ... }
func (u *FileUploader) UploadMenuItems(path, loc string, date time.Time) error   { ... }
func (u *FileUploader) UploadPayments(path, loc string, date time.Time) error    { ... }
// ... (Deposits, Discounts, Employees, LaborDetail, TaxDetails, Voids)

// UploadBytes / UploadCSVBytes — for in-memory data (no temp file)
func (u *FileUploader) UploadCSVBytes(data []byte, fileType, locationIdentifier string) (*SFTPUploadResult, error) { ... }

// IsSFTPConfigured — checks if R365_FTP_HOST, USERNAME, PASSWORD are all set
func IsSFTPConfigured() bool { ... }
```

**Additional env vars:**

| Variable | Default | Purpose |
|----------|---------|---------|
| `R365_FTP_HOST` | — | SFTP server hostname |
| `R365_FTP_PORT` | `22` | SFTP port |
| `R365_FTP_USERNAME` | — | SFTP username |
| `R365_FTP_PASSWORD` | — | SFTP password |
| `R365_FTP_PROTOCOL` | `sftp` | Only "sftp" supported |
| `R365_FTP_TIMEOUT` | `30s` | Connection timeout |
| `R365_FTP_MAX_RETRIES` | `3` | Reconnect attempts |
| `R365_FTP_RETRY_DELAY` | `5s` | Delay between retries |
| `R365_FTP_KEEPALIVE_INTERVAL` | `15s` | SSH keep-alive interval |
| `R365_FTP_REMOTE_PATH` | `/` | Remote directory for uploads |

---

## 30. Firebase Config

**File:** `internal/config/firebase.go`

Firebase credentials are reconstructed from individual env vars (set by the config service) into a JSON service account blob at runtime.

```go
package config

import (
    "encoding/json"
    "os"
)

type FirebaseWebConfig struct {
    ApiKey            string
    AuthDomain        string
    ProjectID         string
    StorageBucket     string
    MessagingSenderID string
    AppID             string
    MeasurementID     string
    DatabaseURL       string
}

func GetFirebaseDatabaseURL() string {
    if v := os.Getenv("FIREBASE_DB_URL"); v != "" {
        return v
    }
    return "https://restaurant-pos-terminal-default-rtdb.firebaseio.com"
}

// GetFirebaseServiceAccountJSON reconstructs the service account JSON from
// individual env vars. This is needed because the config service stores each
// field separately, not as a single JSON blob.
func GetFirebaseServiceAccountJSON() ([]byte, error) {
    data := map[string]interface{}{
        "type":                        os.Getenv("FIREBASE_AUTH_TYPE"),
        "project_id":                  os.Getenv("FIREBASE_PROJECT_ID"),
        "private_key_id":              os.Getenv("FIREBASE_PRIVATE_KEY_ID"),
        "private_key":                 os.Getenv("FIREBASE_PRIVATE_KEY"),
        "client_email":                os.Getenv("FIREBASE_CLIENT_EMAIL"),
        "client_id":                   os.Getenv("FIREBASE_CLIENT_ID"),
        "auth_uri":                    os.Getenv("FIREBASE_AUTH_URI"),
        "token_uri":                   os.Getenv("FIREBASE_TOKEN_URI"),
        "auth_provider_x509_cert_url": os.Getenv("FIREBASE_AUTH_PROVIDER_CERT_URL"),
        "client_x509_cert_url":        os.Getenv("FIREBASE_CLIENT_CERT_URL"),
    }
    return json.Marshal(data)
}
```

**Additional env vars:**

| Variable | Purpose |
|----------|---------|
| `FIREBASE_AUTH_TYPE` | Usually `"service_account"` |
| `FIREBASE_PROJECT_ID` | GCP project ID |
| `FIREBASE_PRIVATE_KEY_ID` | Key ID |
| `FIREBASE_PRIVATE_KEY` | PEM key (sanitized by `LoadEnvironment`) |
| `FIREBASE_CLIENT_EMAIL` | Service account email |
| `FIREBASE_CLIENT_ID` | Client ID |
| `FIREBASE_AUTH_URI` | OAuth auth endpoint |
| `FIREBASE_TOKEN_URI` | Token endpoint |
| `FIREBASE_AUTH_PROVIDER_CERT_URL` | Auth provider cert URL |
| `FIREBASE_CLIENT_CERT_URL` | Client cert URL |
| `FIREBASE_DB_URL` | Realtime Database URL |

---

## 31. Vision AI Config

**File:** `internal/config/visionai.go`

Configurable event filtering for vision AI camera data.

```go
package config

import (
    "os"
    "strings"
    "reports-command-service/internal/shared/types"
)

// GetQueueLocation returns the camera location to filter on.
// Defaults to "QUEUE". Override via VISION_QUEUE_LOCATION.
func GetQueueLocation() string {
    if v := os.Getenv("VISION_QUEUE_LOCATION"); v != "" {
        return strings.ToUpper(strings.TrimSpace(v))
    }
    return types.CameraLocationQueue
}

// GetQueueEventTypes returns which event types trigger queue notifications.
// Defaults to ["line_cross_enter", "entry"]. Override via VISION_QUEUE_EVENT_TYPES (comma-separated).
func GetQueueEventTypes() []string {
    if v := os.Getenv("VISION_QUEUE_EVENT_TYPES"); v != "" {
        parts := strings.Split(v, ",")
        res := make([]string, 0, len(parts))
        for _, p := range parts {
            p = strings.TrimSpace(p)
            if p != "" {
                res = append(res, strings.ToLower(p))
            }
        }
        if len(res) > 0 { return res }
    }
    return []string{"line_cross_enter", "entry"}
}

// IsQueueEventType checks if an event matches the configured queue event types.
func IsQueueEventType(event string) bool {
    e := strings.ToLower(strings.TrimSpace(event))
    for _, allowed := range GetQueueEventTypes() {
        if e == allowed { return true }
    }
    return false
}
```

**Additional env vars:**

| Variable | Default | Purpose |
|----------|---------|---------|
| `VISION_QUEUE_LOCATION` | `QUEUE` | Camera location filter |
| `VISION_QUEUE_EVENT_TYPES` | `line_cross_enter,entry` | Comma-separated event types |

---

## 32. Shared Helpers

**Directory:** `internal/shared/helpers/`

### CSV Generator (reflection-based)

Converts any slice of structs to CSV bytes using `csv` struct tags.

```go
// internal/shared/helpers/csv_generator.go

// GenerateCSVFromStructs converts []SomeStruct → []byte (CSV)
// Uses the 'csv' struct tag to determine column names.
func GenerateCSVFromStructs(data interface{}) ([]byte, error) {
    val := reflect.ValueOf(data)
    if val.Kind() != reflect.Slice { return nil, fmt.Errorf("data must be a slice") }
    if val.Len() == 0 { return nil, fmt.Errorf("data slice is empty") }

    firstElem := val.Index(0)
    headers, fieldIndices := extractCSVHeaders(firstElem.Type())

    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    writer.Write(headers)

    for i := 0; i < val.Len(); i++ {
        elem := val.Index(i)
        row := make([]string, len(fieldIndices))
        for j, idx := range fieldIndices {
            row[j] = formatCSVValue(elem.Field(idx))
        }
        writer.Write(row)
    }
    writer.Flush()
    return buf.Bytes(), writer.Error()
}
```

**Usage with partner export structs:**

```go
type SalesDetailRow struct {
    OrderID     string  `csv:"OrderID"`
    GrossSales  float64 `csv:"GrossSales"`
    NetSales    float64 `csv:"NetSales"`
    // ...
}

csvBytes, err := helpers.GenerateCSVFromStructs(salesRows)
// → upload to S3 or SFTP
```

### Struct Validator (reflection-based)

Uses `validate` and `default` struct tags to validate/set-defaults on any struct.

```go
// internal/shared/helpers/safe_type.go

// Tags:
//   validate:"required"                — returns error if field is zero-value
//   validate:"setDefault" default:"x"  — sets field to "x" if zero-value

func ValidateStruct(s interface{}) error {
    // Recursively validates nested structs
    // Handles: string, int, bool, ptr, slice, map
}
```

**Usage:**

```go
type OrderInput struct {
    RefID  string `validate:"required"`
    Status string `validate:"setDefault" default:"pending"`
}

input := OrderInput{RefID: "abc"}
helpers.ValidateStruct(&input) // input.Status is now "pending"
```

### Struct Comparison (reflection-based)

Deep-compares two structs of the same type and returns a list of differences. Used by the `/internal/orders/compare-order/:refId` resiliency endpoint.

```go
// internal/shared/helpers/compare_struct.go

func CompareFields(v1, v2 reflect.Value, prefix string) []string {
    // Recursively compares all fields including:
    // - Pointers (nil checks + deref)
    // - Slices (length + element-wise)
    // - Nested structs
    // - time.Time (using .Equal())
    // Returns: ["field.SubField: valueA != valueB", ...]
}
```

---

*End of architecture reference.*


Config service config.json
{
  "LOAD_CREDS_FROM_FILE": true,
  "DB_URI": "{@mongodb.connectionString}",
  "DB_NAME": "bakeit360db_{@env}",
  "MERCHANT_COLLECTION": "merchant_accounts",
  "SOURCE_QUEUE_URI": "https://sqs.ap-south-1.amazonaws.com/654654626788/pl-d-as1-EmafReconcileQueue-{@env}",
  "KAFKA_HOST": {@kafka.host},
  "KAFKA_SSL": false,
  "KAFKA_SASL": true,
  "KAFKA_SASL_USERNAME": "{@kafka.username}",
  "KAFKA_SASL_PASSWORD": "{@kafka.password}",
  "PRODUCER_TOPIC": "reconcile-events-{@env}",
  "PRODUCER_EVENT": "RECONCILIATION_MATCHED",
  "LOGGER_HOST": "{@internalUrl.logging}/{@env}-reconcile-service.log",
  "REMOTE_LOG": true
}