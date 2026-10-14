package autonomera

import (
	"errors"
	"plate-service/internal/domain"
	"plate-service/internal/provider"
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
	result, err := newTestMapper().MapOfferToDomain(defaultRow().selection(t), domain.OfferStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Plate.Number != "а123аа77" {
		t.Errorf("number = %q", result.Plate.Number)
	}
	if result.Plate.Type != domain.PlateTypeCar {
		t.Errorf("type = %q", result.Plate.Type)
	}
	if result.Offer.ExternalID != "12345" {
		t.Errorf("external id = %q, want 12345", result.Offer.ExternalID)
	}
	if result.Offer.URL != "https://autonomera777.ru/standart/а123аа77" {
		t.Errorf("url = %q", result.Offer.URL)
	}
	if result.Offer.Price == nil || *result.Offer.Price != 1200000 {
		t.Errorf("price = %v", result.Offer.Price)
	}
	if result.Offer.Status != domain.OfferStatusActive {
		t.Errorf("status = %q", result.Offer.Status)
	}
	if result.Offer.PlateID != result.Plate.ID {
		t.Errorf("offer.PlateID %s != plate.ID %s", result.Offer.PlateID, result.Plate.ID)
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
		want domain.PlateType
	}{
		{"/standart/а123аа77", domain.PlateTypeCar},
		{"/moto/а123аа77", domain.PlateTypeMoto},
		{"/trailer/а123аа77", domain.PlateTypeTrailer},
	}

	for _, c := range cases {
		fixture := defaultRow()
		fixture.href = c.href

		result, err := newTestMapper().MapOfferToDomain(fixture.selection(t), domain.OfferStatusActive)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.href, err)
		}
		if result.Plate.Type != c.want {
			t.Errorf("%s: type = %q, want %q", c.href, result.Plate.Type, c.want)
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

			_, err := newTestMapper().MapOfferToDomain(fixture.selection(t), domain.OfferStatusActive)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, provider.ErrBrokenOffer) {
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
		{"цена ноль", func(r *row) { r.price = "0 ₽" }},
		{"номер короче восьми", func(r *row) { r.title = "а123" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fixture := defaultRow()
			c.mutate(&fixture)

			_, err := newTestMapper().MapOfferToDomain(fixture.selection(t), domain.OfferStatusActive)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, provider.ErrRowSkipped) {
				t.Fatalf("want ErrRowSkipped, got %v", err)
			}
		})
	}
}

func TestMapToDomainNegotiablePrice(t *testing.T) {
	fixture := defaultRow()
	fixture.price = "Договорная"

	result, err := newTestMapper().MapOfferToDomain(fixture.selection(t), domain.OfferStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Offer.Price != nil {
		t.Errorf("price = %v, want nil (negotiable)", *result.Offer.Price)
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
			if price == nil || *price != c.want {
				t.Fatalf("price = %v, want %v", price, c.want)
			}
		})
	}
}

// profileTable - таблица деталки с блоком продавца
func profileTable(t *testing.T, rows string) *goquery.Selection {
	t.Helper()

	document, err := goquery.NewDocumentFromReader(strings.NewReader(
		`<div class="article__table user-data-table">` + rows + `</div>`))
	if err != nil {
		t.Fatalf("build fixture: %v", err)
	}

	return document.Find(".article__table.user-data-table")
}

func profileRow(title, value string) string {
	return `<div class="user-data-table__tr"><div class="user-data-table__th">` + title +
		`</div><div class="user-data-table__td">` + value + `</div></div>`
}

func TestParseProfileFromDetail(t *testing.T) {
	table := profileTable(t,
		profileRow("Цена авто с номером", `<div class="table-price">40 000</div>`)+
			profileRow("Логин", `<a href="/user?user_id=55716" class="table-user"><div class="table-user__name"> Dan&#39;ka </div></a>`)+
			profileRow("Имя", ` Владислав  `)+
			profileRow("Телефон", `<span class="user-data-table__data user-data-table__data--conf">+7 (966) 032-65-80</span>`)+
			profileRow("E-mail", `seller@mail.ru`)+
			profileRow("Рейтинг пользователя", `<span class="user-data-table__rating">30 баллов</span>`))

	profile, err := newTestMapper().parseProfileFromDetail(table)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile == nil {
		t.Fatal("expected profile, got nil")
	}

	if profile.ExternalID != "55716" || profile.URL != "https://autonomera777.ru/user?user_id=55716" {
		t.Fatalf("unexpected id/url: %s %s", profile.ExternalID, profile.URL)
	}

	if profile.Login != "Dan'ka" || profile.Name == nil || *profile.Name != "Владислав" {
		t.Fatalf("unexpected login/name: %q %v", profile.Login, profile.Name)
	}

	if profile.Phone == nil || *profile.Phone != "79660326580" {
		t.Fatalf("unexpected phone: %v", profile.Phone)
	}

	if profile.Email == nil || *profile.Email != "seller@mail.ru" {
		t.Fatalf("unexpected email: %v", profile.Email)
	}

	if strings.Contains(profile.Raw, "Цена авто с номером") || !strings.Contains(profile.Raw, "Рейтинг пользователя") {
		t.Fatalf("raw must contain only seller rows: %s", profile.Raw)
	}
}

func TestParseProfileFromDetailHidden(t *testing.T) {
	hidden := `<span class="item-notextended">Скрыт</span>`
	table := profileTable(t,
		profileRow("Логин", `<a href="javascript:;" class="table-user"><div class="table-user__name">`+hidden+`</div></a>`)+
			profileRow("Имя", hidden)+
			profileRow("Телефон", hidden)+
			profileRow("E-mail", hidden))

	profile, err := newTestMapper().parseProfileFromDetail(table)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile != nil {
		t.Fatalf("expected nil profile for hidden seller, got %+v", profile)
	}
}

func TestParseProfileFromDetailRejectsEmptyLogin(t *testing.T) {
	table := profileTable(t,
		profileRow("Логин", `<a href="/user?user_id=55716" class="table-user"><div class="table-user__name"> </div></a>`))

	if _, err := newTestMapper().parseProfileFromDetail(table); !errors.Is(err, provider.ErrRowSkipped) {
		t.Fatalf("expected ErrRowSkipped, got %v", err)
	}
}
