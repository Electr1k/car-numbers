package domain

// Region - Регион
type Region struct {
	ID   int    `validate:"required"`       // ID - Идентификатор
	Name string `validate:"required,min=3"` // Name - Название региона
}

// RegionCode - код региона
type RegionCode struct {
	Code     string `validate:"required"`
	RegionID int    `validate:"required"`
}

type RegionWithCodes struct {
	Region      Region
	RegionCodes []RegionCode
}

// RestoreRegion - Восстановление существующего региона из хранилища
func RestoreRegion(
	id int,
	name string,
) (*Region, error) {
	return newRegion(id, name)
}

func newRegion(
	id int,
	name string,
) (*Region, error) {
	n := &Region{
		ID:   id,
		Name: name,
	}

	if err := validate.Struct(n); err != nil {
		return nil, err
	}

	return n, nil
}

// RestoreRegionCode - Восстановление существующего кода региона из хранилища
func RestoreRegionCode(
	code string,
	regionID int,
) (*RegionCode, error) {
	return newRegionCode(code, regionID)
}

func newRegionCode(
	code string,
	regionID int,
) (*RegionCode, error) {
	c := &RegionCode{
		Code:     code,
		RegionID: regionID,
	}

	if err := validate.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}
