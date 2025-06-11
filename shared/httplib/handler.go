package httplib

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
)

func HandlerJSON(handle HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With("path", r.URL.Path, "method", r.Method)

		logger.InfoContext(r.Context(), "request accepted")

		res, err := handle(r)

		if err != nil {
			var httpErr errs.HTTPError
			if errors.As(err, &httpErr) {
				w.WriteHeader(httpErr.Code)

				data, _ := json.Marshal(map[string]any{
					"error": httpErr.Message,
				})
				w.Write(data)

				logger.InfoContext(r.Context(), "request served", "code", httpErr.Code, "error", err)
			} else {
				w.WriteHeader(500)
				w.Write([]byte(`{"error": "Internal server error"}`))

				logger.ErrorContext(r.Context(), "request failed", "code", 500, "error", err)
			}
			return
		}

		w.WriteHeader(200)

		if res != nil {
			data, err := json.Marshal(res)
			if err != nil {
				panic(err)
			}

			w.Write(data)
		}

		logger.InfoContext(r.Context(), "request served", "code", 200)
	}
}
