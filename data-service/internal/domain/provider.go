package domain

// Provider - Провайдер
type Provider string

const (
	// ProviderAutonomera - autonomera777.ru
	ProviderAutonomera Provider = "autonomera"
	// ProviderGosnomeru - gosnomeru.com
	ProviderGosnomeru Provider = "gosnomeru"
	// ProviderAnomera - anomera.ru
	ProviderAnomera Provider = "anomera"
)

func GetAllProviders() []Provider {
	return []Provider{
		// ProviderAutonomera, TODO: пока берем из архива
		ProviderGosnomeru,
		ProviderAnomera,
	}
}
