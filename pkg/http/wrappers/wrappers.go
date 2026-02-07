package wrappers

import (
	"net/http"
	"encoding/json"
)

type Error interface {
	Error() string
	Status() int
}

type HTTPError struct {
	Err error
	Code int
}
func (e HTTPError) Error() string {
	return e.Err.Error()
}
func (e HTTPError) Status() int {
	return e.Code
}

type Handler func(r *http.Request) (interface{}, Error)

type Response struct {
	Status string `json:"status"`
	Message string `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func Success(r any) any {
	return Response{
		Status: "success",
		Data: r,
	}
}

func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h(r)
		// return error if there's any
		if err != nil {
			msg := Response{
				Status: "error",
				Message: err.Error(),
			}
			returnJson(msg, err.Status(), w)
			return
		}

		msg := Response{Status: "success",
			Data: data,
		}
		returnJson(msg, http.StatusOK, w)
	}
}

func returnJson(resp any, code int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	if encdata, err := json.Marshal(resp); err == nil {
		w.WriteHeader(code)
		w.Write(encdata)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
