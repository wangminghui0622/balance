package controllers

import (
	"balance/admin/dto/auth"
	"balance/admin/services"
	"balance/internal/constants"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ctrl *AuthController) GetByPartnerId(c *gin.Context) {
	authCfg, err := ctrl.authService.GetByPartnerId()
	if err != nil {
		c.JSON(http.StatusOK, auth.AuthResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, auth.AuthResponse{
		Code:    constants.ResponseCodeSuccess,
		Message: "success",
		Data: auth.AuthData{
			PartnerID:  authCfg.PartnerID,
			PartnerKey: authCfg.PartnerKey,
			Redirect:   authCfg.Redirect,
			IsSandbox:  authCfg.IsSandbox,
		},
	})
}
