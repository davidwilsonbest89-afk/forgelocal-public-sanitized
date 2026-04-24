package fingerprint

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// AdjustToLocal adjusts fingerprint geo fields to match the actual public IP
func AdjustToLocal(fp map[string]any) {
	// Always check actual public IP first (handles VPN/WireGuard scenarios)
	if country, tz := detectPublicIPGeo(); country != "" {
		geo, ok := countryToGeo[country]
		if ok {
			fp["timezone"] = tz
			fp["locale:language"] = geo.Language[:2]
			fp["locale:region"] = country
			fp["navigator.language"] = geo.Language
			fp["navigator.languages"] = geo.Languages
			fp["headers.Accept-Language"] = buildAcceptLanguage(geo.Languages)
			return
		}
	}

	// Fallback: use local system timezone
	tz := detectLocalTimezone()
	country := timezoneToCountry(tz)
	if geo, ok := countryToGeo[country]; ok {
		fp["timezone"] = tz
		fp["locale:language"] = geo.Language[:2]
		fp["locale:region"] = country
		fp["navigator.language"] = geo.Language
		fp["navigator.languages"] = geo.Languages
		fp["headers.Accept-Language"] = buildAcceptLanguage(geo.Languages)
	} else {
		fp["timezone"] = tz
	}
}

// detectPublicIPGeo queries a free API to get the actual public IP's country and timezone
func detectPublicIPGeo() (country, timezone string) {
	client := &http.Client{Timeout: 5 * time.Second}
	return queryIPGeo(client)
}

// DetectProxyGeoResult returns timezone and locale detected through the proxy.
func DetectProxyGeoResult(proxyType, host string, port int, username, password string) (timezone, locale string) {
	client := buildProxyClient(proxyType, host, port, username, password)
	country, tz := queryIPGeo(client)
	if country == "" {
		return "America/New_York", "en-US"
	}
	geo, ok := countryToGeo[country]
	if !ok {
		return tz, "en-US"
	}
	return tz, geo.Language
}

// DetectLocalGeoResult returns timezone and locale from the machine's public IP.
func DetectLocalGeoResult() (timezone, locale string) {
	client := &http.Client{Timeout: 5 * time.Second}
	country, tz := queryIPGeo(client)
	if country == "" {
		tz = detectLocalTimezone()
		country = timezoneToCountry(tz)
	}
	geo, ok := countryToGeo[country]
	if !ok {
		return tz, "en-US"
	}
	return tz, geo.Language
}

func queryIPGeo(client *http.Client) (country, timezone string) {
	resp, err := client.Get("http://ip-api.com/json/?fields=countryCode,timezone")
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	var result struct {
		CountryCode string `json:"countryCode"`
		Timezone    string `json:"timezone"`
	}
	if json.NewDecoder(resp.Body).Decode(&result) == nil {
		return result.CountryCode, result.Timezone
	}
	return "", ""
}

func buildProxyClient(proxyType, host string, port int, username, password string) *http.Client {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	if proxyType == "socks5" {
		var auth *proxy.Auth
		if username != "" {
			auth = &proxy.Auth{User: username, Password: password}
		}
		dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
		if err == nil {
			return &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
			}
		}
	}

	// HTTP/HTTPS proxy
	proxyURL := &url.URL{Scheme: proxyType, Host: addr}
	if username != "" {
		proxyURL.User = url.UserPassword(username, password)
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
	}
}

func detectLocalTimezone() string {
	// Method 1: Go's time.Local
	zone, _ := time.Now().Zone()
	
	// Method 2: macOS/Linux system timezone (more accurate)
	if out, err := exec.Command("readlink", "/etc/localtime").Output(); err == nil {
		// /etc/localtime -> /var/db/timezone/zoneinfo/Asia/Taipei
		parts := strings.Split(strings.TrimSpace(string(out)), "/zoneinfo/")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	
	// Method 3: macOS systemsetup
	if out, err := exec.Command("systemsetup", "-gettimezone").Output(); err == nil {
		// "Time Zone: Asia/Taipei"
		s := strings.TrimSpace(string(out))
		if idx := strings.Index(s, ": "); idx >= 0 {
			return s[idx+2:]
		}
	}

	// Fallback: map Go zone abbreviation
	return zoneAbbrevToTZ(zone)
}

var timezoneCountryMap = map[string]string{
	"Asia/Taipei":       "TW",
	"Asia/Tokyo":        "JP",
	"Asia/Seoul":        "KR",
	"Asia/Shanghai":     "CN",
	"Asia/Hong_Kong":    "HK",
	"Asia/Singapore":    "SG",
	"Asia/Bangkok":      "TH",
	"Asia/Ho_Chi_Minh":  "VN",
	"Asia/Jakarta":      "ID",
	"Asia/Manila":       "PH",
	"Asia/Kuala_Lumpur": "MY",
	"Asia/Kolkata":      "IN",
	"Asia/Dubai":        "AE",
	"Asia/Jerusalem":    "IL",
	"Europe/London":     "GB",
	"Europe/Berlin":     "DE",
	"Europe/Paris":      "FR",
	"Europe/Rome":       "IT",
	"Europe/Madrid":     "ES",
	"Europe/Amsterdam":  "NL",
	"Europe/Stockholm":  "SE",
	"Europe/Warsaw":     "PL",
	"Europe/Istanbul":   "TR",
	"Europe/Kyiv":       "UA",
	"Europe/Moscow":     "RU",
	"America/New_York":      "US",
	"America/Chicago":       "US",
	"America/Denver":        "US",
	"America/Los_Angeles":   "US",
	"America/Toronto":       "CA",
	"America/Sao_Paulo":     "BR",
	"America/Mexico_City":   "MX",
	"Australia/Sydney":      "AU",
}

func timezoneToCountry(tz string) string {
	if c, ok := timezoneCountryMap[tz]; ok {
		return c
	}
	// Guess from timezone prefix
	if strings.HasPrefix(tz, "America/") {
		return "US"
	}
	if strings.HasPrefix(tz, "Europe/") {
		return "DE"
	}
	if strings.HasPrefix(tz, "Asia/") {
		return "JP"
	}
	return "US"
}

func zoneAbbrevToTZ(abbrev string) string {
	m := map[string]string{
		"CST": "Asia/Taipei", // Could be US Central, but on a TW machine it's likely Taipei
		"JST": "Asia/Tokyo",
		"KST": "Asia/Seoul",
		"EST": "America/New_York",
		"PST": "America/Los_Angeles",
		"MST": "America/Denver",
		"GMT": "Europe/London",
		"CET": "Europe/Berlin",
		"IST": "Asia/Kolkata",
		"AEST": "Australia/Sydney",
	}
	if tz, ok := m[abbrev]; ok {
		return tz
	}
	return "America/New_York"
}
