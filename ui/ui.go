package ui

import (
	"embed"
	"fmt"
	"github.com/zserge/lorca"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"time"
	"tsetmc/utils/logger"
	"tsetmc/utils/settings"
)

//go:embed www
var fs embed.FS
var ui lorca.UI
var uiFilled bool

func UiInit() (lorca.UI, net.Listener) {
	args := []string{}
	conf := settings.LoadConfiguration()
	width := conf.Width
	height := conf.Height
	if runtime.GOOS == "windows" {
		fmt.Printf("%dx%d\n", width, height)
	} else {
		args = append(args, "--class=Lorca")
	}
	args = append(args, "--force-devtools-available")
	args = append(args, "--disable-web-security")
	args = append(args, "--remote-allow-origins=*")
	args = append(args, "--Log.enable false")
	args = append(args, "--viewPort=\"maximized\"")

	var err error
	ui, err = makeWindowsUI(width, height, args)
	if err != nil {
		panic(fmt.Errorf("failed to create lorca UI: %w", err))
	}
	ui.Bind("start", func() {
		fmt.Println("UI is ready")
	})

	ln, _ := makeListener()
	ln2, _ := makeListener()
	go http.Serve(ln, http.FileServer(http.FS(fs)))
	fserver := http.FileServer(http.Dir("./sound"))
	go http.Serve(ln2, fserver)
	fmt.Println("---------------**")
	fmt.Println(fmt.Sprintf("http://%s/www", ln.Addr()))
	fmt.Println(fmt.Sprintf("http://%s/astane.mp3", ln2.Addr()))
	ui.Load(fmt.Sprintf("http://%s/www", ln.Addr()))

	ui.Bind("goSaveResolution", goSaveResolution)
	ui.Bind("loadForce", loadForce)
	ui.Bind("loadBlock", loadBlock)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			func() {
				defer func() {
					if r := recover(); r != nil {

					}
				}()
				ui.Eval("window.sound = '" + ln2.Addr().String() + "';")
			}()
		}
	}()
	go func() {
		time.Sleep(10 * time.Second)
		ui.Eval("saveResolution();")
	}()

	return ui, ln

}

func SetBind(name string, fnc interface{}) {
	for uiFilled == false {
		if uiFilled {
			break
		}
	}
	ui.Bind(name, fnc)
}

func goSaveResolution(width int, height int) {
	if reflect.TypeOf(width).Name() != "int" {
		return
	}
	if reflect.TypeOf(height).Name() != "int" {
		return
	}
	settings.ChangeLoadedConfig("Width", width)
	settings.ChangeLoadedConfig("Height", height)
	settings.SaveLoadedConfiguration()

}

func loadForce() {
	fmt.Println("loadForce")
	settings.LoadForce()
}

func loadBlock() {
	fmt.Println("LoadBlock")
	settings.LoadBlock()
}

func EvalJS(js string) {
	// You may use console.log to debug your JS code, it will be printed via
	// log.Println(). Also exceptions are printed in a similar manner.
	ui.Eval(js)
}

func makeWindowsUI(width int, height int, args []string) (lorca.UI, error) {
	dir, _ := ioutil.TempDir("", "drmoo19")
	ui, err := lorca.New("", dir, width, height, args...)
	if err != nil {
		logger.GetFlog().Err(err).Msg("failed to create lorca UI")
		return nil, err
	}
	uiFilled = true
	return ui, err
}

func makeListener() (ln net.Listener, err error) {
	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.GetFlog().Err(err)
	}
	return ln, err

}

func WaitUntilUiClosed() {
	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	defer signal.Stop(sigc)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	logger.GetFlog().Info().Msg("exiting...")
}

//https://stackoverflow.com/a/48187712/4650634
//var (
//	user32           = syscall.NewLazyDLL("User32.dll")
//	getSystemMetrics = user32.NewProc("GetSystemMetrics")
//)
//
//func GetSystemMetrics(nIndex int) int {
//	index := uintptr(nIndex)
//	ret, _, _ := getSystemMetrics.Call(index)
//	return int(ret)
//}
