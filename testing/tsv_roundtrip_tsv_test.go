package testing

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/liy-che/jjogaegi/formatters"
	"github.com/liy-che/jjogaegi/parsers"
	"github.com/liy-che/jjogaegi/pkg"
	"github.com/stretchr/testify/assert"
)

func TestRoundtripTSV(t *testing.T) {
	file, err := os.Open("../testing/fixtures/sample-1.tsv")
	assert.Nil(t, err)

	inBytes, err := ioutil.ReadAll(file)
	assert.Nil(t, err)

	in := bytes.NewBuffer(inBytes)

	items := make(chan *pkg.Item, 100)
	err = parsers.ParseTSV(context.Background(), in, items, map[string]string{})
	assert.Nil(t, err)

	close(items)

	out := &bytes.Buffer{}
	err = formatters.FormatTSV(context.Background(), items, out, map[string]string{})
	assert.Nil(t, err)

	assert.Equal(t, string(inBytes), out.String())
}
