package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/gopress/cmd/gopress/install"
	"github.com/yourorg/gopress/internal/bootstrap"
	"github.com/yourorg/gopress/internal/config"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	showVer   bool
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gopress",
		Short: "GoPress - A modern content management system",
		Long: `GoPress is a powerful, extensible content management system
built with Go and Gin framework.`,
		Run: func(cmd *cobra.Command, args []string) {
			if showVer {
				fmt.Printf("GoPress version %s (commit: %s, date: %s)\n", version, commit, date)
				return
			}
			cmd.Help()
		},
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GoPress version %s (commit: %s, date: %s)\n", version, commit, date)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Install command
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Initialize GoPress and configure database and admin account",
		Long: `Run the interactive installation wizard to configure:
- Database connection (PostgreSQL/MySQL/SQLite)
- Redis connection
- Admin account credentials
- Site settings`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := install.Run(); err != nil {
				log.Fatalf("Installation failed: %v", err)
			}
		},
	}
	rootCmd.AddCommand(installCmd)

	// Serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the GoPress server",
		Long: `Start the GoPress HTTP server with the configuration from config.yaml`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
	rootCmd.AddCommand(serveCmd)

	// Wire up config flag for serve
	serveCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Ensure config flag is available
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer() {
	// Load configuration
	cfg, err := bootstrap.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize application
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to bootstrap app: %v", err)
	}

	// Start HTTP server
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[GoPress] Server started on %s (env=%s)", addr, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[GoPress] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("[GoPress] Goodbye!")
}

// GetVersion returns the current version string
func GetVersion() string {
	return version
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	if configPath == "" {
		return "config.yaml"
	}
	return configPath
}

// LoadAppConfig loads and returns the application configuration
func LoadAppConfig() (*config.Config, error) {
	return config.Load(GetConfigPath())
}
