package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"github.com/ljl8086/pse/controls"
	"reflect"
	"github.com/ljl8086/pse/utils"
	"github.com/weilaihui/goconfig/config"
//	"path"
//	"image"
//	_ "image/png"
//	"image/jpeg"
//	"fmt"
//	"code.google.com/p/graphics-go/graphics"
)

//url与执行方法的映射表。
var routeFuncMap map[string]reflect.Value
var port string;

func init(){
	routeFuncMap = make(map[string]reflect.Value)
	routeFuncMap["/down"] = reflect.ValueOf(controls.Down)
	routeFuncMap["/upload"] = reflect.ValueOf(controls.Upload)
	
	cf,err := config.ReadDefault("conf/pse.conf")
	utils.CheckError(err)
	
	port,err = cf.RawString("web","port")
	utils.CheckError(err)
}

//URL路由映射选择器。
func route(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	url := req.URL
	if strings.HasPrefix(url.Path, "/static/") {
		html := strings.TrimLeft(url.Path, "/")
		srcFile, err := os.Open(html)
		if err == nil {
			defer srcFile.Close()
			io.Copy(res, srcFile)
		}
	}else{
		params := make([]reflect.Value, 2)
		params[0] = reflect.ValueOf(res)
  		params[1] = reflect.ValueOf(req)
		val,ok := routeFuncMap[url.Path]
		if ok{
			val.Call(params)
		}
	}
}

func main() {
	http.HandleFunc("/", route)
	err := http.ListenAndServe(port, nil)
	utils.CheckError(err)

//	file,err := os.Open("c:/IMG_2329.JPG")
//	utils.CheckError(err)
//	
//	src,name,err := image.Decode(file)
//	utils.CheckError(err)
//	fmt.Println(name)
//	
//	dstImg := image.NewRGBA(image.Rect(0,0,200,300))
//	graphics.Thumbnail(dstImg,src)
//	
//	dstFile,err := os.OpenFile("c:/test00.png",os.O_CREATE,0777)
//	utils.CheckError(err)
//	
//	quality := jpeg.Options{50}
//	err = jpeg.Encode(dstFile,dstImg,&quality)
//	utils.CheckError(err)
//	
//	fmt.Println("=============================")
	
}
