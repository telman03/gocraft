package handlers

import (
	"archive/zip"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/builder"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/utils"
)

// VerifyGeneration godoc
// @Summary Verify project generation and list contents
// @Description Generates a project and returns the list of files that would be included
// @Tags Generator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /verify [post]
func VerifyGeneration(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Generate project
	zipPath, err := builder.GenerateProject(req.ProjectName, req.Features)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}

	// Open and inspect the zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to read generated project")
	}
	defer reader.Close()

	// Collect file information
	files := make([]map[string]interface{}, 0)
	var totalSize int64

	for _, file := range reader.File {
		fileInfo := map[string]interface{}{
			"name":         file.Name,
			"size":         file.UncompressedSize64,
			"is_hidden":    file.Name[0] == '.',
			"is_config":    file.Name == ".env.example" || file.Name == ".gitignore" || file.Name == "go.mod",
			"modified":     file.Modified,
		}
		files = append(files, fileInfo)
		totalSize += int64(file.UncompressedSize64)
	}

	response := map[string]interface{}{
		"project_name":    req.ProjectName,
		"features":        req.Features,
		"total_files":     len(files),
		"total_size":      totalSize,
		"files":           files,
		"zip_path":        zipPath,
		"download_ready":  true,
		"important_files": map[string]bool{
			"has_env_example": containsFile(files, ".env.example"),
			"has_gitignore":   containsFile(files, ".gitignore"),
			"has_go_mod":      containsFile(files, "go.mod"),
			"has_main_go":     containsFile(files, "main.go"),
		},
	}

	return c.JSON(response)
}

func containsFile(files []map[string]interface{}, filename string) bool {
	for _, file := range files {
		if file["name"] == filename {
			return true
		}
	}
	return false
}