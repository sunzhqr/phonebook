package service

import "time"

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type PhoneIn struct {
	Label     string `validate:"max=40"`
	PhoneRaw  string `validate:"required,min=5,max=32"`
	IsPrimary bool
}

type ContactCreateIn struct {
	FirstName string    `validate:"required,min=1,max=40"`
	LastName  string    `validate:"required,min=1,max=40"`
	Company   string    `validate:"max=40"`
	Phones    []PhoneIn `validate:"required,min=1,dive"`
}

type ContactUpdateIn struct {
	FirstName *string    `validate:"omitempty,min=1,max=40"`
	LastName  *string    `validate:"omitempty,min=1,max=40"`
	Company   *string    `validate:"omitempty,max=40"`
	Phones    *[]PhoneIn `validate:"omitempty,dive"`
}
type ListFilter struct {
	FirstName string
	LastName  string
	Company   string
	Phone     string
	AfterID   int64
	Limit     int
	Sort      string
	Order     string
}

type PhoneOut struct {
	Label     string `json:"label"`
	PhoneRaw  string `json:"phone_raw"`
	PhoneE164 string `json:"phone_e164"`
	IsPrimary bool   `json:"is_primary"`
}

type ContactOut struct {
	ID        int64      `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Company   string     `json:"company"`
	Phones    []PhoneOut `json:"phones"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PageOut struct {
	NextAfterID int64 `json:"next_after_id"`
	HasMore     bool  `json:"has_more"`
	Limit       int   `json:"limit"`
}

type ListOut struct {
	Items []ContactOut `json:"items"`
	Page  PageOut      `json:"page"`
}
