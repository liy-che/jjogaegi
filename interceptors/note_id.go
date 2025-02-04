package interceptors

import (
	"github.com/google/uuid"
	"github.com/liy-che/jjogaegi/pkg"
)

func GenerateNoteId(item *pkg.Item, options map[string]string) error {
	if item.NoteID == "" {
		item.NoteID = uuid.New().String()
	}

	return nil
}
