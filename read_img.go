package stegoimg

import (
	"errors"
//	"image"
	"io"
)

/*
A StegoImgReader wraps a go image, and can be used to
read the steganographically encoded content of the image.
*/
type StegoImgReader struct {
	a int
}

func NewStegoImgReader(img_file io.Reader) (img *StegoImgReader, e error) {
	e = errors.New("NewStegoImgReader not yet implemented")
	img = nil
	return
}
