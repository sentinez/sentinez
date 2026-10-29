// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package geoip fills the client location of the logs that ask for it
// from the MaxMind GeoLite2 databases (GeoLite2-Country.mmdb,
// GeoLite2-City.mmdb, GeoLite2-ASN.mmdb) of a file system, e.g. the
// ones the enterprise build compiles in, ee.GeoIP. Without them the
// lookups are no-ops.
package geoip

import (
	"errors"
	"fmt"
	"io/fs"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

// The database files looked up in the data directory.
const (
	CountryFile = "GeoLite2-Country.mmdb"
	CityFile    = "GeoLite2-City.mmdb"
	AsnFile     = "GeoLite2-ASN.mmdb"
)

// CountrySetter is implemented by the logs that ask for the ISO 3166-1
// country code of the client.
type CountrySetter interface {
	SetCountry(isoCode string)
}

// CitySetter is implemented by the logs that ask for the English name
// of the client city.
type CitySetter interface {
	SetCity(name string)
}

// AsnSetter is implemented by the logs that ask for the autonomous
// system number of the client.
type AsnSetter interface {
	SetAsn(number int64)
}

// AsnOrgSetter is implemented by the logs that ask for the organization
// owning the autonomous system of the client.
type AsnOrgSetter interface {
	SetAsnOrg(org string)
}

// DB looks the client location up in the loaded databases. The zero
// value and nil have none loaded: Fill is then a no-op.
type DB struct {
	country *maxminddb.Reader
	city    *maxminddb.Reader
	asn     *maxminddb.Reader
}

// asnRecord holds the fields read from a GeoLite2-ASN record.
type asnRecord struct {
	Number uint   `maxminddb:"autonomous_system_number"`
	Org    string `maxminddb:"autonomous_system_organization"`
}

// New loads the CountryFile, CityFile and AsnFile databases found at the
// root of fsys; the missing ones are skipped.
func New(fsys fs.FS) (*DB, error) {
	var (
		db  DB
		err error
	)

	if db.country, err = open(fsys, CountryFile); err != nil {
		return nil, err
	}
	if db.city, err = open(fsys, CityFile); err != nil {
		return nil, err
	}
	if db.asn, err = open(fsys, AsnFile); err != nil {
		return nil, err
	}

	return &db, nil
}

// Loaded returns the names of the loaded database files.
func (db *DB) Loaded() []string {
	if db == nil {
		return nil
	}

	var names []string
	for _, f := range [...]struct {
		name string
		r    *maxminddb.Reader
	}{
		{CountryFile, db.country},
		{CityFile, db.city},
		{AsnFile, db.asn},
	} {
		if f.r != nil {
			names = append(names, f.name)
		}
	}

	return names
}

// Fill sets the location of ip on dst, through the setters it
// implements (CountrySetter, CitySetter, AsnSetter, AsnOrgSetter). The
// fields not found, or of a database not loaded, are left unset.
func (db *DB) Fill(ip string, dst any) {
	if db == nil {
		return
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return
	}
	addr = addr.Unmap()

	db.fillCountry(addr, dst)
	db.fillCity(addr, dst)
	db.fillAsn(addr, dst)
}

func (db *DB) fillCountry(addr netip.Addr, dst any) {
	s, ok := dst.(CountrySetter)
	if !ok {
		return
	}

	// the City database holds the countries too
	r := db.country
	if r == nil {
		r = db.city
	}
	if iso := lookupString(r, addr, "country", "iso_code"); iso != "" {
		s.SetCountry(iso)
	}
}

func (db *DB) fillCity(addr netip.Addr, dst any) {
	s, ok := dst.(CitySetter)
	if !ok {
		return
	}

	if name := lookupString(
		db.city, addr, "city", "names", "en"); name != "" {
		s.SetCity(name)
	}
}

func (db *DB) fillAsn(addr netip.Addr, dst any) {
	num, wantNum := dst.(AsnSetter)
	org, wantOrg := dst.(AsnOrgSetter)
	if db.asn == nil || !wantNum && !wantOrg {
		return
	}

	var rec asnRecord
	if err := db.asn.Lookup(addr).Decode(&rec); err != nil {
		return
	}

	if wantNum && rec.Number != 0 {
		num.SetAsn(int64(rec.Number))
	}
	if wantOrg && rec.Org != "" {
		org.SetAsnOrg(rec.Org)
	}
}

// open opens the database name of fsys, or returns nil when absent.
func open(fsys fs.FS, name string) (*maxminddb.Reader, error) {
	b, err := fs.ReadFile(fsys, name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("geoip: read %s: %w", name, err)
	}

	r, err := maxminddb.OpenBytes(b)
	if err != nil {
		return nil, fmt.Errorf("geoip: open %s: %w", name, err)
	}

	return r, nil
}

// lookupString returns the string at path of the addr record of r, or
// "" when r is nil or the record has none.
func lookupString(
	r *maxminddb.Reader, addr netip.Addr, path ...any,
) string {
	if r == nil {
		return ""
	}

	var s string
	if err := r.Lookup(addr).DecodePath(&s, path...); err != nil {
		return ""
	}

	return s
}
