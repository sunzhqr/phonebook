package handler

type PhoneDTO struct {
	Label     string `json:"label"`
	PhoneRaw  string `json:"phone_raw"`
	IsPrimary bool   `json:"is_primary"`
}

type ContactCreateDTO struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Company   string     `json:"company"`
	Phones    []PhoneDTO `json:"phones"`
}

type ContactUpdateDTO struct {
	FirstName *string     `json:"first_name"`
	LastName  *string     `json:"last_name"`
	Company   *string     `json:"company"`
	Phones    *[]PhoneDTO `json:"phones"`
}
