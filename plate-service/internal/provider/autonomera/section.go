package autonomera

import (
	"fmt"
	"plate-service/internal/domain"
)

// Section - раздел объявлений на сайте провайдера (активные/архив)
type Section string

const (
	// SectionActive - активные объявления
	SectionActive Section = "active"

	// SectionArchive - архив: снятые и проданные
	SectionArchive Section = "archive"
)

// blog - значение параметра blog
func (s Section) blog() string {
	if s == SectionArchive {
		return "numbersarchive"
	}

	return "numbers"
}

// pagePath - страница раздела, с которой браузер шлёт ajax
func (s Section) pagePath() string {
	if s == SectionArchive {
		return "/arkhiv"
	}

	return "/"
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
