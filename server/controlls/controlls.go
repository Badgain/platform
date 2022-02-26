package controlls
import
(
	"github.com/gorilla/mux"
)
type RouteController interface {
	Link(*mux.Router, ...mux.MiddlewareFunc)
}
