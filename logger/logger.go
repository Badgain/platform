package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Badgain/platform/utils"
)

const (
	MAX_LOGGER_FILE_SIZE = 2000000
	LOG_FILE_TEMPLATE    = "%s/log_%d-%s-%d-%d-%d-%d.txt"
)

type Logger struct {
	size           int64
	file           *os.File
	readStream     chan []byte
	availableSpace int64
	logPath        string
}

func (l *Logger) NewFile() error {
	l.file.Close()
	now := time.Now()
	name := fmt.Sprintf(LOG_FILE_TEMPLATE, l.logPath, now.Year(), now.Month().String(), now.Day(), now.Hour(), now.Minute(), now.Second())

	var err error
	if l.file, err = utils.OpenOrCreateFile(name); err != nil {
		l.file = nil
		return err
	}
	l.size = 0
	l.availableSpace = MAX_LOGGER_FILE_SIZE
	return nil
}

func (l *Logger) New(logPath string) error {
	var err error

	l.size = 0
	l.availableSpace = MAX_LOGGER_FILE_SIZE
	readStream := make(chan []byte)
	l.readStream = readStream

	currentDir, err := os.ReadDir(l.logPath)
	if err != nil {
		return err
	}
	filename := ""
	if len(currentDir) > 0 {
		fileinfo, err := currentDir[len(currentDir)-1].Info()
		if err == nil && float64(fileinfo.Size()) < float64(MAX_LOGGER_FILE_SIZE*0.95) {
			filename = l.logPath + "/" + fileinfo.Name()
		}
	}
	if filename == "" {
		now := time.Now()
		filename = fmt.Sprintf(LOG_FILE_TEMPLATE, l.logPath, now.Year(), now.Month().String(), now.Day(), now.Hour(), now.Minute(), now.Second())
	}

	if l.file, err = utils.OpenOrCreateFile(filename); err != nil {
		l.file = nil
		return err
	}
	return nil
}

func (l *Logger) Log(data []byte, api string) error {
	RawLog := LogStructUnit{}
	RawLog.Date = time.Now().String()
	RawLog.Message = string(data)
	RawLog.Api = api
	log, err := json.Marshal(&RawLog)
	if err != nil {
		return err
	}
	l.readStream <- log
	return nil
}

func (l *Logger) LogHard(data []byte) {
	l.readStream <- data
}

func (l *Logger) WriteLog(data []byte) error {
	if l.availableSpace <= int64(len(data)) {
		err := l.NewFile()
		return err
	}
	n, err := l.file.Write(data)
	if err != nil {
		return err
	}
	l.size += int64(n)
	l.availableSpace -= int64(n)
	return nil
}

func (l *Logger) Terminate() {
	l.size = 0
	l.availableSpace = 0
	l.file.Close()
}

type LogStructUnit struct {
	Date    string `json:"Date"`
	Api     string `json:"Api"`
	Message string `json:"Message"`
}

func InitLogger(sig chan bool, path string, lg *Logger, hardWrite bool) error {
	fmt.Println("Init Logger...")
	err := lg.New(path)
	if err != nil {
		lg.Terminate()
		return err
	}

	var LoggerError error
	var logFileData []byte
	sig <- true
	for {
		data, ok := <-lg.readStream
		if !ok {
			break
		}
		if !hardWrite {
			logFileData, LoggerError = json.Marshal(&LogStructUnit{
				Date:    time.Now().String(),
				Message: string(data),
			})
			if LoggerError != nil {
				logFileData, _ = json.Marshal(&LogStructUnit{
					Date:    time.Now().String(),
					Message: err.Error(),
				})
			}
			LoggerError = lg.WriteLog(logFileData)
			if LoggerError != nil {
				lg.Terminate()
				return LoggerError
			}
		} else {
			LoggerError = lg.WriteLog(data)
			if LoggerError != nil {
				lg.Terminate()
				return LoggerError
			}
		}

	}
	logFileData, _ = json.Marshal(&LogStructUnit{
		Date:    time.Now().String(),
		Message: "Logger is terminating...",
	})
	LoggerError = lg.WriteLog(logFileData)
	if LoggerError != nil {
		return LoggerError
	}
	lg.Terminate()
	return LoggerError
}

func (l *Logger) DeleteEmptyLogs() {
	files, _ := os.ReadDir(l.logPath)
	for _, file := range files {
		fileinfo, _ := file.Info()
		if fileinfo.Size() == 0 {
			_ = os.Remove(l.logPath + "/" + file.Name())

		}
	}
}
