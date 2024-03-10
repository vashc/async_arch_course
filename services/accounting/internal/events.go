package internal

import "github.com/google/uuid"

type UserCreatedIn struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
}

type TaskAssignedIn struct {
	Amount     int       `json:"amount"`
	AssigneeID uuid.UUID `json:"assignee_id"`
}

type TaskCompletedIn struct {
	Amount     int       `json:"amount"`
	AssigneeID uuid.UUID `json:"assignee_id"`
}
