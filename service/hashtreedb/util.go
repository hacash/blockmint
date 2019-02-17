package hashtreedb

func copyCoverBytes(base []byte, cover []byte, seek int) {
	for i := seek; i < len(cover)+seek && i < len(base)+seek; i++ {
		base[i] = cover[i]
	}
}

///////////////////////////////////////////

func ReverseHashOrder(hash []byte) []byte {
	var length = len(hash)
	var hsdt = make([]byte, length)
	copy(hsdt, hash)
	for i := 0; i < length/2; i++ {
		hsdt[i], hsdt[length-i-1] = hsdt[length-i-1], hsdt[i]
	}
	return hsdt
}
