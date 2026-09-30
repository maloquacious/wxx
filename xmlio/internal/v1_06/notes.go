// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/maloquacious/wxx"
)

// decodeNotes copies each <note> (with its <notetext> CDATA body and its
// <location>) into the domain map.
//
// @key states the note's position a second time, as "<viewLevel>,<x>,<y>".
// The domain model does not keep it: the encoder derives it from Location. So a
// key that disagrees with the <location> is refused here rather than dropped,
// because one of the two positions would otherwise be lost without a word.
// Every note Worldographer 2.06, 2.07 and 2.08 wrote in the samples agrees.
//
// A note with no <location> decodes with a nil Location (issue #94); the
// encoder refuses to write it.
func decodeNotes(src Notes_t, w *wxx.Map_t) error {
	var err error
	for i, note := range src.Notes {
		wNote := &wxx.Note_t{
			OriginalViewLevel: note.OriginalViewLevel,
			Filename:          note.Filename,
			Parent:            note.Parent,
			IsWorld:           note.IsWorld,
			IsContinent:       note.IsContinent,
			IsKingdom:         note.IsKingdom,
			IsProvince:        note.IsProvince,
			Title:             note.Title,
			NoteText:          note.NoteText,
		}
		if wNote.Color, err = decodeRgba(note.Color); err != nil {
			return fmt.Errorf("note.color: %w", err)
		}
		if note.Location != nil {
			wNote.Location = &wxx.NoteLocation_t{
				ViewLevel: note.Location.ViewLevel,
				X:         note.Location.X,
				Y:         note.Location.Y,
			}
			if !noteKeyMatches(note.Key, wNote.Location) {
				return errors.Join(wxx.ErrNoteKeyMismatch,
					fmt.Errorf("map/notes/note[%d] (title %q): key %q, location %q", i+1, note.Title, note.Key, noteKey(wNote.Location)))
			}
		}
		w.Notes = append(w.Notes, wNote)
	}
	return nil
}

// noteKey spells a note's @key from its location: "<viewLevel>,<x>,<y>", with
// the coordinates formatted as every other float in the file ("375.0",
// "2343.75").
func noteKey(loc *wxx.NoteLocation_t) string {
	return loc.ViewLevel + "," + floats(loc.X) + "," + floats(loc.Y)
}

// noteKeyMatches reports whether key names the same position as loc. The
// coordinates are compared as numbers, so a key is not refused for spelling a
// number differently from the way this encoder would; it is refused only when
// it names a different place.
func noteKeyMatches(key string, loc *wxx.NoteLocation_t) bool {
	parts := strings.Split(key, ",")
	if len(parts) != 3 || parts[0] != loc.ViewLevel {
		return false
	}
	x, errX := strconv.ParseFloat(parts[1], 64)
	y, errY := strconv.ParseFloat(parts[2], 64)
	return errX == nil && errY == nil && x == loc.X && y == loc.Y
}

// encodeNotes writes <notes>. Every note is checked before anything is
// written: a note with no Location has no position to write, and writing it
// as 0,0 would move it to the map's corner (issue #94).
func encodeNotes(notes []*wxx.Note_t, wb *bytes.Buffer) error {
	for i, note := range notes {
		if note.Location == nil {
			return errors.Join(wxx.ErrNoteWithoutLocation,
				fmt.Errorf("map/notes/note[%d] (title %q): Location is nil", i+1, note.Title))
		}
	}
	wb.WriteString("<notes>\n")
	for _, note := range notes {
		if err := encodeNote(note, wb); err != nil {
			return err
		}
	}
	wb.WriteString("</notes>\n")
	return nil
}

// encodeNote writes one <note> in the order and layout Worldographer 2.06,
// 2.07 and 2.08 use: the start tag, a newline, <notetext>, then <location>.
// @key is derived from Location, never carried, so it cannot go stale.
func encodeNote(note *wxx.Note_t, wb *bytes.Buffer) error {
	wb.WriteString("<note")
	wb.WriteString(fmt.Sprintf(" key=%s", xmlAttr(noteKey(note.Location))))
	wb.WriteString(fmt.Sprintf(" originalViewLevel=%s", xmlAttr(note.OriginalViewLevel)))
	wb.WriteString(fmt.Sprintf(" filename=%s", xmlAttr(note.Filename)))
	wb.WriteString(fmt.Sprintf(" parent=%s", xmlAttr(note.Parent)))
	wb.WriteString(fmt.Sprintf(" color=%s", xmlAttr(rgbans(note.Color)))) // decodeRgba
	wb.WriteString(fmt.Sprintf(" isWorld=%s", xmlAttr(bools(note.IsWorld))))
	wb.WriteString(fmt.Sprintf(" isContinent=%s", xmlAttr(bools(note.IsContinent))))
	wb.WriteString(fmt.Sprintf(" isKingdom=%s", xmlAttr(bools(note.IsKingdom))))
	wb.WriteString(fmt.Sprintf(" isProvince=%s", xmlAttr(bools(note.IsProvince))))
	wb.WriteString(fmt.Sprintf(" title=%s", xmlAttr(note.Title)))
	wb.WriteString(">\n")
	// notetext is CDATA HTML; emit it verbatim so the round-trip preserves it.
	wb.WriteString("<notetext><![CDATA[")
	wb.WriteString(note.NoteText)
	wb.WriteString("]]></notetext>")
	wb.WriteString("<location")
	wb.WriteString(fmt.Sprintf(" viewLevel=%s", xmlAttr(note.Location.ViewLevel)))
	wb.WriteString(fmt.Sprintf(" x=%s", xmlAttr(floats(note.Location.X))))
	wb.WriteString(fmt.Sprintf(" y=%s", xmlAttr(floats(note.Location.Y))))
	wb.WriteString(" />")
	wb.WriteString("</note>\n")
	return nil
}
