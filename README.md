# Maxmind GeoIP Lookup service

Provides a simple yet effective geoip lookup service

## How to use

```bash

curl  http://localhost:4990/api/v1/geoiplookup/134.121.12.54
```


```json
{"message":"","IsError":false,"geo_location":{"is_found":true,"is_private":false,"ip_addr":"134.121.12.54","city":"Pullman","country_name":"United States","country_iso_code":"US","time_zone":"America/Los_Angeles","is_proxy":false,"is_eu":false,"asn":"11827","organization":"WSU-AS","network":""}}
```

## TODO

- Perhaps look into using redis service for in memory caching
- The files are static, so a redeploy is needed to refresh the database

