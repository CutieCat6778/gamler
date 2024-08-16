package service

import (
	"gambler/backend/database/models"
	"gambler/backend/database/models/customTypes"
	"gambler/backend/handlers"
	"gambler/backend/tools"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type (
	CreateBetReq struct {
		Name        string   `json:"name" validate:"required,min=3,max=50,ascii"`
		Description string   `json:"description" validate:"required,min=3,max=50,ascii"`
		BetOptions  []string `json:"betOptions" validate:"required,min=2,dive,min=3,max=50,ascii"`
	}
)

func CreateBet(c *fiber.Ctx) error {
	req := new(CreateBetReq)

	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(tools.GlobalErrorHandlerResp{
			Success: false,
			Message: "[Parser] Bad request",
			Code:    400,
		})
	}

	if errs := handlers.VHandler.Validate(req); len(errs) > 0 && errs[0].Error {
		return c.Status(400).JSON(tools.GlobalErrorHandlerResp{
			Success: false,
			Message: "[Validator] Bad request",
			Code:    400,
			Body:    errs,
		})
	}

	bet := models.Bet{
		Name:        req.Name,
		Description: req.Description,
		BetOptions:  pq.StringArray(req.BetOptions),
		Status:      customTypes.Open,
		UserBets:    []models.UserBet{},
	}

	err := handlers.DB.CreateBet(bet)
	if err != -1 {
		if err == tools.DB_DUP_KEY {
			return c.Status(409).JSON(tools.GlobalErrorHandlerResp{
				Success: false,
				Message: "User already exists",
				Code:    409,
			})
		} else {
			return c.Status(500).JSON(tools.GlobalErrorHandlerResp{
				Success: false,
				Message: "Internal server error",
				Code:    500,
			})
		}
	}

	return c.Status(201).JSON(tools.GlobalErrorHandlerResp{
		Success: true,
		Message: "Bet created",
		Code:    201,
		Body:    bet,
	})
}
