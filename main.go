package main

import (
	"fmt"
	"github.com/Sqkam/systemdd/color"
	"github.com/Sqkam/systemdd/global"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
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

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

}

func run(cmdRaw, workDir string) {
	cmdString, args := splitCmd(cmdRaw)
	cmdName := filepath.Base(cmdString)
	absWorkDir, _ := filepath.Abs(workDir)
	absCwd, _ := filepath.Abs(".")

	err := os.MkdirAll("./log", 0644)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(cmdString, args...)
	cmd.Dir = absWorkDir

	logFileName := filepath.Join("./log", cmdName+".log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    20,   // A file can be up to 20M.
		MaxBackups: 5,    // Save up to 5 files at the same time.
		MaxAge:     10,   // A file can exist for a maximum of 10 days.
		Compress:   true, // Compress with gzip.
	}

	cmd.Stdout = lumberjackLogger
	cmd.Stderr = lumberjackLogger
	err = cmd.Start()
	if err != nil {
		fmt.Printf("[systemdd] err |%s %s %s|\n", color.Yellow, "start failed attempting to use default path", color.Reset)
		cmd.Dir = absCwd
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		fmt.Printf("[systemdd] success |%s %s %s|\n", color.Blue, "using default path start success", color.Reset)

	}
	fmt.Printf("[systemdd] started |%s %s is running,work_dir:%s %s|\n", color.Magenta, cmdName, cmd.Dir, color.Reset)

	err = cmd.Wait()
	var errMessage string
	if err == nil {
		errMessage = "nil"
	} else {
		errMessage = err.Error()
	}
	fmt.Printf("[systemdd] finished |%s %s %s| errorMessage: %s %s %s\n", color.Green, cmdRaw, color.Reset, color.Red, errMessage, color.Reset)
}
func splitCmd(cmdRaw string) (cmd string, args []string) {
	cmdRaw = strings.Trim(cmdRaw, " ")
	data := make([]byte, 0, len(cmdRaw))
	j := false

	for _, d := range []byte(cmdRaw) {
		if d >= 'A' && d <= 'z' || d == []byte("-")[0] || d == []byte("'")[0] || d == []byte("\"")[0] {
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
