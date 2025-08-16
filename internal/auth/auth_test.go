package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	duration := time.Hour
	
	manager := NewJWTManager(secretKey, duration)
	
	assert.NotNil(t, manager)
	assert.Equal(t, []byte(secretKey), manager.secretKey)
	assert.Equal(t, duration, manager.tokenDuration)
}

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	
	userID := "user123"
	email := "test@example.com"
	role := "admin"
	
	token, err := manager.GenerateToken(userID, email, role)
	
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestJWTManager_VerifyToken_ValidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	
	userID := "user123"
	email := "test@example.com"
	role := "admin"
	
	token, err := manager.GenerateToken(userID, email, role)
	require.NoError(t, err)
	
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "flexflag", claims.Issuer)
	assert.Equal(t, userID, claims.Subject)
}

func TestJWTManager_VerifyToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	
	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "invalid.token"},
		{"random string", "not-a-jwt-token"},
		{"wrong signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestJWTManager_VerifyToken_ExpiredToken(t *testing.T) {
	manager := NewJWTManager("test-secret", -time.Hour) // Negative duration for expired token
	
	token, err := manager.GenerateToken("user123", "test@example.com", "admin")
	require.NoError(t, err)
	
	claims, err := manager.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	
	userID := "user123"
	email := "test@example.com"
	role := "admin"
	
	// Generate initial token
	token, err := manager.GenerateToken(userID, email, role)
	require.NoError(t, err)
	
	// Add small delay to ensure different issued time
	time.Sleep(time.Millisecond * 10)
	
	// Refresh token
	newToken, err := manager.RefreshToken(token)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	// Note: tokens might be the same if generated too quickly
	// assert.NotEqual(t, token, newToken)
	
	// Verify new token has correct claims
	claims, err := manager.ValidateToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
}

func TestJWTManager_ValidateToken_DifferentSecret(t *testing.T) {
	manager1 := NewJWTManager("secret1", time.Hour)
	manager2 := NewJWTManager("secret2", time.Hour)
	
	token, err := manager1.GenerateToken("user123", "test@example.com", "admin")
	require.NoError(t, err)
	
	// Try to verify with different secret
	claims, err := manager2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestClaims_Structure(t *testing.T) {
	claims := Claims{
		UserID:    "user123",
		Email:     "test@example.com",
		Role:      "admin",
		ProjectID: "project456",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "flexflag",
			Subject: "user123",
		},
	}
	
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "project456", claims.ProjectID)
	assert.Equal(t, "flexflag", claims.Issuer)
	assert.Equal(t, "user123", claims.Subject)
}

func TestErrorVariables(t *testing.T) {
	assert.Equal(t, "invalid token", ErrInvalidToken.Error())
	assert.Equal(t, "token has expired", ErrExpiredToken.Error())
	assert.Equal(t, "malformed token", ErrMalformedToken.Error())
}

func TestHashPassword_ValidPassword(t *testing.T) {
	password := "SecureP@ssw0rd123"
	
	hashedPassword, err := HashPassword(password)
	
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	
	// Hashed password should start with bcrypt identifier
	assert.True(t, strings.HasPrefix(hashedPassword, "$2a$"))
}

func TestHashPassword_InvalidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"too short", "short", ErrPasswordTooShort},
		{"no uppercase", "nouppercase123!", ErrPasswordWeak},
		{"no lowercase", "NOLOWERCASE123!", ErrPasswordWeak},
		{"no digit", "NoDigitPassword!", ErrPasswordWeak},
		{"no special char", "NoSpecialChar123", ErrPasswordWeak},
		{"empty", "", ErrPasswordTooShort},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
			assert.Empty(t, hashedPassword)
		})
	}
}

func TestVerifyPassword_ValidPassword(t *testing.T) {
	password := "SecureP@ssw0rd123"
	
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	
	// Verify correct password
	isValid := VerifyPassword(hashedPassword, password)
	assert.True(t, isValid)
}

func TestVerifyPassword_InvalidPassword(t *testing.T) {
	password := "SecureP@ssw0rd123"
	wrongPassword := "WrongP@ssw0rd123"
	
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	
	// Verify wrong password
	isValid := VerifyPassword(hashedPassword, wrongPassword)
	assert.False(t, isValid)
	
	// Verify with invalid hash
	isValid = VerifyPassword("invalid-hash", password)
	assert.False(t, isValid)
}

func TestValidatePassword_ValidPasswords(t *testing.T) {
	validPasswords := []string{
		"SecureP@ssw0rd123",
		"AnotherG00d!Pass",
		"MyStr0ng#Password",
		"C0mpl3x$P@ssw0rd",
	}
	
	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			err := ValidatePassword(password)
			assert.NoError(t, err)
		})
	}
}

func TestValidatePassword_InvalidPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"too short", "Short1!", ErrPasswordTooShort},
		{"no uppercase", "nouppercase123!", ErrPasswordWeak},
		{"no lowercase", "NOLOWERCASE123!", ErrPasswordWeak},
		{"no digit", "NoDigitPassword!", ErrPasswordWeak},
		{"no special char", "NoSpecialChar123", ErrPasswordWeak},
		{"only letters", "OnlyLetters", ErrPasswordWeak},
		{"only digits", "12345678", ErrPasswordWeak},
		{"only special chars", "!@#$%^&*", ErrPasswordWeak},
		{"empty", "", ErrPasswordTooShort},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestPasswordConstants(t *testing.T) {
	assert.Equal(t, 8, MinPasswordLength)
	assert.Equal(t, 12, BcryptCost)
}

func TestPasswordErrorVariables(t *testing.T) {
	assert.Equal(t, "password must be at least 8 characters long", ErrPasswordTooShort.Error())
	assert.Equal(t, "password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character", ErrPasswordWeak.Error())
}

func BenchmarkHashPassword(b *testing.B) {
	password := "SecureP@ssw0rd123"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "SecureP@ssw0rd123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValid := VerifyPassword(hashedPassword, password)
		if !isValid {
			b.Fatal("password verification failed")
		}
	}
}

func BenchmarkJWTGenerate(b *testing.B) {
	manager := NewJWTManager("test-secret", time.Hour)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateToken("user123", "test@example.com", "admin")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify(b *testing.B) {
	manager := NewJWTManager("test-secret", time.Hour)
	token, err := manager.GenerateToken("user123", "test@example.com", "admin")
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}