package formatters

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/liy-che/jjogaegi/pkg"
)

func formatXSV(ctx context.Context, items <-chan *pkg.Item, w io.Writer, options map[string]string, delim rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
	for item := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var firstExample pkg.Translation
		if len(item.Examples) > 0 {
			firstExample = item.Examples[0]
		}

		var secondExample pkg.Translation
		if len(item.Examples) > 1 {
			secondExample = item.Examples[1]
		}

		cw.Write([]string{
			item.NoteID,
			item.ExternalID,
			item.Hangul,
			item.Hanja,
			item.Def.Korean,
			item.Def.English,
			item.Pronunciation,
			item.AudioTag,
			item.ImageTag,
			item.Grade,
			item.Antonym,
			firstExample.Korean,
			firstExample.English,
			secondExample.Korean,
			secondExample.English,
		})
		cw.Flush()
	}
	return nil
}
