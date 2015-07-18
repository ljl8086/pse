package vo

import (
	"strconv"
)

type ImageVO struct{
	FileName string //文件名
	Width int	//图片宽（象素）
	Height int	//图片高（象素）
	Thumb int	//缩略图  1 是  0否
	Quality int	//质量:10最高
}

func Map2Vo(m map[string][]string) ImageVO {
	var vo  ImageVO;
	if len(m["n"])>0{
		vo.FileName = m["n"][0]
	}
	if len(m["w"])>0{
		vo.Width,_ = strconv.Atoi(m["w"][0])
	}
	if len(m["h"])>0{
		vo.Height,_ = strconv.Atoi(m["h"][0])
	}
	if len(m["t"])>0{
		vo.Thumb,_ = strconv.Atoi(m["t"][0])
	}
	if len(m["q"])>0{
		vo.Quality,_ = strconv.Atoi(m["q"][0])
	}
    return vo
}

