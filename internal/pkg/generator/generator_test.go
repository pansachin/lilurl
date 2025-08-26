package generator

import (
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Zero", 0, "0"},
		{"One", 1, "1"},
		{"Ten", 10, "A"},
		{"SixtyOne", 61, "z"},
		{"SixtyTwo", 62, "10"},
		{"Large", 3844, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Generator(tt.input)
			if result != tt.expected {
				t.Errorf("Generator(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratorSha256(t *testing.T) {
	url := "https://example.com"
	
	// Test without salt
	result1 := GeneratorSha256(url, "")
	if len(result1) != 8 {
		t.Errorf("Expected length 8, got %d", len(result1))
	}
	
	// Test with salt - should produce different result
	result2 := GeneratorSha256(url, "somesalt")
	if result1 == result2 {
		t.Error("Results with and without salt should be different")
	}
	
	// Test deterministic behavior (same input should produce same output)
	result3 := GeneratorSha256(url, "")
	result4 := GeneratorSha256(url, "")
	if result3 != result4 {
		t.Error("Same input should produce same output when using same RNG seed")
	}
}

func TestGeneratorSha256Secure(t *testing.T) {
	url := "https://example.com"
	
	// Test with different lengths
	lengths := []int{5, 7, 10, 15}
	for _, length := range lengths {
		result, err := GeneratorSha256Secure(url, "", length)
		if err != nil {
			t.Errorf("GeneratorSha256Secure failed: %v", err)
		}
		if len(result) != length {
			t.Errorf("Expected length %d, got %d", length, len(result))
		}
		
		// Check that result contains only base62 characters
		for _, char := range result {
			if !strings.ContainsRune(base62Charset, char) && !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", char) {
				t.Errorf("Invalid character in result: %c", char)
			}
		}
	}
}

func TestNewSalt(t *testing.T) {
	// Generate multiple salts to check randomness
	salts := make(map[string]bool)
	for i := 0; i < 100; i++ {
		salt := NewSalt()
		if len(salt) != 10 {
			t.Errorf("Expected salt length 10, got %d", len(salt))
		}
		salts[salt] = true
	}
	
	// With secure random generation, we should have mostly unique salts
	if len(salts) < 95 {
		t.Errorf("Expected at least 95 unique salts out of 100, got %d", len(salts))
	}
}

func TestShortURLGenerator(t *testing.T) {
	generator := NewShortURLGenerator(7, 100)
	
	// Mock database check function
	existingURLs := map[string]bool{
		"abc1234": true,
		"def5678": true,
	}
	
	checkExistence := func(short string) (bool, error) {
		return existingURLs[short], nil
	}
	
	// Generate URL
	result, err := generator.Generate("https://example.com/test", checkExistence)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if len(result) < 7 {
		t.Errorf("Expected at least 7 characters, got %d", len(result))
	}
	
	// Should not be in existing URLs
	if existingURLs[result] {
		t.Error("Generated URL already exists")
	}
}

func BenchmarkGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generator(int64(i))
	}
}

func BenchmarkGeneratorSha256(b *testing.B) {
	url := "https://example.com/benchmark"
	for i := 0; i < b.N; i++ {
		GeneratorSha256(url, "")
	}
}

func BenchmarkGeneratorSha256Secure(b *testing.B) {
	url := "https://example.com/benchmark"
	for i := 0; i < b.N; i++ {
		_, _ = GeneratorSha256Secure(url, "", 7)
	}
}

func BenchmarkNewSalt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSalt()
	}
}
