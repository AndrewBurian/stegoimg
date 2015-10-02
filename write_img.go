package stegoimg

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

/*
StegoImgWriter doc
*/
type StegoImgWriter struct {
	is_closed  bool
	orig_img   image.Image
	new_img    *image.NRGBA64
	output     io.Writer
	format     string
	data []byte
	space_left int
}

/*
Error returned when a write fills the remaining space in an image.

The return value of Write should be consulted prior to consideration
of this error, as it will be retured even if all bytes were written
successfully if it fills the final available byte."
*/
var ImageFullError = errors.New("Image full. No more data can be written.")

/*
Create a new StegoImgWriter.
orig_img is a reader which should be the file of the image to encode the data into.
new_img is the file that the encoded image will be written to.
format can be one of "png", "jpeg", or "gif"
*/
func NewStegoImgWriter(orig_img io.Reader, new_img io.Writer) (img *StegoImgWriter, e error) {

	// create the image writer
	tmp_img := new(StegoImgWriter)

	// attempt to decode the original image
	tmp_img.orig_img, tmp_img.format, e = image.Decode(orig_img)

	if e != nil {
		return
	}

	// create a new image
	tmp_img.new_img = image.NewNRGBA64(tmp_img.orig_img.Bounds())

	// keep track of the space left
	tmp_img.space_left = (((tmp_img.orig_img.Bounds().Max.X - 1) * (tmp_img.orig_img.Bounds().Max.Y - 1)) * 3) - 3

	// make the data array to store all potential data
	tmp_img.data = make([]byte, 0, tmp_img.space_left)

	// all was successful, return the new pointer
	img = tmp_img
	return
}

/*
Write data into an image. This will encode the data into the least significant
bits of the color values of each pixel.

After each pixel has had data encoded into it, subsequent calls to Write will
return 0 bytes written, as well as a ImageFullError.

Note that the image is not actually created until the call to Close, even if
the medium is full.
*/
func (img *StegoImgWriter) Write(p []byte) (n int, err error) {

	// keep track of written bytes
	n = 0
	to_write := len(p)

	if to_write > img.space_left {
		to_write = img.space_left
	}

	// watch the maximums
	bounds := img.orig_img.Bounds()
	bound_x := bounds.Max.X
	bound_y := bounds.Max.Y

	var orig_color color.Color
	var vals [3]uint32
	var tmp uint32

	// for each row in the picture
	for ; img.y < bound_y && to_write > 0; img.y++ {

		// for each column in the row
		for ; img.x < bound_x && to_write > 0; img.x++ {

			// get the pixel at the row/column
			orig_color = img.orig_img.At(img.x, img.y)

			// get the colors from the pixel
			vals[0], vals[1], vals[2], _ = orig_color.RGBA()

			// for each of the 3 color values per pixel
			for ; img.col < 3 && to_write > 0; img.col++ {

				// encode a single byte into the value
				// update loop trackers
				to_write--
				img.space_left--
				n++
			}

		}
	}

	// if the image is full, return the error
	if img.y == bound_y && img.x == bound_x {
		return n, ImageFullError
	}

	return n, nil
}

/*
Close finishes encoding data into the image. This will cause the image data to
be written to the new_img file provided upon creation.
*/
func (img *StegoImgWriter) Close() error {

	img.is_closed = true

	// determine the format to encode into
	switch img.format {
	case "png":
		png.Encode(img.output, img.new_img)
	case "jpeg":
		jpeg.Encode(img.output, img.new_img, &jpeg.Options{100})
	case "gif":
		gif.Encode(img.output, img.new_img, &gif.Options{256, nil, nil})
	default:
		return errors.New("Somehow got an unaccepted format of " + img.format)
	}

	return nil

}
