package middleware

import (
	"fmt"
	"gambler/backend/tools"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Jwt struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func Sign(username string) (*Jwt, int) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "Gambler Backend Service",
		Subject:   username,
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "Gambler Backend Service",
		Subject:   username,
	})
	AccessToken, err := accessToken.SignedString(tools.JWT_SECRET)
	if err != nil {
		return nil, tools.JWT_FAILED_TO_SIGN
	}
	RefreshToken, err := refreshToken.SignedString(tools.JWT_SECRET)
	if err != nil {
		return nil, tools.JWT_FAILED_TO_SIGN
	}
	return &Jwt{
		AccessToken,
		RefreshToken,
	}, -1
}

func Decode(token string) (jwt.Claims, int) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return tools.JWT_SECRET, nil
	})
	if err != nil {
		return nil, tools.JWT_FAILED_TO_DECODE
	}
	if !t.Valid {
		rawExpireIn := t.Claims.(jwt.MapClaims)["exp"]
		expireIn := time.Now().Unix() - rawExpireIn.(int64)
		if expireIn > 0 {
			return nil, tools.JWT_EXPIRED
		} else {
			return nil, tools.JWT_INVALID
		}
	}
	return t.Claims.(jwt.Claims), -1
}

func JwtGuardHandler(c *fiber.Ctx) error {
	// Check if the request context is authorized
	// If not, return an error
	// If it is, continue to the next handler
	rawIsAuthorized := c.Locals("isAuthorized")
	rawExpireIn := c.Locals("expireIn")
	if rawIsAuthorized != nil && rawExpireIn != nil {
		isAuthorized := rawIsAuthorized.(bool)
		expireIn := time.Now().Unix() - rawExpireIn.(int64)
		fmt.Println(isAuthorized, expireIn)
		if isAuthorized && expireIn < 0 {
			return c.Next()
		}
	}

	// Check if the request is authorized
	// If not, return an error
	headers := c.GetReqHeaders()
	if len(headers) == 0 {
		return c.Status(400).JSON(tools.GlobalErrorHandlerResp{
			Success: false,
			Message: "Bad Request, no headers",
			Code:    400,
		})
	}
	token := headers["Authorization"][0]
	if !strings.HasPrefix(token, "Bearer ") {
		return c.Status(400).JSON(tools.GlobalErrorHandlerResp{
			Success: false,
			Message: "Bad Request, invalid token",
			Code:    400,
		})
	}
	token = strings.TrimPrefix(token, "Bearer ")

	claims, err := Decode(token)
	if err != -1 {
		if err == tools.JWT_FAILED_TO_DECODE {
			return c.Status(401).JSON(tools.GlobalErrorHandlerResp{
				Success: false,
				Message: "Failed to decode token",
				Code:    401,
			})
		} else if err == tools.JWT_INVALID {
			return c.Status(401).JSON(tools.GlobalErrorHandlerResp{
				Success: false,
				Message: "Invalid token",
				Code:    401,
			})
		}
	}

	exp, tErr := claims.GetExpirationTime()
	if tErr != nil {
		return c.Status(401).JSON(tools.GlobalErrorHandlerResp{
			Success: false,
			Message: "Invalid token",
			Code:    401,
		})
	}

	c.Locals("claims", claims)
	c.Locals("isAuthorized", true)
	c.Locals("expireIn", exp.Unix())
	return c.Next()
}