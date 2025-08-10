// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

//var (
//	// versionsMap lists the supported versions of the Worldographer data
//	// along with the marshal/unmarshal function that works with it.
//	versionsMap = map[string]struct {
//		marshal   func(*Map) ([]byte, error)
//		unmarshal func([]byte) (*Map, error)
//	}{
//		"1.73": {nil, unmarshal_v1_0_0},
//		"1.74": {nil, unmarshal_v1_0_0},
//	}
//)
//
//// fetchMarshaler uses the Worldographer application version from the
//func fetchMarshaler(appVersion string) (func(*Map) ([]byte, error), error) {
//	fns, ok := versionsMap[appVersion]
//	if !ok {
//		return nil, ErrUnsupportedVersion
//	}
//	return fns.marshal, nil
//}
//
//// fetchUnmarshaler uses the Worldographer application version from the
//func fetchUnmarshaler(appVersion string) (func([]byte) (*Map, error), error) {
//	fns, ok := versionsMap[appVersion]
//	if !ok {
//		return nil, ErrUnsupportedVersion
//	}
//	return fns.unmarshal, nil
//}
