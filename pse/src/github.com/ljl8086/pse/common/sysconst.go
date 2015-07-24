package common

import (
	fc "github.com/ljl8086/fdfsclient"
	"github.com/ljl8086/pse/utils"	
	"github.com/weilaihui/goconfig/config"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/Sirupsen/logrus"
	"strings"
	"strconv"
	"os"
	"io"
)


var(
	FdfsClient *fc.FdfsClient
	Cf *config.Config
	Db *sql.DB//数据库连接
	Log *Logger
)

//配置文件
var(	
	CfTypeDeadLineMap map[string]int64 //上传文件类型与有效期MAP.key 类型  val 有效期
	CfMaxFileUploadSize = 0 //上传最大文件限制
	CfFileSuffixs = "" //上传文件支持的文件类型
)
var(
	CfWebPort string
)
var(
	CfDBType string
	CfDBUrl string
)
var(
	CfLogOut string
	CfLogLevel string
	CfLogPath string
)

func init(){
	CfTypeDeadLineMap = make(map[string]int64)
	cf,err := config.ReadDefault("conf/pse.conf")
	CfMaxFileUploadSize,_ = cf.Int("upload","file.maxsize")
	CfMaxFileUploadSize = CfMaxFileUploadSize * 1024 * 1024
	
	CfFileSuffixs,_ = cf.RawString("upload","file.suffixs")
	typeDeadline,_ := cf.RawString("upload","type.deadline")
	tdls := strings.Split(typeDeadline,",")
	
	for index:=len(tdls)-1;index>=0;index-- {
		 dls := strings.Split(tdls[index],":")
		 dl,_ := strconv.ParseInt(dls[1],10,64)
		 CfTypeDeadLineMap[dls[0]] = dl
	}
	
	
	CfDBType, err := cf.RawString("db", "type")
	utils.CheckError(err)

	CfDBUrl, err = cf.RawString("db", "url")
	utils.CheckError(err)

	Db, err = sql.Open(CfDBType, CfDBUrl)
	utils.CheckError(err)
	
	FdfsClient, err = fc.NewFdfsClient("conf/client.conf")
	utils.CheckError(err)
	
	CfWebPort,err = cf.RawString("web","port")
	utils.CheckError(err)
	
	
	CfLogOut,err := cf.RawString("log","out")
	utils.CheckError(err)
	
	var logHandle io.Writer
	if strings.EqualFold("file",CfLogOut) {
		CfLogPath,err = cf.RawString("log","path")
		utils.CheckError(err)
		
		logHandle,err = os.OpenFile(CfLogPath,os.O_RDWR|os.O_CREATE,0777)
		utils.CheckError(err)
	}else{
		logHandle = os.Stderr;
	}
	
	levelStr,err := cf.RawString("log","level")
	utils.CheckError(err)
	
	logLevel,err := ParseLevel(levelStr)
	utils.CheckError(err)
	
	Log = &Logger{Out:logHandle,Formatter: new(TextFormatter),Hooks:make(LevelHooks),Level:logLevel}
}