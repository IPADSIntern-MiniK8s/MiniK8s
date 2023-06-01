package resourceutils


import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)


// ParseQuantity parses the quantity string to float64
func ParseQuantity(str string) (float64, error) {
	var quantity float64
	var err error

	// use the regex to separate the quantity and unit
	// for example, 1m -> 1, m
	re := regexp.MustCompile(`^([\d\.]+)([a-zA-Z]*)$`)
	matches := re.FindStringSubmatch(str)

	if len(matches) == 3 {
		// parse the quantity part to float64
		quantity, err = strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse quantity: %s", err.Error())
		}

		// convert the unit part to standard unit, for example, m -> 1/100
		log.Info("[ParseQuantity] matches[2]: ", matches[2])
		switch matches[2] {
		case "m":
			quantity /= 1000
		case "Ki":
			quantity *= 1024
		case "Mi":
			quantity *= 1024 * 1024
		case "M":
			quantity *= 1000 * 1000
		case "K":
			quantity *= 1000
		default:
			log.Info("[ParseQuantity] invalid unit: ", matches[2])
		}

		return quantity, nil
	}

	return 0, fmt.Errorf("invalid quantity string: %s", str)
}


func PackQuantity(quantity float64, unit string) string {
	switch unit {
	case "m":
		quantity *= 1000
	case "Ki":
		quantity /= 1024
	case "Mi":
		quantity /= 1024 * 1024
	case "M":
		quantity /= 1000 * 1000
	case "K":
		quantity /= 1000
	default:
		log.Info("[PackQuantity] invalid unit: ", unit)
	}
	return fmt.Sprintf("%f%s", quantity, unit)
}


func GetUnit(quantity string) string {
	re := regexp.MustCompile(`^([\d\.]+)([a-zA-Z]*)$`)
	matches := re.FindStringSubmatch(quantity)
	if len(matches) == 3 {
		return matches[2]
	}
	return ""
}

