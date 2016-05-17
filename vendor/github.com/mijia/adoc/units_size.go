package adoc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Copied from github.com/docker/docker/pkg/units/size.go

// See: http://en.wikipedia.org/wiki/Binary_prefix
const (
	// Decimal

	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

type unitMap map[string]int64

var (
	decimalMap = unitMap{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	binaryMap  = unitMap{"k": KiB, "m": MiB, "g": GiB, "t": TiB, "p": PiB}
	sizeRegex  = regexp.MustCompile(`^(\d+)([kKmMgGtTpP])?[bB]?$`)
)

var (
	binaryMap2 = unitMap{"KiB": KiB, "MiB": MiB, "GiB": GiB, "TiB": TiB, "PiB": PiB, "B": 1}
)

var decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var binaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

// HumanSize returns a human-readable approximation of a size
// using SI standard (eg. "44kB", "17MB")
func HumanSize(size float64) string {
	return intToString(float64(size), 1000.0, decimapAbbrs)
}

func BytesSize(size float64) string {
	return intToString(size, 1024.0, binaryAbbrs)
}

func intToString(size, unit float64, _map []string) string {
	i := 0
	for size >= unit {
		size = size / unit
		i++
	}
	return fmt.Sprintf("%.4g %s", size, _map[i])
}

// ParseBytesSize parse a string like "1.795 GiB" into an int64
func ParseBytesSize(size string) (int64, error) {
	parts := strings.SplitN(strings.TrimSpace(size), " ", 2)
	memory, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return -1, err
	}
	if factor, ok := binaryMap2[parts[1]]; !ok {
		return -1, fmt.Errorf("Invalid size unit %q for %q", parts[1], size)
	} else {
		return int64(memory * float64(factor)), nil
	}
}

// FromHumanSize returns an integer from a human-readable specification of a
// size using SI standard (eg. "44kB", "17MB")
func FromHumanSize(size string) (int64, error) {
	return parseSize(size, decimalMap)
}

// RAMInBytes parses a human-readable string representing an amount of RAM
// in bytes, kibibytes, mebibytes, gibibytes, or tebibytes and
// returns the number of bytes, or -1 if the string is unparseable.
// Units are case-insensitive, and the 'b' suffix is optional.
func RAMInBytes(size string) (int64, error) {
	return parseSize(size, binaryMap)
}

// Parses the human-readable size string into the amount it represents
func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	fmt.Println(matches)
	if len(matches) != 3 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	size, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[2])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= mul
	}

	return size, nil
}
