package data

import (
	"plate-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// ProfileMatch - признак, по которому другой аккаунт считается тем же продавцом
type ProfileMatch string

const (
	// ProfileMatchPhone - совпал телефон
	ProfileMatchPhone ProfileMatch = "phone"

	// ProfileMatchEmail - совпал email
	ProfileMatchEmail ProfileMatch = "email"

	// ProfileMatchLogin - совпал логин
	ProfileMatchLogin ProfileMatch = "login"
)

// ProfileOffers - количество предложений продавца
type ProfileOffers struct {
	Active int
	Total  int
}

// ProfileDetail - профиль продавца для выдачи
type ProfileDetail struct {
	ID       uuid.UUID
	Provider domain.Provider
	URL      string
	IconURL  *string
	Login    string
	Name     *string
	Phone    *string
	Email    *string
	Offers   ProfileOffers
	Related  []RelatedProfile
}

// RelatedProfile - аккаунт, совпавший с профилем по телефону, email или логину
type RelatedProfile struct {
	ID       uuid.UUID
	Provider domain.Provider
	URL      string
	IconURL  *string
	Login    string
	Name     *string
	Offers   ProfileOffers
	Matches  []ProfileMatch
}

// ProfileOffer - предложение продавца в списке его предложений
type ProfileOffer struct {
	ID          uuid.UUID
	PlateID     uuid.UUID
	Number      string
	Price       *float64
	Status      domain.OfferStatus
	PostedAt    time.Time
	RefreshedAt time.Time
	URL         string
}

// ProfileOffersCursor - позиция в списке предложений продавца
type ProfileOffersCursor struct {
	ID          uuid.UUID `json:"id"`
	RefreshedAt time.Time `json:"r"`
}
