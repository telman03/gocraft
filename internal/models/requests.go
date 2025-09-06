package models

type RegisterInput struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type LoginInput struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type GenerateRequest struct {
	ProjectName string   `json:"projectName" example:"my-awesome-app"`
	Features    []string `json:"features"`
}
