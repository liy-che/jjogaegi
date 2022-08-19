package interceptors

import (
	"os"
	"testing"

	"github.com/liy-che/jjogaegi/pkg"
	mocks "github.com/liy-che/jjogaegi/testing"
	"github.com/stretchr/testify/assert"
)

func TestKrDictEnhance(t *testing.T) {
	ts := mocks.NewKrdictMockServer()
	defer ts.Close()
	options := map[string]string{
		pkg.OPT_KRDICT_API_URL: ts.URL,
		pkg.OPT_KRDICT_API_KEY: os.Getenv("KRDICT_API_KEY"),
	}

	for _, expectedItem := range xmlTestItems {
		if expectedItem.Hangul == "참기름" {
			t.Skip()
		}

		t.Run(expectedItem.Hangul, func(tr *testing.T) {
			actualItem := &pkg.Item{
				ExternalID: expectedItem.ExternalID,
			}

			err := KrDictEnhance(actualItem, options)
			assert.NoError(t, err)
			assert.Equal(tr, expectedItem, actualItem)
		})
	}
}
