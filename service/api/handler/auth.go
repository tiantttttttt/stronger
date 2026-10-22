package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"stronger/server/service/api/logic"
	"stronger/server/service/api/logic/requests"
	"stronger/server/service/api/logic/responses"
	"stronger/server/service/api/middleware"
)

// Auth 处理开发登录与身份查询。
type Auth struct {
	Debug  bool
	Tokens *middleware.TokenManager
}

// @Summary 开发测试登录（仅 DEBUG=true）
// @Tags auth
// @Param body body requests.DevLogin false "开发用户，省略 user_id 使用默认测试用户"
// @Success 200 {object} responses.Login
// @Failure 400,404,500 {object} responses.Error
// @Router /api/v1/auth/dev/login [post]
func (h Auth) DevLogin(c *gin.Context) {
	if !h.Debug {
		c.JSON(http.StatusNotFound, gin.H{"code": "not_found", "message": "route not found"})
		return
	}

	var request requests.DevLogin
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
	if err := logic.AuthLogic.EnsureDevUser(c.Request.Context(), request.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "dev_user_setup_failed", "message": "could not prepare development user"})
		return
	}

	accessToken, err := h.Tokens.Issue(request.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "token_issue_failed", "message": "could not issue access token"})
		return
	}
	c.JSON(http.StatusOK, responses.Login{AccessToken: accessToken, TokenType: "Bearer", ExpiresIn: h.Tokens.ExpiresIn()})
}

// @Summary 获取当前用户身份
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} responses.Identity
// @Failure 401 {object} responses.Error
// @Router /api/v1/me [get]
func (h Auth) Me(c *gin.Context) {
	userID, _ := middleware.CurrentUserID(c)
	c.JSON(http.StatusOK, responses.Identity{UserID: userID})
}
