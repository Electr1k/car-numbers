package autonomera

import (
	"data-service/internal/domain"
	"fmt"
)

// Section - раздел объявлений на сайте провайдера (активные/архив)
type Section string

const (
	// SectionActive - активные объявления
	SectionActive Section = "active"

	// SectionArchive - архив: снятые и проданные
	SectionArchive Section = "archive"
)

// queryValue - значение параметра blog
func (s Section) queryValue() (string, bool) {
	switch s {
	case SectionArchive:
		return "numbersarchive", true
	default:
		return "", false
	}
}

// statusForSection - в каком статусе находятся предложения из этого раздела
func statusForSection(section Section) (domain.OfferStatus, error) {
	switch section {
	case SectionActive:
		return domain.OfferStatusActive, nil
	case SectionArchive:
		return domain.OfferStatusInactive, nil
	default:
		return "", fmt.Errorf("unknown section %q", section)
	}
}
