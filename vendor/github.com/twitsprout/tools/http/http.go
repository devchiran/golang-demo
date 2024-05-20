package http

import (
	"io"
	"net/http"
	"net/url"

	"github.com/twitsprout/tools/json"
)

// JSONRes represents the high level successful response where Data is the
// data (any type) being sent.
type JSONRes struct {
	Data interface{} `json:"data"`
}

// JSONErrRes represents the high level error response where Error is the error
// object containing the error message.
type JSONErrRes struct {
	Error JSONErr `json:"error"`
}

// JSONErr represents a json error message.
type JSONErr struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

// ReadJSON unmarshals json from the provided reader into 'v', returning any
// error encountered.
func ReadJSON(r io.Reader, v interface{}) error {
	return json.Decode(r, v)
}

// WriteJSONData wraps the provided data in the JSONRes struct and writes the
// json representation to the response writer with the proper status code.
// E.g.
//     	type myType struct {
//     		Msg string `json:"msg"`
//     	}
//	data := myType{"message"}
//	WriteJSONData(w, data, 200)
//
// Writes:
// 	{"data":{"msg":"message"}}
// to w with the status code '200'.
func WriteJSONData(w http.ResponseWriter, v url.Values, data interface{}, code int) error {
	res := JSONRes{data}
	return WriteJSON(w, v, res, code)
}

// WriteJSONError wraps the provided error message in the JSONErrRes struct and
// writes the json representation to the response writer with the proper status
// code.
// E.g.
//	msg := "message"
//	WriteJSONError(w, msg, 400)
//
// Writes:
// 	{"error":{"message":"message"}}
// to w with the status code '400'.
func WriteJSONError(w http.ResponseWriter, v url.Values, msg string, code int) error {
	res := JSONErrRes{Error: JSONErr{Message: msg}}
	return WriteJSON(w, v, res, code)
}

// WriteJSON marshals the provided res to json and writes it to the response
// writer with the proper status code.
func WriteJSON(w http.ResponseWriter, v url.Values, res interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.Encode(w, res, indent(v))
}

func indent(v url.Values) string {
	if isPretty(v) {
		return "  "
	}
	return ""
}

// isPretty returns true of the "pretty" query parameter is provided and is not
// "false".
func isPretty(v url.Values) bool {
	if v == nil {
		return false
	}
	vals, ok := v["pretty"]
	if !ok {
		return false
	}
	for _, val := range vals {
		if val == "false" {
			return false
		}
	}
	return true
}
