package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"stronger/server/service/model"
)

const currentUserIDKey = "current_user_id"

// Claims contains the minimal access-token claims used by the API.
type Claims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies access tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *TokenManager) Issue(userID string) (string, error) {
	now := m.now()
	claims := Claims{
		SessionID: uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) Verify(rawToken string) (Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(rawToken, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, model.ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithTimeFunc(m.now))
	if err != nil || !token.Valid || claims.Subject == "" || claims.SessionID == "" {
		return Claims{}, fmt.Errorf("%w: %v", model.ErrInvalidToken, err)
	}
	return claims, nil
}

// Middleware authenticates Bearer access tokens and stores the current user ID.
func (m *TokenManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			abortUnauthorized(c)
			return
		}

		claims, err := m.Verify(strings.TrimSpace(parts[1]))
		if err != nil {
			abortUnauthorized(c)
			return
		}
		c.Set(currentUserIDKey, claims.Subject)
		c.Next()
	}
}

// CurrentUserID returns the authenticated user ID stored by Middleware.
func CurrentUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(currentUserIDKey)
	userID, ok := value.(string)
	return userID, exists && ok && userID != ""
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    "unauthorized",
		"message": "valid Bearer access token required",
	})
}

// ExpiresIn returns the access-token lifetime in seconds.
func (m *TokenManager) ExpiresIn() int { return int(m.ttl.Seconds()) }
