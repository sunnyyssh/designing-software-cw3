package httplib

import (
	"encoding/json"
	"io"
	"net/http"
)

func UnmarshalBody[T any](req *http.Request) (mock T, _ error) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return mock, err
	}

	var res T
	if err := json.Unmarshal(data, &res); err != nil {
		return mock, err
	}

	return res, nil
}

func Send(w http.ResponseWriter, code int, body any) error {
	w.WriteHeader(code)

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
