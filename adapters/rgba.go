// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

//// rgbaToXmlAttr converts an RGBA struct into an XML attribute string.
//// RGBA struct contains four fields, each representing Red, Green, Blue and Alpha respectively.
//// Each field is a float and the rgbaToXmlAttr function formats them as a string, separated by commas.
//// If the provided rgba pointer is nil, it defaults to "0.0,0.0,0.0,1.0".
////
//// It uses the FToXF function to format the float values into an XML-friendly format.
////
//// Parameters:
//// - rgba: a pointer to a RGBA struct. Can be nil.
////
//// Returns:
//// - A XML attribute string representing the rgba. If rgba is nil, returns "0.0,0.0,0.0,1.0"
//func rgbaToXmlAttr(rgba *models.RGBA_t) string {
//	if rgba == nil {
//		return "0.0,0.0,0.0,1.0"
//	}
//	return fmt.Sprintf("%s,%s,%s,%s",
//		FToXF(rgba.R),
//		FToXF(rgba.G),
//		FToXF(rgba.B),
//		FToXF(rgba.A))
//}
//
//func rgbaToNullableXmlAttr(rgba *models.RGBA_t) string {
//	s := rgbaToXmlAttr(rgba)
//	if s == "0.0,0.0,0.0,1.0" {
//		s = "null"
//	}
//	return s
//}
