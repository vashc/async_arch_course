package internal

import "github.com/google/uuid"

type UserCreatedIn struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
}

type TaskCreatedOut struct {
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	AssigneeID  uuid.UUID  `json:"assignee_id"`
}

type TaskAssignedOut struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
}

type TaskCompletedOut struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
}
