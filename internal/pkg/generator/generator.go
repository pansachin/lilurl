package generator

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

const base62Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	// Use crypto/rand for secure random number generation
	cryptoRandReader = rand.Reader
	base62Len        = big.NewInt(62)
)

// Generator generates a base62 string from an integer (unchanged, as it's deterministic and works well)
func Generator(id int64) string {
	var result strings.Builder

	if id == 0 {
		return base62Charset[0:1]
	}

	for id > 0 {
		result.WriteByte(base62Charset[id%62])
		id /= 62
	}

	return reverse(result.String())
}

func reverse(str string) string {
	result := []rune(str)

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// GeneratorSha256Secure generates a secure short URL using SHA256 and secure randomness
func GeneratorSha256Secure(url string, salt string, length int) (string, error) {
	if length <= 0 {
		length = 7 // Default length matching your schema
	}

	// Combine URL and salt
	data := url
	if salt != "" {
		data = url + salt
	}

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert to base64 for more entropy than hex
	hashB64 := base64.RawURLEncoding.EncodeToString(hash[:])

	// Remove any padding and special characters
	hashB64 = strings.ReplaceAll(hashB64, "-", "")
	hashB64 = strings.ReplaceAll(hashB64, "_", "")

	// If the hash is long enough, take the first N characters
	if len(hashB64) >= length {
		return hashB64[:length], nil
	}

	// Otherwise, use the hash as a seed for secure random generation
	return generateSecureRandom(length)
}

// GeneratorSha256 maintains backward compatibility
func GeneratorSha256(url string, salt string) string {
	result, _ := GeneratorSha256Secure(url, salt, 8)
	return result
}

// generateSecureRandom generates a cryptographically secure random string
func generateSecureRandom(length int) (string, error) {
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(cryptoRandReader, base62Len)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = base62Charset[num.Int64()]
	}

	return string(result), nil
}

// NewSaltSecure generates a cryptographically secure random salt
func NewSaltSecure(length int) (string, error) {
	if length <= 0 {
		length = 10 // Default length
	}
	return generateSecureRandom(length)
}

// NewSalt maintains backward compatibility
func NewSalt() string {
	salt, _ := NewSaltSecure(10)
	return salt
}

// ShortURLGenerator provides a more robust URL shortening solution
type ShortURLGenerator struct {
	mu              sync.Mutex
	collisionCache  map[string]bool
	maxCacheSize    int
	preferredLength int
}

// NewShortURLGenerator creates a new generator with collision tracking
func NewShortURLGenerator(preferredLength int, maxCacheSize int) *ShortURLGenerator {
	if preferredLength <= 0 {
		preferredLength = 7
	}
	if maxCacheSize <= 0 {
		maxCacheSize = 10000
	}

	return &ShortURLGenerator{
		collisionCache:  make(map[string]bool),
		maxCacheSize:    maxCacheSize,
		preferredLength: preferredLength,
	}
}

// Generate creates a short URL with collision detection
func (g *ShortURLGenerator) Generate(url string, checkExistence func(string) (bool, error)) (string, error) {
	maxAttempts := 10
	salt := ""

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate candidate
		candidate, err := GeneratorSha256Secure(url, salt, g.preferredLength)
		if err != nil {
			return "", err
		}

		// Check in local cache first
		g.mu.Lock()
		if g.collisionCache[candidate] {
			g.mu.Unlock()
			salt, _ = NewSaltSecure(10)
			continue
		}
		g.mu.Unlock()

		// Check in database
		exists, err := checkExistence(candidate)
		if err != nil {
			return "", err
		}

		if !exists {
			// Add to cache
			g.mu.Lock()
			if len(g.collisionCache) >= g.maxCacheSize {
				// Simple cache eviction - remove 10% of entries
				count := 0
				for k := range g.collisionCache {
					delete(g.collisionCache, k)
					count++
					if count >= g.maxCacheSize/10 {
						break
					}
				}
			}
			g.collisionCache[candidate] = true
			g.mu.Unlock()

			return candidate, nil
		}

		// Generate new salt for next attempt
		salt, _ = NewSaltSecure(10)
	}

	// If all attempts failed, try with a longer URL
	return GeneratorSha256Secure(url, salt, g.preferredLength+2)
}
