package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"github.com/ljl8086/pse/controls"
	"reflect"
	"github.com/ljl8086/pse/utils"
	cm "github.com/ljl8086/pse/common"
	"fmt"
//	"path"
//	"image"
//	_ "image/png"
//	"image/jpeg"
//	"code.google.com/p/graphics-go/graphics"
)

//url与执行方法的映射表。
var routeFuncMap map[string]reflect.Value

func init(){
	routeFuncMap = make(map[string]reflect.Value)
	routeFuncMap["/down"] = reflect.ValueOf(controls.Down)
	routeFuncMap["/upload"] = reflect.ValueOf(controls.Upload)
}

//URL路由映射选择器。
func route(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	url := req.URL
	if strings.EqualFold(url.Path,"/"){
		srcFile, err := os.Open("static/index.html")
		if err == nil {
			defer srcFile.Close()
			io.Copy(res, srcFile)
		}
	}else if strings.HasPrefix(url.Path, "/static/") {
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
	fmt.Println("Welcome to PSE world!")
	http.HandleFunc("/", route)
	err := http.ListenAndServe(cm.CfWebPort, nil)
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
