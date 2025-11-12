package main

import (
	"context"
	"embed"
	"fmt"
	"go-proxy/common"
	"go-proxy/proxyserver"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var wailsContext *context.Context

func main() {
	// fmt.Println("Hey skipped Wails app. Remember to turn it back on")
	// a2 := NewApp()
	// a2.startup(context.TODO())

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "go-proxy",
		Width:  1224,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "goproxy-db2d80a7-7116-4ed5-a61c-b641cf3d1b8c",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []any{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	wailsContext = &ctx
	a.ctx = ctx

	ListenerServerManager.AddListeners([]*LocalListener{
		NewLocalListener(8000, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"ssh"}}),
		NewLocalListener(8001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"http"}}),
		NewLocalListener(8002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{Tags: []string{"socks5"}}),

		NewLocalListener(9001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{IgnoreAll: true}),
		NewLocalListener(9002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}, ServerFilter{IgnoreAll: true}),
	})

	var wg sync.WaitGroup
	wg.Go(func() {
		ListenerServerManager.Serve()
	})
	<-time.After(time.Second)

	ListenerServerManager.AddServers([]*proxyserver.Server{
		proxyserver.NewServer("localhost", 2222, &common.ProxyAuth{
			Username: "ubuntu",
			Password: "ubuntu",
		}),
		proxyserver.NewServer("127.0.0.1", 9001, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}),
		proxyserver.NewServer("::1", 9002, &common.ProxyAuth{
			Username: "khanh",
			Password: "khanh",
		}),
	})

	for _, s := range ListenerServerManager.Servers {
		fmt.Println(s)
	}

	wg.Wait()
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs := secondInstanceData.Args

	println("user opened second instance", strings.Join(secondInstanceData.Args, ","))
	println("user opened second from", secondInstanceData.WorkingDirectory)
	runtime.WindowUnminimise(*wailsContext)
	runtime.Show(*wailsContext)
	go runtime.EventsEmit(*wailsContext, "launchArgs", secondInstanceArgs)
}
