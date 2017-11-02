package logger

import (
	"os"
	"log"
	"time"
	"strings"
	"config"
)

type LogInterface interface {
	Create() error
	Write(...string)
	WriteError(err error)
}

type FileLog struct {
	Enable bool
	FileName string
	Source os.File
}

func (fl *FileLog) Create() error {
	var err error = nil

	if fl.Enable {
		fileName := "./" + fl.FileName + ".log"
		var flag int
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			flag = os.O_RDWR|os.O_CREATE
		} else {
			flag = os.O_APPEND|os.O_WRONLY
		}
		f, err := os.OpenFile(fileName, flag, 0755)

		if err != nil {
			log.Fatal(err)
		}

		fl.Source = *f
	}

	return err
}

func (fl *FileLog) Write(messages ...string) {
	if fl.Enable {
		str := strings.Replace("{date} {text} \r\n", "{date}", time.Now().String(), -1)

		var text string
		for _, message := range messages {
			text += message + " "
		}

		str = strings.Replace(str, "{text}", text, -1)
		fl.Source.WriteString(str)
	} else {
		log.Println(messages)
	}
}

func (fl *FileLog) WriteError(err error) {
	if fl.Enable {
		fl.Write(err.Error())
	} else {
		log.Println(err)
	}
}

func (fl *FileLog) Close() {
	fl.Write("Close log")
	fl.Source.Close()
}

func NewLogger(fileName string, conf *config.Configuration) *FileLog {
	fl := &FileLog{Enable: conf.Logger, FileName: fileName}
	fl.Create()
	return fl
}
