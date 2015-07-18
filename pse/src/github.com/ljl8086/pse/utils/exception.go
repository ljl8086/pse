package utils

import ()

//处理异常信息
func CheckError(err error) {
	if err != nil {
		panic(err);
	}
}
