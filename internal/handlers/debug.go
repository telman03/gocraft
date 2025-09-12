package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// DebugRequest godoc
// @Summary Debug request tracking
// @Description Logs request details to help debug duplicate requests
// @Tags Debug
// @Accept json
// @Produce json
// @Router /debug/request [post]
func DebugRequest(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	userAgent := c.Get("User-Agent")
	referer := c.Get("Referer")
	origin := c.Get("Origin")
	
	// Log request details
	fmt.Printf("[DEBUG:%s] Method: %s, Path: %s, IP: %s, UserAgent: %s, Referer: %s, Origin: %s\n", 
		requestID, c.Method(), c.Path(), c.IP(), userAgent, referer, origin)
	
	// Get request body if any
	body := c.Body()
	if len(body) > 0 {
		fmt.Printf("[DEBUG:%s] Body: %s\n", requestID, string(body))
	}
	
	return c.JSON(fiber.Map{
		"request_id": requestID,
		"timestamp": time.Now().UTC(),
		"method": c.Method(),
		"path": c.Path(),
		"ip": c.IP(),
		"user_agent": userAgent,
		"headers": c.GetReqHeaders(),
		"message": "Request logged for debugging",
	})
}

// DebugDownload godoc
// @Summary Debug download tracking
// @Description Tracks download requests to identify duplicates
// @Tags Debug
// @Accept json
// @Produce json
// @Router /debug/download [post]
func DebugDownload(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		body = map[string]interface{}{"error": "Could not parse body"}
	}
	
	fmt.Printf("[DOWNLOAD-DEBUG:%s] Download request received: %+v\n", requestID, body)
	
	return c.JSON(fiber.Map{
		"request_id": requestID,
		"timestamp": time.Now().UTC(),
		"body": body,
		"message": "Download request logged - check server logs for duplicates",
		"note": "If you see this message twice quickly, you have a duplicate request issue",
	})
}