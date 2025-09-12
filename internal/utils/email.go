package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

// ResendEmailPayload represents the payload for the Resend API
type ResendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

// OTPTemplateData represents the data needed for the OTP email template
type OTPTemplateData struct {
	OTP      string
	AppName  string
	UserName string
}

// WelcomeTemplateData represents the data needed for the welcome email template
type WelcomeTemplateData struct {
	AppName  string
	UserName string
}

// GenerateOTP generates a cryptographically secure random 6-digit OTP
func GenerateOTP() string {
	// Generate 6 random bytes
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based generation if crypto/rand fails
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}

	// Convert to 6-digit number
	var otp int
	for i, b := range bytes {
		otp += int(b) * (1 << (8 * i))
	}
	otp = (otp % 900000) + 100000

	return fmt.Sprintf("%06d", otp)
}

// SendOTPEmail sends an OTP verification email using Resend API
func SendOTPEmail(toEmail, otp string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("EMAIL_FROM")
	fromName := os.Getenv("EMAIL_FROM_NAME")

	if apiKey == "" || fromEmail == "" {
		return fmt.Errorf("missing Resend API key or sender email in environment variables")
	}

	// If fromName is provided, format the from field
	from := fromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	// Create template data
	data := OTPTemplateData{
		OTP:      otp,
		AppName:  fromName,
		UserName: toEmail,
	}

	// Parse the email template
	tmpl, err := template.New("otp_email").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        .header {
            text-align: center;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        .content {
            padding: 20px 0;
        }
        .otp-code {
            font-size: 32px;
            font-weight: bold;
            text-align: center;
            letter-spacing: 5px;
            margin: 30px 0;
            color: #4a6ee0;
        }
        .footer {
            padding-top: 10px;
            border-top: 1px solid #eee;
            font-size: 12px;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Email Verification</h2>
        </div>
        <div class="content">
            <p>Hi there,</p>
            <p>Thank you for registering with {{.AppName}}. To complete your registration, please enter the following verification code:</p>
            <div class="otp-code">{{.OTP}}</div>
            <p>This code will expire in 10 minutes.</p>
            <p>If you didn't request this code, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; {{.AppName}}</p>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		return err
	}

	// Execute the template with our data
	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return err
	}

	// Create the Resend API payload
	payload := ResendEmailPayload{
		From:    from,
		To:      []string{toEmail},
		Subject: "Your Verification Code",
		Html:    emailBody.String(),
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create the Resend API request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("resend API error: %v", errorResponse)
		}
		return fmt.Errorf("resend API error: status code %d", resp.StatusCode)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to the newly registered user
func SendWelcomeEmail(toEmail string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("EMAIL_FROM")
	fromName := os.Getenv("EMAIL_FROM_NAME")

	if apiKey == "" || fromEmail == "" {
		return fmt.Errorf("missing Resend API key or sender email in environment variables")
	}

	// If fromName is provided, format the from field
	from := fromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	// Extract username from email
	username := toEmail
	if at := bytes.IndexByte([]byte(toEmail), '@'); at >= 0 {
		username = toEmail[:at]
	}

	// Create template data
	data := WelcomeTemplateData{
		AppName:  fromName,
		UserName: username,
	}

	// Parse the email template
	tmpl, err := template.New("welcome_email").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        .header {
            text-align: center;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        .content {
            padding: 20px 0;
        }
        .button {
            display: inline-block;
            padding: 10px 20px;
            background-color: #4a6ee0;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            padding-top: 10px;
            border-top: 1px solid #eee;
            font-size: 12px;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Welcome to {{.AppName}}!</h2>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p>Thank you for verifying your email and joining {{.AppName}}. We're excited to have you on board!</p>
            <p>With your account, you can now generate custom backend projects tailored to your needs.</p>
            <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
            <p>Happy coding!</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; {{.AppName}}</p>
        </div>
    </div>
</body>
</html>
`)
	if err != nil {
		return err
	}

	// Execute the template with our data
	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return err
	}

	// Create the Resend API payload
	payload := ResendEmailPayload{
		From:    from,
		To:      []string{toEmail},
		Subject: "Welcome to " + fromName + "!",
		Html:    emailBody.String(),
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create the Resend API request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("resend API error: %v", errorResponse)
		}
		return fmt.Errorf("resend API error: status code %d", resp.StatusCode)
	}

	return nil
}
