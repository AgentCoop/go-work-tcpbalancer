package imgresize

type ImgType int

const (
	Jpg ImgType = iota
	Png
	Gif
)

type Request struct {
	OriginalName string
	Typ ImgType
	TargetWidth uint
	TargetHeight uint
	ImgData []byte
}

type Response struct {
	Width uint
	Height uint
	ImgData []byte
}