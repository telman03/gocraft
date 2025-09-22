package models

type RegisterInput struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8,max=128,safe_string" example:"password123"`
}

type VerifyOTPInput struct {
	Email string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	OTP   string `json:"otp" validate:"required,len=6,numeric" example:"123456"`
}

type ResendOTPInput struct {
	Email string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=1,max=128,safe_string" example:"password123"`
}

type GenerateRequest struct {
	ProjectName string   `json:"projectName" validate:"required,min=1,max=50,project_name,safe_string" example:"my-awesome-app"`
	Framework   string   `json:"framework,omitempty" validate:"oneof=gin echo fiber" example:"gin"`
	Features    []string `json:"features" validate:"required,min=1,dive,safe_string"`
}
