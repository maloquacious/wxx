// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

//import (
//	"bytes"
//	"compress/gzip"
//	"encoding/binary"
//	"encoding/xml"
//	"errors"
//	"github.com/maloquacious/wxx/xml_1_0_0"
//	"io"
//	"unicode/utf16"
//	"unicode/utf8"
//)
//
////// UnmarshalJSON implements the json.Unmarshaler interface.
////// It returns an error if there is an error unmarshalling the input.
////func (m *Map) UnmarshalJSON(input []byte) error {
////	return ErrNotImplemented
////}
//
//// Unmarshal accepts a slice of bytes from a Worldographer file and returns a *Map.
//// An error is returned if issues are encountered with running gunzip, reading, or converting the data.
//func Unmarshal(data []byte) (*Map, error) {
//	// gunzip the data
//	input, err := gunzip(data)
//	if err != nil {
//		return nil, errors.Join(ErrGUnZipFailed, err)
//	}
//	// the input should start with the UTF-16 BOM
//	if !verifyBOM(input) {
//		return nil, ErrMissingBOM
//	}
//	// remove the bom before proceeding
//	input = input[2:]
//	// verify XML UTF-16 encoding header
//	if !verifyXMLEncodingHeader(input) {
//		return nil, ErrInvalidEncodingHeader
//	}
//	// remove the header before converting
//	_, input = stripXMLEncodingHeader(input)
//	// convert from UTF-16 to UTF-8
//	input, err = utf16to8(input)
//	if err != nil {
//		return nil, err
//	}
//	// extract the application version number so that we can find the right unmarshaler
//	appVersion, err := fetchWorldographerApplicationVersion(input)
//	if err != nil {
//		return nil, err
//	}
//	// find the right unmarshaler
//	unmarshaler, err := fetchUnmarshaler(appVersion)
//	if err != nil {
//		return nil, err
//	}
//	// unmarshal to *Map
//	return unmarshaler(input)
//}
//
//// Valid returns true if the data meets the following conditions:
////  1. It is a Worldographer file
////  2. It has a supported Worldographer application version
////  3. The data can be unmarshalled to a *Map without errors
//func Valid(data []byte) bool {
//	_, err := Unmarshal(data)
//	return err == nil
//}
//
//// fetchWorldographerApplicationVersion attempts to retrieve the version information from the input.
//// If successful, it returns the version information. Otherwise, it returns nil and an error.
//func fetchWorldographerApplicationVersion(input []byte) (string, error) {
//	// read the version from the xml data
//	// warning: the XML unmarshal is aggressive about finding the version string.
//	// it might search pretty deep into the input to find a match.
//	var version struct {
//		Version string `xml:"version,attr"`
//	}
//	err := xml.Unmarshal(input, &version)
//	if err != nil {
//		return "", err
//	} else if len(version.Version) == 0 {
//		return "", ErrMissingVersion
//	}
//	return version.Version, nil
//}
//
//// gunzip returns the uncompressed data or an error.
//func gunzip(input []byte) ([]byte, error) {
//	// create a new gzip reader to process the source
//	gzr, err := gzip.NewReader(bytes.NewReader(input))
//	if err != nil {
//		return nil, err
//	}
//	defer func(gzr *gzip.Reader) {
//		_ = gzr.Close() // ignore errors
//	}(gzr)
//	return io.ReadAll(gzr)
//}
//
//// stripBOM returns the BOM and the remainder of the input.
//// If there is no BOM, nil and the original slice are returned.
//// This function never allocates additional storage.
//func stripBOM(input []byte) ([]byte, []byte) {
//	if !bytes.HasPrefix(input, []byte{0xfe, 0xff}) {
//		return nil, input
//	}
//	return input[:2], input[2:]
//}
//
//// stripXMLEncodingHeader returns the header and the remainder of the input.
//// If there is no header, nil and the original slice are returned.
//// This function never allocates storage.
//func stripXMLEncodingHeader(input []byte) ([]byte, []byte) {
//	// verify the xml header, which should be <?xml version='1.0' encoding='utf-16'?>
//	xmlHeader := []byte{
//		0x00, 0x3c, 0x00, 0x3f, 0x00, 0x78, 0x00, 0x6d, 0x00, 0x6c,
//		0x00, 0x20, 0x00, 0x76, 0x00, 0x65, 0x00, 0x72, 0x00, 0x73,
//		0x00, 0x69, 0x00, 0x6f, 0x00, 0x6e, 0x00, 0x3d, 0x00, 0x27,
//		0x00, 0x31, 0x00, 0x2e, 0x00, 0x30, 0x00, 0x27, 0x00, 0x20,
//		0x00, 0x65, 0x00, 0x6e, 0x00, 0x63, 0x00, 0x6f, 0x00, 0x64,
//		0x00, 0x69, 0x00, 0x6e, 0x00, 0x67, 0x00, 0x3d, 0x00, 0x27,
//		0x00, 0x75, 0x00, 0x74, 0x00, 0x66, 0x00, 0x2d, 0x00, 0x31,
//		0x00, 0x36, 0x00, 0x27, 0x00, 0x3f, 0x00, 0x3e, 0x00, 0x0a,
//	}
//	if !bytes.HasPrefix(input, xmlHeader) {
//		return nil, input
//	}
//	// remove the header and the rest as separate slices
//	return input[:len(xmlHeader)], input[len(xmlHeader):]
//}
//
//func unmarshal_v1_0_0(input []byte) (*Map, error) {
//	src, err := xml_1_0_0.UnmarshalXML(input)
//	if err != nil {
//		return nil, err
//	}
//	w, err := xml_1_0_0_to_Map(src)
//	if err != nil {
//		return nil, err
//	}
//	return w, nil
//}
//
//// utf16to8 converts the input from UTF-16 to UTF-8, returning the first error encountered.
//// Uses a lot of temporary memory and assumes that the caller has verified that the data is encoded correctly.
//func utf16to8(input []byte) ([]byte, error) {
//	if len(input)%2 != 0 {
//		// UTF-16 data must contain an even number of bytes
//		return nil, ErrMissingFinalByte
//	}
//	// convert the slice of byte to a slice of uint16
//	chars := make([]uint16, len(input)/2)
//	if err := binary.Read(bytes.NewReader(input), binary.BigEndian, &chars); err != nil {
//		return nil, err
//	}
//	// create a buffer for the results
//	dst := bytes.Buffer{}
//	// convert the UTF-16 to runes, then to UTF-8 bytes
//	var utfBuffer [utf8.UTFMax]byte
//	for _, r := range utf16.Decode(chars) {
//		utf8Size := utf8.EncodeRune(utfBuffer[:], r)
//		dst.Write(utfBuffer[:utf8Size])
//	}
//	// finally, return the results
//	return dst.Bytes(), nil
//}
//
//// verifyBOM returns true if the input starts with the expected BOM.
//func verifyBOM(input []byte) bool {
//	bom, _ := stripBOM(input)
//	return bom != nil
//}
//
//// verifyXMLEncodingHeader returns true if the input starts with the expected encoding header.
//func verifyXMLEncodingHeader(input []byte) bool {
//	header, _ := stripXMLEncodingHeader(input)
//	return header != nil
//}
