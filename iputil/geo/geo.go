package geo

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	geoip2 "github.com/oschwald/geoip2-golang/v2"
)

type Reader interface {
	Country(net.IP) (Country, error)
	City(net.IP) (City, error)
	ASN(net.IP) (ASN, error)
	BuildDate() string
	IsEmpty() bool
}

type Country struct {
	Name string
	ISO  string
	IsEU bool
}

type City struct {
	Name       string
	Latitude   float64
	Longitude  float64
	PostalCode string
	Timezone   string
	MetroCode  uint
	RegionName string
	RegionCode string
}

type ASN struct {
	AutonomousSystemNumber       uint
	AutonomousSystemOrganization string
}

type geoip struct {
	country *geoip2.Reader
	city    *geoip2.Reader
	asn     *geoip2.Reader
	build   string
}

func Open(countryDB, cityDB string, asnDB string) (Reader, error) {
	var country, city, asn *geoip2.Reader
	if countryDB != "" {
		r, err := geoip2.Open(countryDB)
		if err != nil {
			return nil, err
		}
		country = r
	}
	if cityDB != "" {
		r, err := geoip2.Open(cityDB)
		if err != nil {
			return nil, err
		}
		city = r
	}
	if asnDB != "" {
		r, err := geoip2.Open(asnDB)
		if err != nil {
			return nil, err
		}
		asn = r
	}
	return &geoip{country: country, city: city, asn: asn, build: latestBuildDate(country, city, asn)}, nil
}

func latestBuildDate(readers ...*geoip2.Reader) string {
	var latest uint
	for _, reader := range readers {
		if reader == nil {
			continue
		}
		if buildEpoch := reader.Metadata().BuildEpoch; buildEpoch > latest {
			latest = buildEpoch
		}
	}
	if latest == 0 {
		return ""
	}
	return time.Unix(int64(latest), 0).UTC().Format("2006-01-02")
}

func (g *geoip) Country(ip net.IP) (Country, error) {
	country := Country{}
	if g.country == nil {
		return country, nil
	}
	addr, err := addrFromIP(ip)
	if err != nil {
		return country, err
	}
	record, err := g.country.Country(addr)
	if err != nil {
		return country, err
	}
	country.Name = record.Country.Names.English
	if country.Name == "" {
		country.Name = record.RegisteredCountry.Names.English
	}
	if record.Country.ISOCode != "" {
		country.ISO = record.Country.ISOCode
	}
	if record.RegisteredCountry.ISOCode != "" && country.ISO == "" {
		country.ISO = record.RegisteredCountry.ISOCode
	}
	country.IsEU = record.Country.IsInEuropeanUnion
	return country, nil
}

func (g *geoip) City(ip net.IP) (City, error) {
	city := City{}
	if g.city == nil {
		return city, nil
	}
	addr, err := addrFromIP(ip)
	if err != nil {
		return city, err
	}
	record, err := g.city.City(addr)
	if err != nil {
		return city, err
	}
	city.Name = record.City.Names.English
	if len(record.Subdivisions) > 0 {
		city.RegionName = record.Subdivisions[0].Names.English
		if record.Subdivisions[0].ISOCode != "" {
			city.RegionCode = record.Subdivisions[0].ISOCode
		}
	}
	if record.Location.Latitude != nil {
		city.Latitude = *record.Location.Latitude
	}
	if record.Location.Longitude != nil {
		city.Longitude = *record.Location.Longitude
	}
	// Metro code is US Only https://maxmind.github.io/GeoIP2-dotnet/doc/v2.7.1/html/P_MaxMind_GeoIP2_Model_Location_MetroCode.htm
	if record.Location.MetroCode > 0 && record.Country.ISOCode == "US" {
		city.MetroCode = record.Location.MetroCode
	}
	if record.Postal.Code != "" {
		city.PostalCode = record.Postal.Code
	}
	if record.Location.TimeZone != "" {
		city.Timezone = record.Location.TimeZone
	}

	return city, nil
}

func (g *geoip) ASN(ip net.IP) (ASN, error) {
	asn := ASN{}
	if g.asn == nil {
		return asn, nil
	}
	addr, err := addrFromIP(ip)
	if err != nil {
		return asn, err
	}
	record, err := g.asn.ASN(addr)
	if err != nil {
		return asn, err
	}
	if record.AutonomousSystemNumber > 0 {
		asn.AutonomousSystemNumber = record.AutonomousSystemNumber
	}
	if record.AutonomousSystemOrganization != "" {
		asn.AutonomousSystemOrganization = record.AutonomousSystemOrganization
	}
	return asn, nil
}

func addrFromIP(ip net.IP) (netip.Addr, error) {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.Addr{}, fmt.Errorf("invalid IP address %q", ip)
	}
	return addr.Unmap(), nil
}

func (g *geoip) BuildDate() string {
	return g.build
}

func (g *geoip) IsEmpty() bool {
	return g.country == nil && g.city == nil
}
