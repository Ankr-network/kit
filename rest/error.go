package rest

import (
	"com.ankr.kit/rest/proto"
	"encoding/json"
	"net/http"
	"strings"
)

var (
	jsonMarshalErr = []byte(`{"error":"InternalError", "message":"failed to marshal error message"`)
)

func Error(w http.ResponseWriter, err error, code int) {
	errStr := err.Error()
	rspErr := new(proto.Error)
	idx := strings.IndexRune(errStr, ':')
	if idx < 0 {
		rspErr.Error = strings.TrimSpace(errStr)
		rspErr.Message = rspErr.Error
	} else {
		rspErr.Error = strings.TrimSpace(errStr[0:idx])
		rspErr.Message = strings.TrimSpace(errStr[idx+1:])
	}

	errBytes, err := json.Marshal(rspErr)
	if err != nil {
		errBytes = jsonMarshalErr
	}

	addJsonHeader(w)
	w.WriteHeader(code)
	w.Write(errBytes)
}
