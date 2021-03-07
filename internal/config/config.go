package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

//AppConfig Application Configuration
type AppConfig struct {
	GeoIPASNDB  string `envconfig:"GEO_IP_ASN_DB"`  // ASN Database
	GeoIPCityDB string `envconfig:"GEO_IP_CITY_DB"` // City Database

}

//GetIPASNDB returs cache uri for redis
func (a *AppConfig) GetIPASNDB() string {
	return a.GeoIPASNDB
}

//GetDBURI returns mongodb URI
func (a *AppConfig) GetGeoIPCityDB() string {
	return a.GeoIPCityDB
}

// GetConfig get the current configuration from the environment
func GetConfig() (AppConfig, error) {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Println(err)
	}
	return cfg, nil
}
