package main

import (
	"flag"
	"github.com/pobizhe/flora-kit/flora"
)

func main() {
	var configFile, geoipdb string
	flag.StringVar(&configFile, "s", "flora.default.conf", "specify surge config file")
	flag.StringVar(&geoipdb, "d", "geoip.mmdb", "specify geoip db file")
	flora.Run(configFile)

}
