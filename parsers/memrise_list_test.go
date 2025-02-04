package parsers

import (
	"context"
	"os"
	"testing"

	"github.com/liy-che/jjogaegi/pkg"
	"github.com/stretchr/testify/assert"
)

func TestParseMemriseList(t *testing.T) {
	in, err := os.Open("../testing/fixtures/memrise_list.txt")
	assert.Nil(t, err)
	items := make(chan *pkg.Item, 100)
	ParseMemriseList(context.Background(), in, items, map[string]string{})
	assert.Equal(t, &pkg.Item{Hangul: "남성", Def: pkg.Translation{English: "a man (not 남자)"}}, <-items)
	assert.Equal(t, &pkg.Item{Hangul: "너희", Def: pkg.Translation{English: "you, you guys"}}, <-items)
}
