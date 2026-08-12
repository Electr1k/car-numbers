package autonomera

import (
	"data-service/internal/domain"
	"data-service/internal/providers"
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// row - строка выдачи в том виде, в каком её отдаёт провайдер
type row struct {
	id    string
	href  string
	title string
	date  string
	price string
}

func defaultRow() row {
	return row{
		id:    "item_number_id12345",
		href:  "/standart/а123аа77",
		title: "а123аа77",
		date:  "05.08.2026",
		price: "1 200 000 ₽",
	}
}

// html - разметка одной строки выдачи
func (r row) html() string {
	var attrs strings.Builder
	for name, value := range map[string]string{"id": r.id, "href": r.href, "title": r.title} {
		if value != "" {
			attrs.WriteString(` ` + name + `="` + value + `"`)
		}
	}

	return `<a class="table__tr--td"` + attrs.String() + `>` +
		`<span class="table-date"><span>` + r.date + `</span></span>` +
		`<span class="table-price">` + r.price + `</span>` +
		`</a>`
}

func (r row) selection(t *testing.T) *goquery.Selection {
	t.Helper()

	document, err := goquery.NewDocumentFromReader(strings.NewReader(`<table><tr>` + r.html() + `</tr></table>`))
	if err != nil {
		t.Fatalf("build fixture: %v", err)
	}

	selection := document.Find(offerRowSelector)
	if selection.Length() != 1 {
		t.Fatalf("fixture matched %d rows, want 1", selection.Length())
	}

	return selection
}

func newTestMapper() *Mapper {
	return NewMapper("https://autonomera777.ru/")
}

func TestMapToDomainHappyPath(t *testing.T) {
	result, err := newTestMapper().MapToDomain(defaultRow().selection(t), domain.OfferStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Number.Number != "а123аа77" {
		t.Errorf("number = %q", result.Number.Number)
	}
	if result.Number.Type != domain.NumberTypeCar {
		t.Errorf("type = %q", result.Number.Type)
	}
	if result.Offer.ExternalId != "12345" {
		t.Errorf("external id = %q, want 12345", result.Offer.ExternalId)
	}
	if result.Offer.Url != "https://autonomera777.ru/standart/а123аа77" {
		t.Errorf("url = %q", result.Offer.Url)
	}
	if result.Offer.Price != 1200000 {
		t.Errorf("price = %v", result.Offer.Price)
	}
	if result.Offer.Status != domain.OfferStatusActive {
		t.Errorf("status = %q", result.Offer.Status)
	}
	if result.Offer.NumberId != result.Number.Id {
		t.Errorf("offer.NumberId %s != number.Id %s", result.Offer.NumberId, result.Number.Id)
	}
	if result.Offer.PostedAt == nil || result.Offer.PostedAt.Format(dateLayout) != "05.08.2026" {
		t.Errorf("postedAt = %v", result.Offer.PostedAt)
	}
	if !strings.Contains(result.Offer.Raw, "table-price") {
		t.Errorf("raw does not look like row html: %q", result.Offer.Raw)
	}
}

func TestMapToDomainVehicleTypes(t *testing.T) {
	cases := []struct {
		href string
		want domain.NumberType
	}{
		{"/standart/а123аа77", domain.NumberTypeCar},
		{"/moto/а123аа77", domain.NumberTypeMoto},
		{"/trailer/а123аа77", domain.NumberTypeTrailer},
	}

	for _, c := range cases {
		fixture := defaultRow()
		fixture.href = c.href

		result, err := newTestMapper().MapToDomain(fixture.selection(t), domain.OfferStatusActive)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.href, err)
		}
		if result.Number.Type != c.want {
			t.Errorf("%s: type = %q, want %q", c.href, result.Number.Type, c.want)
		}
	}
}

func TestMapToDomainBrokenOffer(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*row)
	}{
		{"нет id", func(r *row) { r.id = "" }},
		{"нет href", func(r *row) { r.href = "" }},
		{"нет title", func(r *row) { r.title = "" }},
		{"id без префикса", func(r *row) { r.id = "number_42" }},
		{"пустой id после префикса", func(r *row) { r.id = "item_number_id" }},
		{"неизвестный тип ТС", func(r *row) { r.href = "/quadro/а123аа77" }},
		{"пустой href-сегмент", func(r *row) { r.href = "/" }},
		{"пустая дата", func(r *row) { r.date = "" }},
		{"дата не по формату", func(r *row) { r.date = "2026-08-05" }},
		{"пустая цена", func(r *row) { r.price = "" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fixture := defaultRow()
			c.mutate(&fixture)

			_, err := newTestMapper().MapToDomain(fixture.selection(t), domain.OfferStatusActive)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, providers.ErrBrokenOffer) {
				t.Fatalf("want ErrBrokenOffer, got %v", err)
			}
		})
	}
}

func TestMapToDomainRowSkipped(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*row)
	}{
		{"договорная цена", func(r *row) { r.price = "Договорная" }},
		{"цена ноль", func(r *row) { r.price = "0 ₽" }},
		{"номер короче восьми", func(r *row) { r.title = "а123" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fixture := defaultRow()
			c.mutate(&fixture)

			_, err := newTestMapper().MapToDomain(fixture.selection(t), domain.OfferStatusActive)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, providers.ErrRowSkipped) {
				t.Fatalf("want ErrRowSkipped, got %v", err)
			}
		})
	}
}

func TestParsePriceSeparators(t *testing.T) {
	cases := []struct {
		name string
		text string
		want float64
	}{
		{"обычный пробел", "1 200 000 ₽", 1200000},
		{"неразрывный", "1 200 000 ₽", 1200000},
		{"тонкий", "1 200 000 ₽", 1200000},
		{"узкий неразрывный", "1 200 000 ₽", 1200000},
		{"без разделителей", "1200000₽", 1200000},
		// float32 округлил бы это значение, float64 держит точно
		{"дорогой номер", "25 400 001 ₽", 25400001},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			price, err := parsePrice(c.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if price != c.want {
				t.Fatalf("price = %v, want %v", price, c.want)
			}
		})
	}
}
