package zap

const hextable = "0123456789ABCDEF"

func hexEncode(dst []byte, src []byte) []byte {
	dst = append(dst, "0x"...)
	for _, v := range src {
		dst = append(dst, hextable[v>>4])
		dst = append(dst, hextable[v&0x0F])
	}
	return dst
}
