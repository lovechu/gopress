package install

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourorg/gopress/internal/user"
)

// Config holds all installation configuration
type Config struct {
	// Database settings
	DatabaseType   string // "mysql", "postgres", "sqlite"
	DBHost         string
	DBPort         int
	DBName         string
	DBUser         string
	DBPassword     string
	DBCharset      string

	// Redis settings
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Admin account
	AdminUsername string
	AdminEmail    string
	AdminPassword string

	// Site settings
	SiteName string
	SiteURL  string
}

// Run executes the installation process
func Run() error {
	fmt.Println("========================================")
	fmt.Println("       Welcome to GoPress Installer     ")
	fmt.Println("========================================")
	fmt.Println()

	// Run the interactive wizard
	cfg, err := RunWizard()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	// Create config.yaml
	if err := generateConfigFile(cfg); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Initialize database
	if err := initializeDatabase(cfg); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create admin user
	if err := createAdminUser(cfg); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Installation completed successfully!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Run 'gopress serve' to start the server")
	fmt.Println()

	return nil
}

// generateConfigFile creates config.yaml from the installation config
func generateConfigFile(cfg *Config) error {
	configContent := fmt.Sprintf(`app:
  name: %s
  env: development
  port: "8080"
  base_url: %s
  base_path: .
  debug: true
  timezone: UTC
  secret_key: "change-this-in-production-$(openssl rand -hex 32)"

database:
  driver: %s
  host: %s
  port: %d
  name: %s
  user: %s
  password: %s
  charset: %s
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300
  log_level: info

redis:
  addr: %s
  password: %s
  db: %d
  pool_size: 10
  prefix: gopress:

jwt:
  access_secret: "change-this-in-production-access"
  refresh_secret: "change-this-in-production-refresh"
  access_expire: 3600
  refresh_expire: 604800

storage:
  driver: local

upload:
  max_size: 10485760
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
  image_quality: 85
  generate_webp: true

media:
  storage: local
  max_file_size: 10485760
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
  local:
    base_path: ./uploads
    base_url: /uploads

cache:
  driver: memory
  default_ttl: 3600

log:
  level: info
  format: json
  output: stdout

cors:
  allowed_origins:
    - "*"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:
    - "*"
  max_age: 86400

rate_limit:
  enabled: true
  requests: 100
  window: 60
`,
		escapeYAML(cfg.SiteName),
		escapeYAML(cfg.SiteURL),
		cfg.DatabaseType,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
		escapeYAML(cfg.DBPassword),
		cfg.DBCharset,
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)

	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Println("[OK] Configuration file created: config.yaml")
	return nil
}

// escapeYAML escapes special characters in YAML string values
func escapeYAML(s string) string {
	// Simple escaping for YAML strings
	result := ""
	for _, c := range s {
		switch c {
		case '"':
			result += `\"`
		case '\n':
			result += "\\n"
		default:
			result += string(c)
		}
	}
	return result
}

// initializeDatabase creates database tables
func initializeDatabase(cfg *Config) error {
	fmt.Println()
	fmt.Println("Initializing database...")

	db, err := openDatabase(cfg)
	if err != nil {
		return err
	}

	// Run auto migrations
	if err := db.AutoMigrate(
		&user.User{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	fmt.Println("[OK] Database tables created")
	return nil
}

// openDatabase opens a connection to the configured database
func openDatabase(cfg *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DatabaseType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBCharset)
		dialector = mysql.Open(dsn)

	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)

	case "sqlite":
		dialector = sqlite.Open(cfg.DBName + ".db")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DatabaseType)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

// createAdminUser creates the initial admin user
func createAdminUser(cfg *Config) error {
	fmt.Println()
	fmt.Println("Creating admin user...")

	db, err := openDatabase(cfg)
	if err != nil {
		return err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	admin := &user.User{
		Username:     cfg.AdminUsername,
		Email:        cfg.AdminEmail,
		PasswordHash: string(hashedPassword),
		DisplayName:  "Administrator",
		Role:         user.RoleAdmin,
		IsActive:     true,
	}

	// Check if admin already exists
	var existing user.User
	if err := db.Where("username = ? OR email = ?", cfg.AdminUsername, cfg.AdminEmail).First(&existing).Error; err == nil {
		fmt.Printf("[SKIP] Admin user '%s' already exists\n", cfg.AdminUsername)
		return nil
	}

	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	fmt.Printf("[OK] Admin user created: %s\n", cfg.AdminUsername)
	return nil
}
