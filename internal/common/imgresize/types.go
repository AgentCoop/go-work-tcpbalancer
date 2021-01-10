package imgresize

type ImgType int

const (
	Jpeg ImgType = iota
	Png
	Gif
)

func (t ImgType) ToFileExt() string {
	return [...]string{".jpeg", ".png", ".giff"}[t]
}

type ImgInfo struct {
	OriginalName string
	Typ ImgType
	Width uint
	Height uint
	ImgData []byte
}

type Request struct {
	ImgInfo
	TargetWidth uint
	TargetHeight uint
	CreatedAt int64
	ImgIndex int
	DryRun bool
}

type Response struct {
	ImgInfo
	ResizedWidth uint
	ResizedHeight uint
	ProcessingTime int64
	CreatedAt int64
	ImgIndex int
}