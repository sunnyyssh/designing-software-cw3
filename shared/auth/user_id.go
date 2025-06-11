package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
)

type ctxKey int

const HeaderUserID = "X-User-ID"

func MiddlewareUserID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		strUserID := r.Header.Get(HeaderUserID)
		if strUserID == "" {
			httplib.Send(w, 400, map[string]any{
				"error": fmt.Sprintf("%s header is not specified", HeaderUserID),
			})
			return
		}

		userID, err := uuid.FromString(strUserID)
		if err != nil {
			httplib.Send(w, 400, map[string]any{
				"error": fmt.Sprintf("cannot parse user ID from %s header: %s", HeaderUserID, err),
			})
			return
		}

		r = r.WithContext(putUserID(r.Context(), userID))

		next(w, r)
	}
}

func putUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKey(0), id)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(ctxKey(0))
	if val == nil {
		return uuid.Nil, false
	}

	return val.(uuid.UUID), true
}

func MustUserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		panic(fmt.Errorf("userID is not found in context.Context, maybe you missed authentication middleware"))
	}
	return userID
}
