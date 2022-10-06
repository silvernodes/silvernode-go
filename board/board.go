package board

import (
	"fmt"
	"net/http"
)

func DashBoard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello Silvernode-Go!</h1>")
}
