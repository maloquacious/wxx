// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package h2025v1

import (
	"bytes"
	"fmt"

	"github.com/maloquacious/wxx"
)

// decodeNotes copies each <note> (with its <notetext> CDATA body) into the
// domain map.
func decodeNotes(src Notes_t, w *wxx.Map_t) error {
	var err error
	for _, note := range src.Notes {
		wNote := &wxx.Note_t{
			Key:       note.Key,
			ViewLevel: note.ViewLevel,
			X:         note.X,
			Y:         note.Y,
			Filename:  note.Filename,
			Parent:    note.Parent,
			Title:     note.Title,
			IsGMOnly:  note.IsGMOnly,
			NoteText:  note.NoteText,
		}
		if wNote.Color, err = decodeRgba(note.Color); err != nil {
			return fmt.Errorf("note.color: %w", err)
		}
		w.Notes = append(w.Notes, wNote)
	}
	return nil
}

func encodeNotes(notes []*wxx.Note_t, wb *bytes.Buffer) error {
	wb.WriteString("<notes>\n")
	for _, note := range notes {
		if err := encodeNote(note, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</notes>\n")
	return nil
}

func encodeNote(note *wxx.Note_t, wb *bytes.Buffer) error {
	wb.WriteString("<note")
	wb.WriteString(fmt.Sprintf(" key=%q", note.Key))
	wb.WriteString(fmt.Sprintf(" viewLevel=%q", note.ViewLevel))
	wb.WriteString(fmt.Sprintf(" x=%q", floats(note.X)))
	wb.WriteString(fmt.Sprintf(" y=%q", floats(note.Y)))
	wb.WriteString(fmt.Sprintf(" filename=%q", note.Filename))
	wb.WriteString(fmt.Sprintf(" parent=%q", note.Parent))
	wb.WriteString(fmt.Sprintf(" color=%q", rgbans(note.Color))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" title=%q", note.Title))
	wb.WriteString(fmt.Sprintf(" isGMOnly=%q", bools(note.IsGMOnly)))
	wb.WriteString(">")
	// notetext is CDATA HTML; emit it verbatim so the round-trip preserves it.
	wb.WriteString("<notetext><![CDATA[")
	wb.WriteString(note.NoteText)
	wb.WriteString("]]></notetext>")
	wb.WriteString("</note>\n")
	return nil
}
