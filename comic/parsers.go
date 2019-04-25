package comic

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/stringutil"
	"strconv"
	"strings"
)

func parseRedisThumbnails(s string, th *CharacterThumbnails) error {
	imgs := strings.Split(s, "-")
	if len(imgs) != 2 {
		return errors.New("error parsing string " + s)
	}
	splVendorImg, splImg := strings.Split(imgs[0], ";"), strings.Split(imgs[1], ";")
	if len(splImg) != 3 || len(splVendorImg) != 3 {
		return fmt.Errorf("error parsing either strings: %v ; %v", splImg, splVendorImg)
	}
	s, m, l := strings.Split(splImg[0], ":")[1], strings.Split(splImg[1], ":")[1], strings.Split(splImg[2], ":")[1]
	vs, vm, vl := strings.Split(splVendorImg[0], ":")[1], strings.Split(splVendorImg[1], ":")[1], strings.Split(splVendorImg[2], ":")[1]
	if vs != "" || vm != "" || vl != "" {
		th.VendorImage = &ThumbnailSizes{
			Small:  vs,
			Medium: vm,
			Large:  vl,
		}
	}
	if s != "" || m != "" || l != "" {
		th.Image = &ThumbnailSizes{
			Small:  s,
			Medium: m,
			Large:  l,
		}
	}
	return nil
}

func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 64)
	return uint(u), err
}

// YearlyAggregateSerializer serializes a struct into a string.
type YearlyAggregateSerializer interface {
	Serialize(aggregates []YearlyAggregate) string
}

// RedisYearlyAggregateSerializer serializes a struct into a string for Redis storage.
type RedisYearlyAggregateSerializer struct {
}

// Serialize serializes the structs into a string for Redis storage.
func (s *RedisYearlyAggregateSerializer) Serialize(aggregates []YearlyAggregate) string {
	val := ""
	lenAggs := len(aggregates)
	// sets the value in the form of `year:main_count:alt_count;year:main_count:alt_count`
	// this is just a fast, simple, and cheap way to keep them sorted and packed.
	for idx, appearance := range aggregates {
		val += fmt.Sprintf("%d:%d:%d", appearance.Year, appearance.Main, appearance.Alternate)
		// if it's not the last one in the slice
		if idx != lenAggs-1 {
			// add a semicolon!
			val += ";"
		}
	}
	return val
}

// YearlyAggregateDeserializer deserializes a string into a struct.
type YearlyAggregateDeserializer interface {
	Deserialize(val string) []YearlyAggregate
}

// RedisYearlyAggregateDeserializer deserializes a Redis string into a struct.
type RedisYearlyAggregateDeserializer struct {
}

// Deserialize deserializes the string into the yearly aggregates structs.
func (p *RedisYearlyAggregateDeserializer) Deserialize(val string) []YearlyAggregate {
	var yearlyAggregates []YearlyAggregate
	// since we sore the appearances values in the form of `1948:1:2;1949:2:0;1950:3:0`, we need to parse out the values.
	values := strings.Split(val, ";")
	for _, v := range values {
		// now we're at 1948:1:2
		counts := strings.Split(v, ":")
		year := stringutil.MustAtoi(counts[0])
		mainCount := stringutil.MustAtoi(counts[1])
		altCount := stringutil.MustAtoi(counts[2])
		yearlyAggregates = append(yearlyAggregates, YearlyAggregate{Year: year, Main: mainCount, Alternate: altCount})
	}
	return yearlyAggregates
}
