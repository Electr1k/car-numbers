package fetchprofile

import "github.com/google/uuid"

// Result - профиль продавца
type Result struct {
	ID       uuid.UUID
	Provider string
	URL      string
	IconURL  *string
	Login    string
	Name     *string
	Phone    *string
	Email    *string
	Offers   Offers
	Related  []Related // Related - вероятно, другие аккаунты того же продавца
}

// Offers - количество предложений продавца
type Offers struct {
	Active int
	Total  int
}

// Related - аккаунт, совпавший по телефону, email или логину
type Related struct {
	ID       uuid.UUID
	Provider string
	URL      string
	IconURL  *string
	Login    string
	Name     *string
	Offers   Offers
	Matches  []string
}
