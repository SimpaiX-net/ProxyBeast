/*
/*

     ProxyBeast GUI

The ultimate proxy checker
       by @z3ntl3

    [proxy.pix4.dev]

License: GNU
Note: Please do give us a star on Github, if you like ProxyBeast

[App core]
*/

package core

import (
	"context"
	"os"
	"path"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type (
	// App instance
	App struct {
		ctx context.Context
	}

	// Describes some operation
	Operation = string

	// Event listeners
	EventListeners struct {
		Name string
		Exec func(optionalData ...interface{})
		Cancel func()
	}

	// Event listeners aggregation
	EventGroup []*EventListeners
)

var (
	APP = New()
)

const (
	SaveFile Operation = "dialog_save_file"
	InputFile Operation = "dialog_input_file"
)

func New() *App {
	return &App{}
}

func (a *App) GetCtx() context.Context {
	return a.ctx
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	runtime.WindowCenter(ctx)

	// Obtain current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	
	// Alias CWD
	RootDir = cwd

	// If <cwd>/saves is not resolvable, then mkdir.
	if _, err := os.Stat(path.Join(cwd,"saves")); err != nil || os.IsNotExist(err){
		if err = os.Mkdir(path.Join(cwd,"saves"), os.ModeDir); err != nil {
			println(err)
		}
	} 

}

// Triggers when all resources are loaded
func (a *App) DomReady(ctx context.Context) {
	a.ctx = ctx

	if _, err := os.Stat(
		path.Join(RootDir, "saves"),
	); err != nil || os.IsNotExist(err) || RootDir == ""{
		runtime.EventsEmit(a.ctx, Fire_ErrSvdirEvent)
		return
	}
	
	MX.Register(context.WithCancel(context.Background()))
		
	MX.fd_pool = make(FD_Pool, 20)
	MX.worker_pool = make(Workers, 2000)

	var events *EventGroup = &EventGroup{
		{
			Name: OnDialog,
			Exec: a.dialog_exec,
		},{
			Name: OnStartScan,
			Exec: func(data ...interface{}) {
				defer func() {
					err := recover()
					if err != nil {
						runtime.EventsEmit(a.ctx, Fire_ErrEvent, err.(error).Error())
					}
				}()
				MX.StartChecking(a.ctx, data[0].(string))
			},
		},
	}
	events.register_eventListeners(a.ctx)
}

func(a *App) dialog(opts runtime.OpenDialogOptions) (string, error){
	return runtime.OpenFileDialog(a.ctx, opts)
}

// Registers listeners for events
func (g *EventGroup) register_eventListeners(ctx context.Context) {
	for _, event := range *g {
		event.Cancel = runtime.EventsOn(ctx, event.Name, event.Exec)
	}
}

func (a *App) dialog_exec(optionalData ...interface{}) {
	var err error
	defer func(err_ *error){
		if *err_ != nil {
			runtime.EventsEmit(a.ctx, Fire_ErrEvent, (*err_).Error())
		}
	}(&err)
	
	props, ok := optionalData[0].(string)
	if !ok {
		err = ErrPropsInvalid
		return
	}

	opts := runtime.OpenDialogOptions{
		DefaultDirectory: path.Join(RootDir),
		Title: "ProxyBeast - File dialog",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Select file (.txt)",
				Pattern: "*.txt",
			},
		},
	}
	if props == SaveFile {
		opts.DefaultDirectory = path.Join(opts.DefaultDirectory, "saves")
	}
	
	loc, err := a.dialog(opts)
	if err != nil {
		return
	}

	f, err := OpenFileRDO(loc)
	if err != nil {
		return
	}

	FD[props].Close()
	FD[props] = f

	runtime.EventsEmit(a.ctx, props, path.Base(loc))
}