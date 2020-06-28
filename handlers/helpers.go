package handlers

import "regexp"

var rgxFilePrefix = regexp.MustCompile("^.{24}_{1}")

// stripDocID strips the leading document id prefix from a stored file's filename
func stripDocID(fname string) string {
	return rgxFilePrefix.ReplaceAllString(fname, "")
}
