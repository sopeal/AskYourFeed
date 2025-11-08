package integration

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	testContainerName = "askyourfeed_test_postgres"
	testDBName        = "askyourfeed_test"
	testDBUser        = "postgres"
	testDBPassword    = "postgres"
	testDBPort        = "5433" // Use different port to avoid conflicts
)

// DockerManager handles PostgreSQL container lifecycle for testing
type DockerManager struct {
	containerName string
	dbName        string
	dbUser        string
	dbPassword    string
	dbPort        string
}

// NewDockerManager creates a new Docker manager with default test configuration
func NewDockerManager() *DockerManager {
	return &DockerManager{
		containerName: testContainerName,
		dbName:        testDBName,
		dbUser:        testDBUser,
		dbPassword:    testDBPassword,
		dbPort:        testDBPort,
	}
}

// SetupDatabase starts PostgreSQL container and creates test database
func (dm *DockerManager) SetupDatabase() error {
	// Check if we should skip database lifecycle management
	if os.Getenv("SKIP_DB_LIFECYCLE") == "true" {
		log.Println("Skipping database lifecycle management (SKIP_DB_LIFECYCLE=true)")
		return nil
	}

	log.Println("Setting up test database...")

	// Start PostgreSQL container
	if err := dm.startContainer(); err != nil {
		return fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Wait for PostgreSQL to be ready
	if err := dm.waitForPostgres(); err != nil {
		dm.Cleanup()
		return fmt.Errorf("PostgreSQL failed to become ready: %w", err)
	}

	// Create test database
	if err := dm.createDatabase(); err != nil {
		dm.Cleanup()
		return fmt.Errorf("failed to create test database: %w", err)
	}

	log.Println("Test database ready")
	return nil
}

// Cleanup stops and removes the PostgreSQL container
func (dm *DockerManager) Cleanup() {
	if os.Getenv("SKIP_DB_LIFECYCLE") == "true" {
		return
	}

	log.Println("Cleaning up test database...")
	dm.stopContainer()
}

// startContainer starts a PostgreSQL container for testing
func (dm *DockerManager) startContainer() error {
	// Check if container already exists
	checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", dm.containerName), "--format", "{{.Names}}")
	output, _ := checkCmd.Output()

	if strings.TrimSpace(string(output)) == dm.containerName {
		log.Printf("Container %s already exists, removing it...", dm.containerName)
		dm.stopContainer()
	}

	// Start new container
	cmd := exec.Command("docker", "run", "-d",
		"--name", dm.containerName,
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", dm.dbPassword),
		"-e", fmt.Sprintf("POSTGRES_USER=%s", dm.dbUser),
		"-p", fmt.Sprintf("%s:5432", dm.dbPort),
		"postgres:16-alpine",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container: %v, output: %s", err, string(output))
	}

	log.Printf("Started PostgreSQL container: %s", dm.containerName)
	return nil
}

// stopContainer stops and removes the PostgreSQL container
func (dm *DockerManager) stopContainer() {
	// Stop container
	stopCmd := exec.Command("docker", "stop", dm.containerName)
	stopCmd.Run()

	// Remove container
	rmCmd := exec.Command("docker", "rm", dm.containerName)
	rmCmd.Run()

	log.Printf("Stopped and removed container: %s", dm.containerName)
}

// waitForPostgres waits for PostgreSQL to be ready to accept connections
func (dm *DockerManager) waitForPostgres() error {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		cmd := exec.Command("docker", "exec", dm.containerName,
			"pg_isready", "-U", dm.dbUser)

		if err := cmd.Run(); err == nil {
			log.Println("PostgreSQL is ready")
			return nil
		}

		log.Printf("Waiting for PostgreSQL to be ready... (%d/%d)", i+1, maxAttempts)
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("PostgreSQL did not become ready in time")
}

// createDatabase creates the test database
func (dm *DockerManager) createDatabase() error {
	cmd := exec.Command("docker", "exec", dm.containerName,
		"psql", "-U", dm.dbUser, "-c", fmt.Sprintf("CREATE DATABASE %s;", dm.dbName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if database already exists
		if strings.Contains(string(output), "already exists") {
			log.Printf("Database %s already exists", dm.dbName)
			return nil
		}
		return fmt.Errorf("failed to create database: %v, output: %s", err, string(output))
	}

	log.Printf("Created test database: %s", dm.dbName)
	return nil
}

// GetDatabaseURL returns the connection URL for the test database
func (dm *DockerManager) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		dm.dbUser, dm.dbPassword, dm.dbPort, dm.dbName)
}
