package vo

import (

)

//文件上传响应接口
type UploadResVO struct{
	ResVO
	Filename string `json:"filename"`
}

type ResVO struct{
	Status int `json:"status"`
	Message string `json:"message"`
}