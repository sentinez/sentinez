# GeoLite2 databases

Files placed here are compiled into the binary (`sentinez.GeoIP`, in
`sentinez.go`), and the edge then fills, through `pkg/geoip`, the client
location of the logs that ask for it. Any subset works; a missing file just leaves its
fields empty.

| File | Fills |
|---|---|
| `GeoLite2-Country.mmdb` | country ISO code |
| `GeoLite2-City.mmdb` | city name, and the country when the Country file is absent |
| `GeoLite2-ASN.mmdb` | AS number and organization |

Download them from MaxMind (free account, license key):

```sh
curl -fsSL -o GeoLite2-City.tar.gz \
  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$MAXMIND_LICENSE_KEY&suffix=tar.gz"
tar -xzf GeoLite2-City.tar.gz --strip-components=1 --wildcards '*.mmdb'
```

The `.mmdb` files are git-ignored: the GeoLite2 license does not allow
redistributing them in this repository.
