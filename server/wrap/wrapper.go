package wrap

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/unrolled/render"
	"io/ioutil"
	"mahjong/server/api"
	"net/http"
)

var respRender = render.New()

type AnyFunc[T any, R any] func(http.ResponseWriter, *http.Request, T) (R, error)

type AnyHandler[T any, R any] struct {
	fn AnyFunc[T, R]
}

type AnyResp[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    *T     `json:"data"`
}

type AnyError struct {
	Code    string
	Message string
}

func (err *AnyError) Error() string {
	return fmt.Sprintf("%s:%s", err.Code, err.Message)
}

func NewAnyError(code string, message string) *AnyError {
	return &AnyError{
		Code:    code,
		Message: message,
	}
}

func (receiver AnyHandler[T, R]) Func() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		//请求 解码
		var t T
		bodyByte, err := ioutil.ReadAll(request.Body)
		err = json.Unmarshal(bodyByte, &t)
		if err != nil {
			_ = respRender.Text(writer, 500, "api json unmarshal error :"+err.Error())
		}

		//header
		apiHeader := &api.UserHeader{
			UserId: request.Header.Get("user_id"),
			Token:  request.Header.Get("token"),
		}
		request.WithContext(context.WithValue(request.Context(), "header", apiHeader))

		//执行
		r, err := receiver.fn(writer, request, t)

		//响应 输出
		if err != nil {
			//自定义
			resp := &AnyResp[R]{Data: nil}
			if ef, ok := err.(*AnyError); ok {
				resp.Code = ef.Code
				resp.Message = ef.Message
			} else {
				resp.Code = "500"
				resp.Message = err.Error()
			}
			_ = respRender.JSON(writer, 200, resp)
		} else {
			_ = respRender.JSON(writer, 200, &AnyResp[R]{Code: "200", Message: "success", Data: &r})
		}
	}
}

func NewWrapper[T any, R any](fn AnyFunc[T, R]) AnyHandler[T, R] {
	return AnyHandler[T, R]{
		fn: fn,
	}
}

func GetHeader(request *http.Request) *api.UserHeader {
	return request.Context().Value("header").(*api.UserHeader)
}
