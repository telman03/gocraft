package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/builder"
	"github.com/telman03/gocraft-backend/internal/models"
	"github.com/telman03/gocraft-backend/internal/utils"
	"github.com/telman03/gocraft-backend/internal/validation"
)

// prepareFeatures merges the Framework field into Features (deduplicating),
// validates conflicts, and returns the adjusted feature list or an error response.
func prepareFeatures(c *fiber.Ctx, req models.GenerateRequest) ([]string, error) {
	all := req.Features
	if req.Framework != "" {
		exists := false
		for _, f := range req.Features {
			if strings.EqualFold(f, req.Framework) {
				exists = true
				break
			}
		}
		if !exists {
			all = append([]string{req.Framework}, req.Features...)
		}
	}

	v := validation.NewTemplateValidator()
	result := v.ValidateFeatures(all)
	if !result.IsValid {
		err := c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "Feature validation failed",
			"validation_result": result,
			"message":           "Please resolve the conflicts and try again",
		})
		return nil, err
	}
	return result.AdjustedFeatures, nil
}

// requestID returns the X-Request-ID header value or a timestamp-based fallback.
func requestID(c *fiber.Ctx) string {
	if id := c.Get("X-Request-ID"); id != "" {
		return id
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// sendZip sets download headers and streams the zip file.
func sendZip(c *fiber.Ctx, zipPath, projectName string) error {
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", projectName))
	c.Set("Content-Type", "application/zip")
	return c.Download(zipPath)
}

// Generate godoc
// @Summary Generate Go backend project (authenticated)
// @Description Accepts selected features and returns a downloadable zip file
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Security BearerAuth
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /generate/authenticated [post]
func Generate(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	features, err := prepareFeatures(c, req)
	if err != nil {
		return err
	}

	reqID := requestID(c)
	fmt.Printf("[AUTH-REQ:%s] user=%v project=%q features=%v\n", reqID, userID, req.ProjectName, features)

	zipPath, err := builder.GenerateProject(req.ProjectName, features)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}
	return sendZip(c, zipPath, req.ProjectName)
}

// GenerateWithOptionalAuth godoc
// @Summary Generate Go backend project (optional auth)
// @Description Generates a project for both guest and authenticated users
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /api/v1/generate [post]
func GenerateWithOptionalAuth(c *fiber.Ctx) error {
	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	features, err := prepareFeatures(c, req)
	if err != nil {
		return err
	}

	reqID := requestID(c)
	isAuth, _ := c.Locals("is_authenticated").(bool)
	if isAuth {
		fmt.Printf("[AUTH-REQ:%s] user=%v project=%q features=%v\n", reqID, c.Locals("user_id"), req.ProjectName, features)
	} else {
		fmt.Printf("[GUEST-REQ:%s] project=%q features=%v\n", reqID, req.ProjectName, features)
	}

	zipPath, err := builder.GenerateProject(req.ProjectName, features)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}
	return sendZip(c, zipPath, req.ProjectName)
}

// GeneratePublic godoc
// @Summary Generate Go backend project (public)
// @Description Generates a project for guest users; no authentication required
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /generate [post]
func GeneratePublic(c *fiber.Ctx) error {
	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	features, err := prepareFeatures(c, req)
	if err != nil {
		return err
	}

	fmt.Printf("[PUBLIC-REQ:%s] project=%q features=%v\n", requestID(c), req.ProjectName, features)

	zipPath, err := builder.GenerateProject(req.ProjectName, features)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}
	return sendZip(c, zipPath, req.ProjectName)
}
