package common

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	errors2 "github.com/pkg/errors"

	"github.com/win5do/golang-microservice-demo/pkg/log"

	errcode2 "github.com/win5do/golang-microservice-demo/pkg/api/errcode"
	"github.com/win5do/golang-microservice-demo/pkg/config"
)

type Message struct {
	Message string `json:"message,omitempty"`
}

func Response(c *gin.Context, err error, data interface{}) {
	if err == nil {
		httpCode := http.StatusOK
		if data != nil {
			c.JSON(httpCode, data)
		} else {
			c.JSON(httpCode, &Message{
				Message: "success",
			})
		}
		return
	}

	// 错误处理
	log.Debugf("err: %+v", err)

	var httpCode int

	var jsonErr = &json.SyntaxError{}

	switch {
	case errors2.Is(err, gorm.ErrRecordNotFound),
		errors2.Is(err, errcode2.Err_not_found):
		httpCode = http.StatusNotFound
		err = errcode2.Err_not_found // 避免将数据库报错暴露出去

	case errors2.Is(err, errcode2.Err_invalid_params),
		errors2.As(err, &jsonErr):
		// *json.SyntaxError implement error, not json.SyntaxError
		httpCode = http.StatusBadRequest
	case errors2.Is(err, errcode2.Err_forbidden):
		httpCode = http.StatusForbidden
	default:
		httpCode = http.StatusInternalServerError
	}

	if !config.GetConfig().Debug {
		// 非debug模式，只返回简略err
		err = errors2.Cause(err)
	}

	c.JSON(httpCode, &Message{
		Message: err.Error(),
	})
}

type ginUtil struct{}

var GinUtil = new(ginUtil)

const (
	OFFSET = "offset"
	LIMIT  = "limit"
)

func (s *ginUtil) GetOffsetLimit(c *gin.Context) (int, int) {
	// 参数不正确当0处理
	offset, err := strconv.Atoi(c.Query(OFFSET))
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(c.Query(LIMIT))
	if err != nil {
		limit = 0
	}

	return offset, limit
}
