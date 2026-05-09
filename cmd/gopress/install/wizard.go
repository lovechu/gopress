package install

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// RunWizard runs the interactive installation wizard
func RunWizard() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &Config{}

	// 1. Database type selection
	fmt.Println("Step 1/5: Database Configuration")
	fmt.Println("--------------------------------")
	dbType := promptSelect(reader, "Select database type", []string{
		"1. SQLite (lightweight, no setup required)",
		"2. PostgreSQL",
		"3. MySQL",
	})
	cfg.DatabaseType = mapDBType(dbType)
	fmt.Printf("Selected: %s\n\n", displayDBType(cfg.DatabaseType))

	// 2. Database connection details
	fmt.Println("Step 2/5: Database Connection")
	fmt.Println("------------------------------")

	switch cfg.DatabaseType {
	case "sqlite":
		cfg.DBName = promptRequired(reader, "Database file name (without .db extension)")
		cfg.DBCharset = "utf8"

	case "mysql", "postgres":
		cfg.DBHost = promptWithDefault(reader, "Database host", "localhost")
		cfg.DBPort = promptIntWithDefault(reader, "Database port", mapDBDefaultPort(cfg.DatabaseType))
		cfg.DBName = promptRequired(reader, "Database name")
		cfg.DBUser = promptWithDefault(reader, "Database username", "root")
		cfg.DBPassword = promptPassword(reader)
		cfg.DBCharset = "utf8mb4"
		if cfg.DatabaseType == "postgres" {
			cfg.DBCharset = "utf8"
		}
	}

	fmt.Println()

	// 3. Redis configuration
	fmt.Println("Step 3/5: Redis Configuration")
	fmt.Println("-----------------------------")
	cfg.RedisAddr = promptWithDefault(reader, "Redis address (host:port)", "localhost:6379")
	cfg.RedisPassword = promptWithDefault(reader, "Redis password (leave empty for none)", "")
	cfg.RedisDB = promptIntWithDefault(reader, "Redis database number", 0)
	fmt.Println()

	// 4. Admin account
	fmt.Println("Step 4/5: Admin Account")
	fmt.Println("------------------------")
	cfg.AdminUsername = promptUsername(reader)
	cfg.AdminEmail = promptEmail(reader)
	cfg.AdminPassword = promptPasswordStrength(reader)
	fmt.Println()

	// 5. Site settings
	fmt.Println("Step 5/5: Site Settings")
	fmt.Println("------------------------")
	cfg.SiteName = promptWithDefault(reader, "Site name", "My GoPress Site")
	cfg.SiteURL = promptWithDefault(reader, "Site URL", "http://localhost:8080")
	fmt.Println()

	return cfg, nil
}

// promptSelect presents a numbered selection menu
func promptSelect(reader *bufio.Reader, prompt string, options []string) int {
	fmt.Println(prompt)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Printf("  Enter choice (1-%d): ", len(options))

	for {
		fmt.Fscan(reader, &input)
		reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if choice, err := parseInt(input); err == nil && choice >= 1 && choice <= len(options) {
			return choice
		}
		fmt.Printf("Invalid input. Please enter a number between 1 and %d: ", len(options))
	}
}

var input string // temp variable for scanning

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// promptRequired asks for a required input
func promptRequired(reader *bufio.Reader, prompt string) string {
	for {
		fmt.Print(prompt + ": ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Println("This field is required. Please enter a value.")
	}
}

// promptWithDefault asks for input with a default value
func promptWithDefault(reader *bufio.Reader, prompt string, defaultVal string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// promptIntWithDefault asks for integer input with a default value
func promptIntWithDefault(reader *bufio.Reader, prompt string, defaultVal int) int {
	for {
		fmt.Printf("%s [%d]: ", prompt, defaultVal)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return defaultVal
		}
		if val, err := parseInt(input); err == nil {
			return val
		}
		fmt.Println("Please enter a valid number.")
	}
}

// promptUsername validates and returns a username
func promptUsername(reader *bufio.Reader) string {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

	for {
		fmt.Print("Admin username (3-32 alphanumeric characters, underscores allowed): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Println("Username is required.")
			continue
		}

		if !usernameRegex.MatchString(input) {
			fmt.Println("Invalid username. Use 3-32 alphanumeric characters or underscores.")
			continue
		}

		return input
	}
}

// promptEmail validates and returns an email address
func promptEmail(reader *bufio.Reader) string {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	for {
		fmt.Print("Admin email address: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Println("Email is required.")
			continue
		}

		if !emailRegex.MatchString(input) {
			fmt.Println("Invalid email format. Please enter a valid email address.")
			continue
		}

		return input
	}
}

// promptPassword asks for a password with strength validation
func promptPassword(reader *bufio.Reader) string {
	fmt.Print("Database password: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptPasswordStrength asks for a password and validates its strength
func promptPasswordStrength(reader *bufio.Reader) string {
	for {
		fmt.Print("Admin password (minimum 8 characters): ")
		password := promptHidden(reader)

		if err := ValidatePasswordStrength(password); err != nil {
			fmt.Printf("Password validation failed: %s\n\n", err.Error())
			continue
		}

		// Confirm password
		fmt.Print("\nConfirm password: ")
		confirm := promptHidden(reader)
		fmt.Println()

		if password != confirm {
			fmt.Println("Passwords do not match. Please try again.")
			continue
		}

		return password
	}
}

// promptHidden reads password input without echoing
func promptHidden(reader *bufio.Reader) string {
	// Try to use syscall for password masking if available
	fmt.Fscan(reader, &input)
	reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// ValidatePasswordStrength checks if a password meets minimum strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be no more than 128 characters")
	}

	// Check for uppercase letters
	hasUpper := false
	// Check for lowercase letters
	hasLower := false
	// Check for digits
	hasDigit := false
	// Check for special characters
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", c):
			hasSpecial = true
		}
	}

	// Require at least 3 of: uppercase, lowercase, digits, special
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	if score < 3 {
		return fmt.Errorf("password must contain at least 3 of: uppercase letters, lowercase letters, digits, special characters")
	}

	return nil
}

// GetPasswordStrengthScore returns a score (0-4) for password strength
func GetPasswordStrengthScore(password string) int {
	score := 0

	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", c):
			hasSpecial = true
		}
	}

	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	// Cap at 4
	if score > 4 {
		score = 4
	}

	return score
}

func mapDBType(choice int) string {
	switch choice {
	case 1:
		return "sqlite"
	case 2:
		return "postgres"
	case 3:
		return "mysql"
	default:
		return "sqlite"
	}
}

func displayDBType(dbType string) string {
	switch dbType {
	case "sqlite":
		return "SQLite"
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	default:
		return dbType
	}
}

func mapDBDefaultPort(dbType string) int {
	switch dbType {
	case "postgres":
		return 5432
	case "mysql":
		return 3306
	default:
		return 3306
	}
}
