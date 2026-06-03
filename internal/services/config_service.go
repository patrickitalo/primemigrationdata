package services

import (
	"time"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/models"
	"github.com/primesoftwaresi/prime-migration-fyne/pkg/db"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (cs *ConfigService) LoadClientConfig(codigoCliente, sistema string) (*models.ClientConfig, error) {
	// Usar LoadClientConfig diretamente (sem a lógica de console do GetClientConfig)
	clientConfig, err := db.LoadClientConfig(codigoCliente, sistema)
	if err != nil {
		return nil, err
	}
	if clientConfig == nil {
		return nil, nil
	}
	return &models.ClientConfig{
		ID:                clientConfig.ID,
		CodigoCliente:     clientConfig.CodigoCliente,
		SistemaOrigem:     clientConfig.SistemaOrigem,
		DbPath:            clientConfig.DbPath,
		DbUser:            clientConfig.DbUser,
		DbPassword:        clientConfig.DbPassword,
		DbHost:            clientConfig.DbHost,
		DbPort:            clientConfig.DbPort,
		AliasPharmacie:    clientConfig.AliasPharmacie,
		IpServerPharmacie: clientConfig.IpServerPharmacie,
		PortaPharmacie:    clientConfig.PortaPharmacie,
		Conversao:         clientConfig.Conversao,
	}, nil
}

func (cs *ConfigService) SaveClientConfig(config *models.ClientConfig) error {
	dbConfig := &db.ClientConfig{
		CodigoCliente:     config.CodigoCliente,
		SistemaOrigem:     config.SistemaOrigem,
		DbPath:            config.DbPath,
		DbUser:            config.DbUser,
		DbPassword:        config.DbPassword,
		DbHost:            config.DbHost,
		DbPort:            config.DbPort,
		AliasPharmacie:    config.AliasPharmacie,
		IpServerPharmacie: config.IpServerPharmacie,
		PortaPharmacie:    config.PortaPharmacie,
		Conversao:         config.Conversao,
		UltimaAtualizacao: time.Now().Format(time.RFC3339),
	}
	return db.SaveClientConfig(dbConfig)
}

func (cs *ConfigService) ListClients() ([]models.ClientConfig, error) {
	configs, err := db.ListClientConfigs()
	if err != nil {
		return nil, err
	}
	result := make([]models.ClientConfig, len(configs))
	for i, c := range configs {
		result[i] = models.ClientConfig{
			ID:                c.ID,
			CodigoCliente:     c.CodigoCliente,
			SistemaOrigem:     c.SistemaOrigem,
			DbPath:            c.DbPath,
			DbUser:            c.DbUser,
			DbPassword:        c.DbPassword,
			DbHost:            c.DbHost,
			DbPort:            c.DbPort,
			AliasPharmacie:    c.AliasPharmacie,
			IpServerPharmacie: c.IpServerPharmacie,
			PortaPharmacie:    c.PortaPharmacie,
			Conversao:         c.Conversao,
		}
	}
	return result, nil
}
