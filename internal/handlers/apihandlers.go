package handlers

import (
	"bytes"
	"encoding/json"
	"net"

	"geolookup/internal/config"

	"github.com/rs/zerolog/log"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus Metrics
var (
	searchesTotalProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "searches_total",
		Help: "The total number of geo search events",
	})

	searchesFailedProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "searches_failed_total",
		Help: "The total number of geo events failed",
	})

	searchLatency = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "search_latency_seconds",
		Help:       "Database Search in Seconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)

//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// GeoIP Object
type GeoIP struct {
	IsFound        bool   `json:"is_found"`
	IsPrivate      bool   `json:"is_private"`
	IPAddress      net.IP `json:"ip_addr"`
	City           string `json:"city"`
	CountryName    string `json:"country_name"`
	CountryISOCode string `json:"country_iso_code"`
	TimeZone       string `json:"time_zone"`
	IsProxy        bool   `json:"is_proxy"`
	IsEU           bool   `json:"is_eu"`
	ASN            string `json:"asn"`
	Organization   string `json:"organization"`
	Network        string `json:"network"`
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
	// TODO: Add IPV6 Ranges here
}

// GetCityStateCountryASNByIPAddressAPI view list of affiliates
func (ctx *HTTPHandlerContext) GetCityStateCountryASNByIPAddressAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	timer := prometheus.NewTimer(searchLatency)
	defer timer.ObserveDuration()

	type Message struct {
		Message     string `json:"message"`
		IsError     bool   `josn:"is_error"`
		GeoLocation GeoIP  `json:"geo_location"`
	}

	var jsonMessage Message

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["ip"]; ok {

	} else {
		log.Error().Msgf("Error getting url variable, ip: %s", val)
		jsonMessage.IsError = true
		jsonMessage.Message = "Could not determine variable from url parameter /api/v1/geoiplookup/{ip}"
	}

	ipAddr := vars["ip"]

	var geoIP GeoIP

	err := geoIP.GeoSearch(ipAddr, ctx.hConfig)

	if err != nil {
		log.Error().Err(err).Str("service", "apihandler").Msgf("Error searching for City using IP %s", ipAddr)
		jsonMessage.IsError = true
		jsonMessage.Message = "Could not determine variable from url parameter /api/v1/geoiplookup/{ip}"
	}

	jsonMessage.GeoLocation = geoIP

	searchesTotalProcessed.Inc()

	sb, err := json.Marshal(jsonMessage)
	w.Write(sb)

	return

}

func (ctx *HTTPHandlerContext) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "healthy")
}

// HomeHandler Displays the home page with list of posts
func (ctx *HTTPHandlerContext) HomeHandler(w http.ResponseWriter, r *http.Request) {
	out := "GeoIP Lookup API"
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

// IsPrivateSubnet - check to see if this ip is in a private subnet
func IsPrivateSubnet(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

// Search get geoip information from ipaddress
func (g *GeoIP) GeoSearch(ipaddress string, config *config.AppConfig) error {

	if ipaddress == "::1" {
		g.Organization = "None"
		g.ASN = "None"
		return nil
	}

	if ipaddress == "127.0.0.1" {
		g.Organization = "None"
		g.ASN = "None"
		return nil
	}

	ip := net.ParseIP(ipaddress)
	if ip == nil {
		g.IsFound = false
		return fmt.Errorf("error converting string [ %s ] to IP Address", ipaddress)
	}

	if IsPrivateSubnet(ip) == true {
		log.Info().Str("service", "apihandler").Msgf("Address appears to be from a private subnet for GEO Search: %s", ipaddress)
		g.CountryISOCode = "US"
		g.CountryName = "Merica"
		g.IsEU = false
		g.IsPrivate = true
		g.IPAddress = net.ParseIP("127.0.0.1")
		g.City = "Boise"
		return nil
	}

	// Begin  ASN Search
	dbasn, err := geoip2.Open("db/GeoLite2-ASN.mmdb")
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "apihandler").Msgf("Could not stat ASN Database: %s", config.GeoIPASNDB)
		return fmt.Errorf("Could not Open ASN Database file: %s with error: %s", config.GeoIPASNDB, err)
	}
	defer dbasn.Close()

	record, err := dbasn.ASN(ip)
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "apihandler").Msgf("Could not find record in ASN Database: %s", config.GeoIPASNDB)
		return fmt.Errorf("error getting database record: %s: with error %s", ip, err)
	}

	g.Organization = record.AutonomousSystemOrganization
	g.ASN = fmt.Sprint(record.AutonomousSystemNumber)

	// Begin City State Search
	dbcity, err := geoip2.Open(config.GeoIPCityDB)
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "apihandler").Msgf("Could not open city database: %s", config.GeoIPCityDB)
		return fmt.Errorf("error opening city geodatabase")
	}

	defer dbcity.Close()

	crecord, err := dbcity.City(ip)
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "apihandler").Msgf("Could not find ip: %s in city database: %s", ip, config.GeoIPASNDB)
		return fmt.Errorf("error finding ip: %s in city database: %s", ip, err)
	}

	// Each language is represented in a map
	g.City = crecord.City.Names["en"]

	// Each language is represented in a map
	g.CountryName = crecord.Country.Names["en"]

	g.CountryISOCode = crecord.Country.IsoCode

	g.IPAddress = ip

	g.TimeZone = crecord.Location.TimeZone

	g.IsProxy = crecord.Traits.IsAnonymousProxy

	g.IsEU = crecord.Country.IsInEuropeanUnion

	g.IsFound = true

	return nil
}
