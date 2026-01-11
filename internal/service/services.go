package service

import (
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// Services holds all service instances used by the application
type Services struct {
	Entry  *EntryService
	Timer  *TimerService
	Report *ReportService
	Search *SearchService
	Stats  *StatsService
	Config *ConfigService
}

// NewServices creates a new Services instance with default paths
func NewServices() (*Services, error) {
	storagePath, err := storage.GetStoragePath()
	if err != nil {
		return nil, err
	}

	timerPath, err := timer.GetTimerPath()
	if err != nil {
		return nil, err
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		return nil, err
	}

	return NewServicesWithPaths(storagePath, timerPath, configPath, cfg), nil
}

// NewServicesWithPaths creates a new Services instance with custom paths (useful for testing)
func NewServicesWithPaths(storagePath, timerPath, configPath string, cfg config.Config) *Services {
	entryService := NewEntryService(storagePath, cfg)
	timerService := NewTimerService(timerPath, storagePath, cfg)
	reportService := NewReportService(storagePath, cfg)
	searchService := NewSearchService(storagePath, cfg)
	statsService := NewStatsService(storagePath, cfg)
	configService := NewConfigService(configPath, cfg)

	return &Services{
		Entry:  entryService,
		Timer:  timerService,
		Report: reportService,
		Search: searchService,
		Stats:  statsService,
		Config: configService,
	}
}
