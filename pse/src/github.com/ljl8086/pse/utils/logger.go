package utils

import (
	. "github.com/Sirupsen/logrus"
	"github.com/weilaihui/goconfig/config"
	"os"
	"strings"
	"io"
)

var( 
	Log *Logger
	logPath string
	logLevel Level
	//logHandle *File
)

func init(){
	var logHandle io.Writer;
	cf,err := config.ReadDefault("conf/pse.conf")
	CheckError(err)
	
	out,err := cf.RawString("log","out")
	CheckError(err)
	if strings.EqualFold("file",out) {
		logPath,err = cf.RawString("log","path")
		CheckError(err)
		
		logHandle,err = os.OpenFile(logPath,os.O_RDWR|os.O_CREATE,0777)
		CheckError(err)
	}else{
		logHandle = os.Stderr;
	}
	
	levelStr,err := cf.RawString("log","level")
	CheckError(err)
	
	logLevel,err = ParseLevel(levelStr)
	CheckError(err)

	//defer logHandle.Close()
	
	Log = &Logger{Out:logHandle,Formatter: new(TextFormatter),Hooks:make(LevelHooks),Level:logLevel}
}
