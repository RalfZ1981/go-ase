package libdsn

import (
	"fmt"
	"net/url"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

// ParseDSN parses a DSN into a DsnInfo struct.
//
// Accepted DSNs are either in URI or simple form:
// URI: ase://user:pass@host:port?key=val
// Simple: username=user password=password host=host port=port key=val
//
// To use special characters in your DSN use the simple form.
//
// When using the simple form values containing whitespaces must be
// quoted with double or single quotation marks.
//		username=user password="a password" host=host port=port
//		username=user password='a password' host=host port=port
//
// The DSN is validated using the struct tags and validator.
// Validation errors from validator are returned as-is for further
// processing.
func ParseDSN(dsn string) (*DsnInfo, error) {
	var dsnInfo *DsnInfo
	var err error

	// Parse DSN
	if strings.HasPrefix(dsn, "ase:/") {
		dsnInfo, err = parseDsnUri(dsn)
	} else {
		dsnInfo, err = parseDsnSimple(dsn)
	}
	if err != nil {
		return nil, err
	}

	var filterFn validator.FilterFunc = filterNoUserStoreKey
	if dsnInfo.Userstorekey != "" {
		filterFn = filterUserStoreKey
	}

	v := validator.New()
	err = v.StructFiltered(dsnInfo, filterFn)
	if err != nil {
		return nil, err
	}

	return dsnInfo, nil
}

// filterUserStoreKey is the validator.FilterFunc for a DsnInfo struct
// with Userstorekey set.
func filterUserStoreKey(ns []byte) bool {
	switch string(ns) {
	case "DsnInfo.Username":
		return true
	case "DsnInfo.Password":
		return true
	case "DsnInfo.Database":
		return true
	case "DsnInfo.Host":
		return true
	case "DsnInfo.Port":
		return true
	}
	return false
}

// filterNoUserStoreKey is the validator.FilterFunc for a DsnInfo struct
// with Userstorekey unset.
func filterNoUserStoreKey(ns []byte) bool {
	return string(ns) == "DsnInfo.Userstorekey"
}

// parseDsnUri parses a DSN in URI form and returns the resulting
// DsnInfo.
func parseDsnUri(dsn string) (*DsnInfo, error) {
	url, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN using url.Parse: %v", err)
	}

	dsni := &DsnInfo{
		Host:         url.Hostname(),
		Port:         url.Port(),
		ConnectProps: url.Query(),
	}

	// Assume that `astring` in the DSN ase://astring@hostname/ is the
	// userstorekey. This is parsed as the username by url.Parse.
	if url.User != nil {
		username := url.User.Username()
		password, ok := url.User.Password()

		if ok {
			// ase://username:password@hostname/
			dsni.Username = username
			dsni.Password = password
		} else {
			// ase://userstorekey@hostname/
			dsni.Userstorekey = username
		}
	}

	return dsni, nil
}

// parseDsnSimple parses a DSN in the simple form and returns the
// resulting DsnInfo without checking for missing values.
func parseDsnSimple(dsn string) (*DsnInfo, error) {
	dsni := &DsnInfo{
		ConnectProps: url.Values{},
	}

	// Valid quotation marks to detect values with whitespaces
	quotations := []byte{'\'', '"'}

	// Split the DSN on whitespace - any quoted values containing
	// whitespaces will be concatenated in the first step in the loop.
	dsnS := strings.Split(dsn, " ")

	// Prepare a tag to field map
	tagToField := dsni.tagToField(true)

	for len(dsnS) > 0 {
		var part string
		part, dsnS = dsnS[0], dsnS[1:]

		// If the value starts with a quotation mark consume more parts
		// until the quotation is finished.
		for _, quot := range quotations {
			if !strings.Contains(part, "="+string(quot)) {
				continue
			}

			for part[len(part)-1] != quot {
				part = strings.Join([]string{part, dsnS[0]}, " ")
				dsnS = dsnS[1:]
			}
			break
		}

		partS := strings.SplitN(part, "=", 2)
		if len(partS) != 2 {
			return nil, fmt.Errorf("Recognized DSN part does not contain key/value parts: %s", partS)
		}

		key, value := partS[0], partS[1]

		// Remove quotation from value
		if value != "" {
			for _, quot := range quotations {
				if value[0] == quot && value[len(value)-1] == quot {
					value = value[1 : len(value)-1]
				}
			}
		}

		if field, ok := tagToField[key]; ok {
			field.SetString(value)
		} else {
			dsni.ConnectProps.Add(key, value)
		}
	}

	return dsni, nil
}
