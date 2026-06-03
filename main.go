package main

import (
	"embed"
	"flag"
	"log"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/app"
	"github.com/primesoftwaresi/prime-migration-fyne/internal/config"
	"github.com/primesoftwaresi/prime-migration-fyne/internal/services"
	"github.com/primesoftwaresi/prime-migration-fyne/pkg/db"
	"github.com/primesoftwaresi/prime-migration-fyne/pkg/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	config.LoadDotEnv()

	historyDSNFlag := flag.String("history-dsn", "", "DSN do Firebird central de histórico")
	flag.Parse()

	logger.SetLevel(logger.INFO)
	logger.ConfigureFileFromEnv()

	if err := db.InitConfigDB(); err != nil {
		log.Fatalf("Erro ao inicializar banco de configurações: %v", err)
	}
	defer db.CloseConfigDB()

	historyService := services.NewHistoryService()
	dsn := config.LoadHistoryDSN(*historyDSNFlag)
	if dsn != "" {
		if err := historyService.Connect(dsn); err != nil {
			log.Printf("Histórico central indisponível: %v", err)
		} else {
			log.Printf("Histórico central conectado.")
		}
	}

	migrationService := services.NewMigrationService(historyService)
	configService := services.NewConfigService()
	authService := services.NewAuthService()

	application := app.NewApp(authService, configService, historyService, migrationService, dsn)

	if err := wails.Run(&options.App{
		Title:     "Prime Migration",
		Width:     1100,
		Height:    750,
		MinWidth:  900,
		MinHeight: 620,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 255},
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
		Bind:             []interface{}{application},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	}); err != nil {
		log.Fatalf("Erro ao iniciar aplicação: %v", err)
	}
}
