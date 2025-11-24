package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"taskmanager/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			return
		}

		token, err := jwt.Parse(authParts[1], func(token *jwt.Token) (interface{}, error) {
			// Check the algorithm (safety measure against "none" attacks)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT"})
			return
		}

		// check if token is valid and cast the claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims format"})
			return
		}

		// extract role
		roleFloat, ok := claims["role"].(float64)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Role claim missing or invalid"})
			return
		}

		// set role and userID for subsequent handlers
		ctx.Set("role", models.UserRole(roleFloat))
		ctx.Set("user_id", claims["user_id"])

		ctx.Next()
	}

}

func AuthorizationMiddleware(requiredRole models.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRoleVal, exists := ctx.Get("role")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		userRole, ok := userRoleVal.(models.UserRole)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format in context"})
			return
		}

		if userRole < requiredRole {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		ctx.Next()

	}
}
