package formatters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/liy-che/jjogaegi/pkg"
)

func FormatJSON(ctx context.Context, items <-chan *pkg.Item, w io.Writer, options map[string]string) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	wrote := false
	for item := range items {
		if !wrote {
			fmt.Fprintf(w, "[\n")
		} else {
			fmt.Fprintf(w, ",\n")
		}

		if err := enc.Encode(item); err != nil {
			return err
		}
		wrote = true
	}
	if wrote {
		fmt.Fprintf(w, "]\n")
	}

	return nil
}
