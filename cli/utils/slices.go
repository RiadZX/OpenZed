package utils

func ByteSliceToUint16Slice(bytes []byte) []uint16 {
	uint16s := make([]uint16, len(bytes))
	//convert each byte to a uint16
	for i, b := range bytes {
		uint16s[i] = uint16(b)
	}
	return uint16s
}
