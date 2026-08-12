package autonomera

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
