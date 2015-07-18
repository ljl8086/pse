package controls

import (
	f "fmt"
	"net/http"
	fc "github.com/ljl8086/fdfsclient"
	"github.com/ljl8086/pse/vo"
	"os/signal"
	"io"
	"io/ioutil"
	. "github.com/ljl8086/pse/utils"
	"code.google.com/p/graphics-go/graphics"
	"image"
	"image/jpeg"
	_ "image/png"
	"bytes"
	"encoding/json"
	"strings"
	"strconv" 
	"errors"
	"time"
)

var fdfsClient *fc.FdfsClient

func init(){
	fclient, err := fc.NewFdfsClient("conf/client.conf")
	if(err!=nil){
		f.Println("fdfsClient init error!")
		signal.Stop(nil)
	}
	fdfsClient = fclient
}

//文件下载接口。
// 如果同时指定了w和h参数，将返回wxh的截图
// 如果同时只指定w参数，将根据w进行等比缩放
func Down(res http.ResponseWriter, req *http.Request) {
	Log.Debug("--------------Down---------------------")
	t1 := time.Now()
	defer logTime("down 所花时间：",t1)
	
	req.ParseForm()
	imageVO := vo.Map2Vo(req.Form);
	fn := ParseFileName(imageVO.FileName)
	
	if(imageVO.Width>0){
		tempName := Join("",fn.Path,"/",fn.Prefix,"_",strconv.Itoa(imageVO.Width))
		if(imageVO.Height>0){
			tempName = Join("","x",tempName,strconv.Itoa(imageVO.Height))
		}
		tempName = Join("",tempName,".",fn.Ext)
		
		buf,err := fdfsClient.DownloadToBuffer(tempName,0,0)
		if(err==nil){
			writeFileBUfRes(res,tempName,buf.Content.([]byte))
			return
		}else{
			buf,err := fdfsClient.DownloadToBuffer(imageVO.FileName,0,0)
			if(err!=nil){
				Log.Error("download file has error:",err.Error())
				io.WriteString(res,"File does not exist")
				return
			}else{
				nbuf,err := makeSizeImg(fn.Ext,fn.FullPath,buf.Content.([]byte),imageVO.Width,imageVO.Height)
				if(err!=nil){
					Log.Error("缩放图片生成失败:",err.Error())
					io.WriteString(res,"Sclae image make fail")
				}else{
					writeFileBUfRes(res,tempName,nbuf.Bytes())
				}
				return
			}
		}
	}else{
		buf,err := fdfsClient.DownloadToBuffer(imageVO.FileName,0,0)
		if err!=nil{
			Log.Error("download file has error:",err.Error())
			io.WriteString(res,"File does not exist")
		}else{
			writeFileBUfRes(res,imageVO.FileName,buf.Content.([]byte))
		}
	}
}

//向响应里写入文件
func writeFileBUfRes(res http.ResponseWriter,fileName string,buf []byte){
	header := res.Header()
	header.Set("Content-Disposition",Join("","attachment; filename=",fileName))
	header.Set("ContentType","image/jpeg")
	res.Write(buf)
}

func logTime(msg string,t time.Time){
	Log.Debug(msg,time.Now().Sub(t).Nanoseconds()/1000000)
}

func Upload(res http.ResponseWriter, req *http.Request) {
	Log.Debug("--------------upload---------------------")
	err := req.ParseMultipartForm(32<<40)
	
	if(err!=nil){
		res.Write(makeErr("download file has error:",err))
		return
	}
	
	file,handle,err := req.FormFile("file")
	if err!=nil{
		res.Write(makeErr("get upload file err::",err))
		return
	}
	defer file.Close()
	
	fn := ParseFileName(handle.Filename)
	buf,err := ioutil.ReadAll(file)
	ures,err := fdfsClient.UploadByBuffer(buf,fn.Ext)
	if err!=nil{
		res.Write(makeErr("upload file to fdfs err:",err))
		return
	}
	
	makeSizeImg(fn.Ext,ures.RemoteFileId,buf,128,128)
	resVO := vo.UploadResVO{Filename:ures.RemoteFileId}
	bytes,_ := json.Marshal(resVO)
	res.Write(bytes)
}

//生成缩略图并上传
func makeSizeImg(ext string,remoteFileId string,buf []byte,width,height int) (bbuf *bytes.Buffer,err error){
//	var bbuf *bytes.Buffer
	if strings.EqualFold("jpg",ext) || strings.EqualFold("jpeg",ext) || strings.EqualFold("png",ext){
		var (
			dstImg *image.RGBA
			prefixName string
			)
		
		bufRead := bytes.NewReader(buf)
		srcImg,_,err := image.Decode(bufRead)
		if(err!=nil){
			Log.Error("image decode has err:",err)
			return bbuf,err
		}
		if(width>0 && height>0){
			dstImg = image.NewRGBA(image.Rect(0,0,width,height))
			graphics.Thumbnail(dstImg,srcImg)
			prefixName = Join("","_",strconv.Itoa(width),"x", strconv.Itoa(height))
		}else{
			bounds := srcImg.Bounds()
			x := bounds.Dx()
			y := bounds.Dy()
			height = int(float32(width) * (float32(y)/float32(x)))
 			dstImg = image.NewRGBA(image.Rect(0,0,width,height))
			graphics.Scale(dstImg,srcImg)
			prefixName = Join("","_",strconv.Itoa(width))
		}
		
		bbuf = new(bytes.Buffer)
		quality := jpeg.Options{50}
		err = jpeg.Encode(bbuf,dstImg,&quality)
		if(err!=nil){
			Log.Error("thumb has err:",err)
			return bbuf,err
		}
		
		_,err = fdfsClient.UploadSlaveByBuffer(bbuf.Bytes(),remoteFileId,prefixName,ext)
		if(err!=nil){
			Log.Error("thumb upload err:",err)
			return bbuf,err
		}
		Log.Debug("缩略图文件已经成功产生,前缀:",prefixName,"原文件名:",remoteFileId)
		return bbuf,err
	}else{
		err = errors.New("只支持PNG、JPG、JPEG")
		return bbuf,err
	}
}

func makeErr(msg string,err error) []byte{
	if len(msg)>0 {
		Log.Error(msg,err.Error())
	}
	vo := vo.ResVO{Status:-998,Message:err.Error()}
	bytes,_ := json.Marshal(vo)
	return bytes
}
