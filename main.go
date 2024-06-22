package main

import (
	"context"
	"fmt"
	"github.com/Sqkam/systemdd/color"
	"github.com/Sqkam/systemdd/global"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
)

func main() {
	v := viper.New()
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&global.ServerConfig); err != nil {
		panic(err)
	}

	for _, val := range global.ServerConfig.Units {
		if !val.Disable {
			go run(val.Exec, val.WorkDir)
		}
	}
	//time.Sleep(time.Second * 3)
	//for _, val := range cancelFuncs {
	//	val()
	//}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

}

func closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

var cancelFuncs []func()

func run(cmdRaw, workDir string) {
	cmdString, args := splitCmd(cmdRaw)
	cmdName := filepath.Base(cmdString)
	absWorkDir, _ := filepath.Abs(workDir)
	absCwd, _ := filepath.Abs(".")

	err := os.MkdirAll("./log", 0644)
	if err != nil {
		fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, err.Error(), color.Reset)
		return
	}
	runningCtx, cancelFunc := context.WithCancel(context.Background())

	cancelFuncs = append(cancelFuncs, cancelFunc)

	logFileName := filepath.Join("./log", cmdName+".log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    20,   // A file can be up to 20M.
		MaxBackups: 5,    // Save up to 5 files at the same time.
		MaxAge:     10,   // A file can exist for a maximum of 10 days.
		Compress:   true, // Compress with gzip.
	}

	started := false
	childFiles := make([]*os.File, 0, 3)
	var childIOFiles []io.Closer
	var parentIOPipes []io.Closer
	var goroutine []func() error

	defer func() {
		closeDescriptors(childIOFiles)
		childIOFiles = nil

		if !started {
			closeDescriptors(parentIOPipes)
			parentIOPipes = nil
		}
	}()
	// input
	fileInput, err := os.Open(os.DevNull)
	if err != nil {
		fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, err.Error(), color.Reset)
		return
	}
	//stdin
	childFiles = append(childFiles, fileInput)
	childIOFiles = append(childIOFiles, fileInput)

	pr, pw, err := os.Pipe()
	if err != nil {
		fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, err.Error(), color.Reset)
		return
	}
	// pw is stdout
	//stout
	childFiles = append(childFiles, pw)
	childIOFiles = append(childIOFiles, pw)
	parentIOPipes = append(parentIOPipes, pr)
	//stderr
	childFiles = append(childFiles, pw)

	goroutine = append(goroutine, func() error {
		_, err := io.Copy(lumberjackLogger, pr)
		pr.Close() // in case io.Copy stopped due to write error
		return err
	})
	lp := cmdString
	if !filepath.IsAbs(cmdString) {
		lp, err = exec.LookPath(cmdString)
		if err != nil {
			fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, err.Error(), color.Reset)
			return
		}
	}
	var process *os.Process
	var runningDir string
	runningDir = absWorkDir

	if workDir == "" {
		fmt.Printf("[systemdd] err |%s %s: %s %s|\n", color.Yellow, cmdString, "workDir empty , attempting to use cmd path", color.Reset)
		runningDir = filepath.Dir(lp)
	}

	process, err = os.StartProcess(lp, append([]string{cmdString}, args...), &os.ProcAttr{
		Dir:   runningDir,
		Files: childFiles,
		//Env:   env,
		//Sys:   c.SysProcAttr,
	})
	if err != nil {
		//auto run with absCwd
		fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, "start failed , attempting to use absCwd path", color.Reset)
		runningDir = absCwd
		process, err = os.StartProcess(lp, append([]string{cmdString}, args...), &os.ProcAttr{
			Dir:   runningDir,
			Files: childFiles,
			//Env:   env,
			//Sys:   c.SysProcAttr,
		})
		if err != nil {
			fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, err.Error(), color.Reset)
			return
		}
	}
	fmt.Printf("[systemdd] started |%s %s is running,work_dir:%s %s|\n", color.Magenta, cmdName, runningDir, color.Reset)

	started = true
	for _, fn := range goroutine {
		go fn()
	}
	finishedCtx, finishedFunc := context.WithCancel(context.Background())
	goroutine = nil // Allow the goroutines' closures to be GC'd when they complete.
	go func(runningCtx context.Context, finishedCtx context.Context, p *os.Process) {
		select {
		case <-runningCtx.Done():
			fmt.Printf("[systemdd] killing |%s %s %s|\n", color.Cyan, cmdName, color.Reset)
			p.Kill()
		case <-finishedCtx.Done():

		}
	}(runningCtx, finishedCtx, process)
	state, err := process.Wait()

	var errMessage string
	if err == nil {
		errMessage = "nil"
	} else {
		errMessage = err.Error()
	}
	finishedFunc()
	fmt.Printf("[systemdd] finished |%s %s %s| state:%s %s %s ,errorMessage: %s %s %s\n", color.Green, cmdRaw, color.Reset, color.Red, state.String(), color.Reset, color.Red, errMessage, color.Reset)
}
func splitCmd(cmdRaw string) (cmd string, args []string) {
	cmdRaw = strings.Trim(cmdRaw, " ")
	data := make([]byte, 0, len(cmdRaw))
	j := false

	for _, d := range []byte(cmdRaw) {
		if d >= 'A' && d <= 'z' || d == []byte("-")[0] || d == []byte("'")[0] || d == []byte("\"")[0] || d == []byte("/")[0] {
			if j {
				data = append(data, ' ')

				j = false
			}
			data = append(data, d)

		} else {
			j = true
		}
	}
	split := strings.Split(string(data), " ")
	return split[0], split[1:]
}
