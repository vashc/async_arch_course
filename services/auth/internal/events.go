package internal

import "github.com/google/uuid"

type UserCreatedOut struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
}
