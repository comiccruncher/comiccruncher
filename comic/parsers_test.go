package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisYearlyAggregateDeserializerDeserialize(t *testing.T) {
	s := "1948:1:2;1949:2:0;1950:3:0"
	p := comic.RedisYearlyAggregateDeserializer{}
	vals := p.Deserialize(s)

	expected := []comic.YearlyAggregate{
		{Year: 1948, Main: 1, Alternate: 2},
		{Year: 1949, Main: 2, Alternate: 0},
		{Year: 1950, Main: 3, Alternate: 0},
	}

	assert.NotNil(t, vals)
	assert.ElementsMatch(t, vals, expected)
}
