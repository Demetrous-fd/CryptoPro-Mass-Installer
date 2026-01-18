package core

import (
	"fmt"
	"strings"
	"unicode/utf8"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"golang.org/x/text/encoding/charmap"
)

type ContainerName struct {
	Normal      string
	Windows1251 string
}

func FormatNewName(pattern string, certificate *cades.GostCertificate) ContainerName {
	result := ContainerName{}
	names := []string{
		"common_name",
		"surname",
		"country_name",
		"locality_name",
		"state_or_province_name",
		"street_address",
		"organization_name",
		"organizational_unit_name",
		"title",
		"telephone_number",
		"name",
		"given_name",
		"initials",
		"pseudonym",
		"email_address",
	}

	symbols := map[string]string{
		"expire_after":  "",
		"expire_before": "",
	}
	for _, name := range names {
		symbols[fmt.Sprintf("subject.%s", name)] = name
		symbols[fmt.Sprintf("issuer.%s", name)] = name
	}

	for key, value := range symbols {
		if !strings.Contains(pattern, "#"+key) {
			continue
		}

		switch strings.Split(key, ".")[0] {
		case "subject":
			if v, ok := certificate.Subject[value]; ok {
				value = v
			} else {
				if value == "initials" {
					if v, ok = certificate.Subject["given_name"]; ok {
						value = ""
						for _, i := range strings.Split(v, " ") {
							r, _ := utf8.DecodeRuneInString(i)
							value += string(r) + "."
						}
					} else {
						value = "None"
					}
				} else {
					value = "None"
				}
			}

		case "issuer":
			if v, ok := certificate.Issuer[value]; ok {
				value = v
			} else {
				if value == "initials" {
					if v, ok = certificate.Issuer["given_name"]; ok {
						value = ""
						for _, i := range strings.Split(v, " ") {
							r, _ := utf8.DecodeRuneInString(i)
							value += string(r) + "."
						}
					} else {
						value = "None"
					}
				} else {
					value = "None"
				}
			}
		case "expire_after":
			value = certificate.NotAfter.Format("02.01.2006")
		case "expire_before":
			value = certificate.NotBefore.Format("02.01.2006")
		default:
		}
		pattern = strings.ReplaceAll(pattern, "#"+key, value)
	}
	result.Normal = pattern

	encoder := charmap.Windows1251.NewEncoder()
	cp1251String, err := encoder.String(pattern)
	if err != nil {
		return ContainerName{}
	}
	result.Windows1251 = cp1251String

	return result
}
