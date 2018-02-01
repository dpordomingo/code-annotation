package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/src-d/code-annotation/server/serializer"
	"github.com/src-d/code-annotation/server/service"
)

// RequestProcessFunc is a function that takes an http.Request, and returns a serializer.Response and an error
type RequestProcessFunc func(*http.Request) (*serializer.Response, error)

// Get returns an http.HandlerFunc that will serve the user request taking the serializer.Response and errors
// from the passed RequestProcessFunc
func Get(rp RequestProcessFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := rp(r)
		if response == nil {
			response = serializer.NewEmptyResponse()
		}

		write(w, r, response, err)
	}
}

// write is the responsible of writing the response with the data from the passed *serializer.Response and error
// If the passed error has StatusCode, the http.Response will be returned with the StatusCode of the passed error
// If the passed error has not StatusCode, the http.Response will be returned as a http.StatusInternalServerError
func write(w http.ResponseWriter, r *http.Request, response *serializer.Response, err error) {
	var statusCode int

	// TODO: There should be no ppl calling write from the outside
	if response == nil {
		response = serializer.NewEmptyResponse()
	}

	if err == nil {
		statusCode = http.StatusOK
	} else if httpError, ok := err.(serializer.HTTPError); ok {
		statusCode = httpError.StatusCode()
		response.Status = statusCode
		response.Errors = []serializer.HTTPError{httpError}
	} else {
		statusCode = http.StatusInternalServerError
		response.Status = statusCode
		response.Errors = []serializer.HTTPError{serializer.NewHTTPError(statusCode, err.Error())}
	}

	if statusCode >= http.StatusBadRequest {
		service.NewLogger().Error(err.Error())
	}

	content, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("response could not be marshalled; %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		service.NewLogger().Error(err.Error())
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(content)
}

// URLParamInt returns the url parameter from a http.Request object. If the
// param cannot be converted to int, it returns a serializer.NewHTTPError
func urlParamInt(r *http.Request, key string) (val int, err error) {
	str := chi.URLParam(r, key)
	val, err = strconv.Atoi(str)

	if err != nil {
		val = 0
		err = serializer.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("Wrong format for URL parameter %q; received %q", key, str),
		)
	}

	return
}
