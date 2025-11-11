package types

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // omit in JSON responses
	CreatedAt time.Time `json:"created_at"`
}
