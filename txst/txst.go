package txst

import (
	"net/http"
	"testing"

	"github.com/thrisp/flotilla/route"
	"github.com/thrisp/flotilla/state"
)

type Tanage func(*testing.T) state.Manage

type TxstApp interface {
	Manage(*route.Route)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
