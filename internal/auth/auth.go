package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const currentUserIDKey = "current_user_id"

var ErrInvalidToken = errors.New("invalid access token")

// DevUserEnsurer creates the local-only user needed by development login.
type DevUserEnsurer interface {
	EnsureDevUser(context.Context, string) error
}

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
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithTimeFunc(m.now))
	if err != nil || !token.Valid || claims.Subject == "" || claims.SessionID == "" {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
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

// RegisterRoutes registers development login and a minimal authenticated identity endpoint.
func RegisterRoutes(router *gin.RouterGroup, environment string, tokens *TokenManager, devUsers DevUserEnsurer) {
	router.POST("/auth/dev/login", func(c *gin.Context) {
		if environment == "production" {
			c.JSON(http.StatusNotFound, gin.H{"code": "not_found", "message": "route not found"})
			return
		}

		var request struct {
			UserID string `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "request body must be valid JSON"})
			return
		}
		if request.UserID == "" {
			request.UserID = "00000000-0000-0000-0000-000000000001"
		}
		if _, err := uuid.Parse(request.UserID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_user_id", "message": "user_id must be a UUID"})
			return
		}
		if devUsers != nil {
			if err := devUsers.EnsureDevUser(c.Request.Context(), request.UserID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": "dev_user_setup_failed", "message": "could not prepare development user"})
				return
			}
		}

		accessToken, err := tokens.Issue(request.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "token_issue_failed", "message": "could not issue access token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "token_type": "Bearer", "expires_in": int(tokens.ttl.Seconds())})
	})

	router.GET("/me", tokens.Middleware(), func(c *gin.Context) {
		userID, _ := c.Get(currentUserIDKey)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
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
