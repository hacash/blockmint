package util

func BytesCopyCover(base []byte, cover []byte, seek int) {
	var i = 0
	for true {
		if seek >= len(base) || i >= len(cover) {
			break
		}
		base[seek] = cover[i]
		seek++
		i++
	}
}
