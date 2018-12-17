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
		return errors.New(fmt.Sprintf("error parsing either strings: %v ; %v", splImg, splVendorImg))
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

// parseYearlyAggregates parses the string value of the redis value into a yearly aggregate.
func parseRedisYearlyAggregates(value string) []YearlyAggregate {
	var yearlyAggregates []YearlyAggregate
	// since we sore the appearances values in the form of `1948:1;1949:2;1950:3`, we need to parse out the values.
	values := strings.Split(value, ";")
	for _, val := range values {
		idx := strings.Index(val, ":")
		year := stringutil.MustAtoi(val[:idx])
		count := stringutil.MustAtoi(val[idx+1:])
		yearlyAggregates = append(yearlyAggregates, YearlyAggregate{Year: year, Count: count})
	}
	return yearlyAggregates
}
