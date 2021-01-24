package errcode

import "errors"

var (
	Err_invalid_params = errors.New("invalid params") // 输入参数错误
	Err_conflict       = errors.New("conflict")       // 数据冲突
	Err_not_found      = errors.New("not found")
	Err_forbidden      = errors.New("forbidden")
)
