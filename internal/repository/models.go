package repository

import "time"

type Contact struct {
	ID        int64
	FirstName string
	LastName  string
	Company   string
	Phones    []Phone
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Phone struct {
	Label       string
	PhoneRaw    string
	PhoneE164   string
	PhoneDigits string
	IsPrimary   bool
}

type PhoneInput struct {
	Label       string
	PhoneRaw    string
	PhoneE164   string
	PhoneDigits string
	IsPrimary   bool
}

type ContactInput struct {
	FirstName string
	LastName  string
	Company   string
	Phones    []PhoneInput
}

type ContactPatch struct {
	FirstName *string
	LastName  *string
	Company   *string
	Phones    *[]PhoneInput
}

type ListFilter struct {
	FirstName string
	LastName  string
	Company   string
	Phone     string
	AfterID   int64
	Limit     int
	SortBy    string
	Order     string
}
