package main

import (
	"embed"
	"log"

	"nebula/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 加载配置文件
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败：%v", err)
	}

	log.Printf("启动 %s v%s", cfg.App.Name, cfg.App.Version)

	// Create an instance of the app structure
	app := NewApp()

	// 设置窗口标题为配置中的应用名称
	windowTitle := cfg.App.Name
	if windowTitle == "" {
		windowTitle = "nebula"
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  windowTitle,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
