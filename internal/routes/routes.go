package routes

import (
	"net/http"
	"os"

	geolookuphandlers "geolookup/internal/handlers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//GetRoutes get the routes for the application
func GetRoutes(hctx *geolookuphandlers.HTTPHandlerContext) *mux.Router {

	r := mux.NewRouter()

	// Index Page
	r.Handle(
		"/",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(hctx.HomeHandler))).Methods("GET")

	// Health check page
	r.Handle(
		"/healthcheck957873",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(hctx.HealthCheckHandler))).Methods("GET")

	// Get City, State, County, ASN by IP Address
	r.Handle(
		"/api/v1/geoiplookup/{ip}",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(hctx.GetCityStateCountryASNByIPAddressAPI))).Methods("GET")

	return r
}
