package redisx

import (
	"context"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestPrivateUrlWithHostname(t *testing.T) {
	// Skip test if Redis is not available
	if !isRedisAvailable() {
		t.Skip("Redis is not available, skipping test")
	}

	// Clean up test data before and after test
	cleanup := func() {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
		defer rdb.Close()
		rdb.Del(context.Background(), "private_db", "private_db:counter")
	}

	cleanup()
	defer cleanup()

	tests := []struct {
		name     string
		url      string
		hostname string
		wantErr  bool
	}{
		{
			name:     "valid URL and hostname",
			url:      "redis://localhost:6379/0",
			hostname: "test-host-1",
			wantErr:  false,
		},
		{
			name:     "empty URL",
			url:      "",
			hostname: "test-host",
			wantErr:  true,
		},
		{
			name:     "empty hostname",
			url:      "redis://localhost:6379/0",
			hostname: "",
			wantErr:  true,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			hostname: "test-host",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := privateUrlWithHostname(tt.url, tt.hostname)

			if tt.wantErr {
				if err == nil {
					t.Errorf("privateUrlWithHostname() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("privateUrlWithHostname() unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Errorf("privateUrlWithHostname() returned empty result")
			}

			// For valid cases, verify the result is a proper Redis URL
			if tt.name == "valid URL and hostname" {
				expectedPrefix := "redis://localhost:6379/"
				if len(result) <= len(expectedPrefix) {
					t.Errorf("privateUrlWithHostname() result too short: %s", result)
				}
				if result[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("privateUrlWithHostname() result doesn't start with expected prefix. Got: %s", result)
				}
			}
		})
	}
}

func TestPrivateUrlWithHostname_Consistency(t *testing.T) {
	// Skip test if Redis is not available
	if !isRedisAvailable() {
		t.Skip("Redis is not available, skipping test")
	}

	// Clean up test data
	cleanup := func() {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
		defer rdb.Close()
		rdb.Del(context.Background(), "private_db", "private_db:counter")
	}

	cleanup()
	defer cleanup()

	url := "redis://localhost:6379/0"
	hostname1 := "test-host-1"
	hostname2 := "test-host-2"

	// First call for hostname1
	result1a, err := privateUrlWithHostname(url, hostname1)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call for hostname1 should return the same result
	result1b, err := privateUrlWithHostname(url, hostname1)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if result1a != result1b {
		t.Errorf("Results for same hostname should be consistent. Got %s and %s", result1a, result1b)
	}

	// Call for hostname2 should return different result
	result2, err := privateUrlWithHostname(url, hostname2)
	if err != nil {
		t.Fatalf("Call for hostname2 failed: %v", err)
	}

	if result1a == result2 {
		t.Errorf("Results for different hostnames should be different. Got %s for both", result1a)
	}
}

func TestPrivateUrlWithHostname_AutoIncrement(t *testing.T) {
	// Skip test if Redis is not available
	if !isRedisAvailable() {
		t.Skip("Redis is not available, skipping test")
	}

	// Clean up test data
	cleanup := func() {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
		defer rdb.Close()
		rdb.Del(context.Background(), "private_db", "private_db:counter")
	}

	cleanup()
	defer cleanup()

	url := "redis://localhost:6379/0"

	// Test multiple hostnames to verify auto-increment behavior
	hostnames := []string{"host-1", "host-2", "host-3"}
	expectedDBs := []string{"1", "2", "3"}

	for i, hostname := range hostnames {
		result, err := privateUrlWithHostname(url, hostname)
		if err != nil {
			t.Fatalf("Call for %s failed: %v", hostname, err)
		}

		expected := "redis://localhost:6379/" + expectedDBs[i]
		if result != expected {
			t.Errorf("Expected %s but got %s for hostname %s", expected, result, hostname)
		}
	}
}

// isRedisAvailable checks if Redis is available for testing
func isRedisAvailable() bool {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer rdb.Close()

	err := rdb.Ping(context.Background()).Err()
	return err == nil
}

func TestPrivateUrlWithHostname_ConnectionError(t *testing.T) {
	// Test with invalid Redis URL to verify error handling
	url := "redis://invalid-host:9999/0"
	hostname := "test-host"

	_, err := privateUrlWithHostname(url, hostname)
	if err == nil {
		t.Errorf("Expected error when connecting to invalid Redis host, but got none")
	}

	// Verify error message contains connection-related information
	if err != nil && !strings.Contains(err.Error(), "failed to") {
		t.Errorf("Expected connection error message, got: %v", err)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	// Skip test if Redis is not available
	if !isRedisAvailable() {
		t.Skip("Redis is not available, skipping test")
	}

	// Clean up test data
	cleanup := func() {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
		defer rdb.Close()
		rdb.Del(context.Background(), "private_db", "private_db:counter")
	}

	cleanup()
	defer cleanup()

	t.Run("default client without options", func(t *testing.T) {
		client, err := NewClient("redis://localhost:6379/0", nil)
		if err != nil {
			t.Errorf("NewClient() unexpected error: %v", err)
			return
		}
		if client == nil {
			t.Errorf("NewClient() returned nil client")
			return
		}
		defer client.Close()

		// Parse URL to verify it's using DB 0
		options, err := redis.ParseURL(client.url)
		if err != nil {
			t.Errorf("Failed to parse client URL: %v", err)
			return
		}
		if options.DB != 0 {
			t.Errorf("Expected default client to use DB 0, but got DB %d", options.DB)
		}
	})

	t.Run("client with WithPrivate option", func(t *testing.T) {
		client, err := NewClient("redis://localhost:6379/0", nil, WithPrivate())
		if err != nil {
			t.Errorf("NewClient() unexpected error: %v", err)
			return
		}
		if client == nil {
			t.Errorf("NewClient() returned nil client")
			return
		}
		defer client.Close()

		// Parse the final URL to check DB number
		options, err := redis.ParseURL(client.url)
		if err != nil {
			t.Errorf("Failed to parse client URL: %v", err)
			return
		}

		if options.DB == 0 {
			t.Errorf("Expected private client to use non-zero DB, but got DB %d", options.DB)
		}

		t.Logf("Private client using DB %d", options.DB)
	})

	t.Run("invalid URL with options", func(t *testing.T) {
		_, err := NewClient("not-a-url", nil, WithPrivate())
		if err == nil {
			t.Errorf("NewClient() expected error but got none")
		}
	})
}
