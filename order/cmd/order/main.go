package main

import (
	"net/http"

	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
)

func main() {
	r := httplib.NewServer()

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(r)
	}
}
