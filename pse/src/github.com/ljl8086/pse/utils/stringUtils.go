package utils

import (
	"strings"
	"path"
)

//文件名解析。如 /usr/abc.d.txt
type FileName struct{
	FullPath string //含路径 /usr/abc.d.txt 
	Full string		//不含路径的文件名 abc.d.txt
	Ext string		//只有后缀 txt
	Prefix string	//只有前缀 abc.d
	Path string //只有路径
}

//返回文件名的后缀名
func ParseFileName(s string) FileName{
	fn := FileName{}
	fn.Full = path.Base(s)
	ext := path.Ext(s)
	fn.Ext = strings.TrimPrefix(ext,".")
	fn.Prefix = strings.TrimSuffix(fn.Full,ext)
	fn.FullPath = s
	fn.Path = path.Dir(s)
	return fn
}

func Join(span string,s ...string) string{
	return strings.Join(s,span)
}