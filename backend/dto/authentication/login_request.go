package authentication

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required,max=254"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
}
