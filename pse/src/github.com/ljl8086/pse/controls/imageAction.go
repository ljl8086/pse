package controls

import (
	"net/http"
	"github.com/ljl8086/pse/vo"
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
	 "github.com/ljl8086/pse/db"
	cm "github.com/ljl8086/pse/common"
)
var(
	picNotExistShowJpgBuf []byte
)

func init(){
	var err error
	picNotExistShowJpgBuf,err = ioutil.ReadFile("static/images/picNotExistShowJpg.jpg")
	CheckError(err)
}

//文件下载接口。
// 如果同时指定了w和h参数，将返回wxh的截图
// 如果同时只指定w参数，将根据w进行等比缩放
func Down(res http.ResponseWriter, req *http.Request) {
	cm.Log.Debug("--------------Down---------------------")
	t1 := time.Now()
	defer logTime("down 所花时间：",t1)
	
	req.ParseForm()
	imageVO := vo.Map2Vo(req.Form);
	fn := ParseFileName(imageVO.FileName)
	
	if(fn.IsImg() && imageVO.Width>0){
		tempName := Join("",fn.Path,"/",fn.Prefix,"_",strconv.Itoa(imageVO.Width))
		if(imageVO.Height>0){
			tempName = Join("",tempName,"x",strconv.Itoa(imageVO.Height))
		}
		tempName = Join("",tempName,".",fn.Ext)
		
		buf,err := cm.FdfsClient.DownloadToBuffer(tempName,0,0)
		if(err==nil){
			writeFileBUfRes(res,tempName,buf.Content.([]byte))
			return
		}else{
			buf,err := cm.FdfsClient.DownloadToBuffer(imageVO.FileName,0,0)
			if(err!=nil){
				cm.Log.Error("download file has error:",err.Error())
				writeFileBUfRes(res,tempName,picNotExistShowJpgBuf)
				return
			}else{
				nbuf,err := makeSizeImg(fn.Ext,fn.FullPath,buf.Content.([]byte),imageVO.Width,imageVO.Height)
				if(err!=nil){
					cm.Log.Error("缩放图片生成失败:",err.Error())
					io.WriteString(res,"Sclae image make fail")
				}else{
					writeFileBUfRes(res,tempName,nbuf.Bytes())
				}
				return
			}
		}
	}else{
		buf,err := cm.FdfsClient.DownloadToBuffer(imageVO.FileName,0,0)
		if err!=nil{
			cm.Log.Error("download file has error:",err.Error())
			if(fn.IsImg()){
				writeFileBUfRes(res,imageVO.FileName,picNotExistShowJpgBuf)
			}else{
				io.WriteString(res,"File does not exist")
			}
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
	cm.Log.Debug(msg,time.Now().Sub(t).Nanoseconds()/1000000)
}

func Upload(res http.ResponseWriter, req *http.Request) {
	cm.Log.Debug("--------------upload---------------------")
	err := req.ParseMultipartForm(32<<40)
	if(err!=nil){
		res.Write(makeErr("upload file has error:",err))
		return
	}
	
	fileType := req.FormValue("type")
	fileTypeI,_ := strconv.Atoi(fileType)
	valTime,ok := cm.CfTypeDeadLineMap[fileType]
	if valTime==0 {
	 	valTime = 0xfffffffffffffff
	}else{
		valTime = time.Now().Unix()+valTime
	}
	
	if(!ok){
		res.Write(makeErr("upload file has error:",errors.New("Parameters in the interface can not be identified.")))
		return
	}
	
	file,handle,err := req.FormFile("file")
	if err!=nil{
		res.Write(makeErr("get upload file err:",err))
		return
	}
	defer file.Close()
	
	fn := ParseFileName(handle.Filename)
	if len(fn.Ext)==0 || !strings.Contains(cm.CfFileSuffixs, fn.Ext) {
		res.Write(makeErr("upload file fail",errors.New(Join("","Only support these types of files:",cm.CfFileSuffixs))))
		return
	}
	
	buf,err := ioutil.ReadAll(file)
	if len(buf) > cm.CfMaxFileUploadSize{
		res.Write(makeErr("upload file fail",errors.New("Only support under 5 MB file")))
		return
	}
	
	ures,err := cm.FdfsClient.UploadByBuffer(buf,fn.Ext)
	if err!=nil{
		res.Write(makeErr("upload file to fdfs err:",err))
		return
	}
	fdb := vo.FilesDB{FileName:ures.RemoteFileId, FileType:fileTypeI, Deadline:valTime, IsSlave:false}
	db.SaveFiles(&fdb,"")
	
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
			cm.Log.Error("image decode has err:",err)
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
			cm.Log.Error("thumb has err:",err)
			return bbuf,err
		}
		
		ures,err :=  cm.FdfsClient.UploadSlaveByBuffer(bbuf.Bytes(),remoteFileId,prefixName,ext)
		if(err!=nil){
			cm.Log.Error("thumb upload err:",err)
			return bbuf,err
		}
		
		fdb := vo.FilesDB{FileName:ures.RemoteFileId, IsSlave:true, SlaveSuffix:prefixName}
		db.SaveFiles(&fdb, remoteFileId)
		
		cm.Log.Debug("缩略图文件已经成功产生,前缀:",prefixName,"原文件名:",remoteFileId)
		return bbuf,err
	}else{
		err = errors.New("只支持PNG、JPG、JPEG")
		return bbuf,err
	}
}

func makeErr(msg string,err error) []byte{
	if len(msg)>0 {
		cm.Log.Error(msg,err.Error())
	}
	vo := vo.ResVO{Status:-998,Message:err.Error()}
	bytes,_ := json.Marshal(vo)
	return bytes
}
