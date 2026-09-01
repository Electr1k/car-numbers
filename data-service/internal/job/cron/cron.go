package cron

import (
	"data-service/config"
	"data-service/internal/job"
	"data-service/internal/scheduler"
	"fmt"
	"log/slog"
)

// Deps - зависимости кронов
type Deps struct {
	Producer   *job.Producer
	Specs      config.CronConfig
	AutoNomera config.AutoNomeraConfig
	Gosnomeru  config.GosnomeruConfig
	Anomera    config.AnomeraConfig
	Logger     *slog.Logger
}

func Register(s *scheduler.Scheduler, d Deps) error {
	err := register(s, d.Logger,
		nameImportAutonomeraActiveOffers,
		d.Specs.ImportAutonomeraActiveOffers,
		&importAutonomeraActiveOffers{
			producer: d.Producer,
			depth:    d.AutoNomera.ImportDepth,
			logger:   d.Logger.With("cron", nameImportAutonomeraActiveOffers),
		},
	)
	if err != nil {
		return fmt.Errorf("register cron %s: %w", nameImportAutonomeraActiveOffers, err)
	}

	err = register(s, d.Logger,
		nameImportAutonomeraArchiveOffers,
		d.Specs.ImportAutonomeraArchiveOffers,
		&importAutonomeraArchiveOffers{
			producer: d.Producer,
			depth:    d.AutoNomera.ImportDepth,
			logger:   d.Logger.With("cron", nameImportAutonomeraArchiveOffers),
		},
	)
	if err != nil {
		return fmt.Errorf("register cron %s: %w", nameImportAutonomeraArchiveOffers, err)
	}

	err = register(s, d.Logger,
		nameImportGosnomeruOffers,
		d.Specs.ImportGosnomeruOffers,
		&importGosnomeruOffers{
			producer: d.Producer,
			depth:    d.Gosnomeru.ImportDepth,
			logger:   d.Logger.With("cron", nameImportGosnomeruOffers),
		},
	)
	if err != nil {
		return fmt.Errorf("register cron %s: %w", nameImportGosnomeruOffers, err)
	}

	err = register(s, d.Logger,
		nameImportAnomeraOffers,
		d.Specs.ImportAnomeraOffers,
		&importAnomeraOffers{
			producer: d.Producer,
			depth:    d.Anomera.ImportDepth,
			logger:   d.Logger.With("cron", nameImportAnomeraOffers),
		},
	)
	if err != nil {
		return fmt.Errorf("register cron %s: %w", nameImportAnomeraOffers, err)
	}

	return nil
}

// register - пропускает крон с пустой спекой, такой крон выключен
func register(s *scheduler.Scheduler, logger *slog.Logger, name string, spec string, task scheduler.Task) error {
	if spec == "" {
		logger.Info("cron disabled", "cron", name)

		return nil
	}

	if err := s.Add(name, spec, task); err != nil {
		return fmt.Errorf("register cron: %w", err)
	}

	return nil
}
