package stegoimg

import (
	"encoding/binary"
	"errors"
	"golang.org/x/image/bmp"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

/*
A StegoImgWriter encompasses an input image, and a resulting
output image. As data is written to it, the data is encoded
into the data from the input and written to the output.
*/
type StegoImgWriter struct {
	is_open  bool
	orig_img image.Image
	new_img  *image.NRGBA64
	output   io.Writer
	format   string
	data     []byte
}

/*
Error returned when a write fills the remaining space in an image.

The return value of Write should be consulted prior to consideration
of this error, as it will be retured even if all bytes were written
successfully if it fills the final available byte."
*/
var ImageFullError = errors.New("Image full. No more data can be written.")

/*
Error returned if the image is closed for writing
*/
var ImageClosedError = errors.New("Image is closed for writing or wasn't opened properly.")

/*
Create a new StegoImgWriter.
orig_img is a reader which should be the file of the image to encode the data into.
new_img is the file that the encoded image will be written to.
format can be one of "png", "jpeg", or "gif"

The orig_img reader need not remain open after the call returns. The writer must remain open until after the call to Close returns.
*/
func NewStegoImgWriter(orig_img io.Reader, new_img io.Writer) (img *StegoImgWriter, e error) {

	// create the image writer
	tmp_img := new(StegoImgWriter)

	// attempt to decode the original image
	tmp_img.orig_img, tmp_img.format, e = image.Decode(orig_img)

	// BUG(Andrew): only PNG format currently works
	tmp_img.format = "png"

	if e != nil {
		return
	}

	// create a new image
	tmp_img.new_img = image.NewNRGBA64(tmp_img.orig_img.Bounds())

	// make the data array to store all potential data
	size := (tmp_img.orig_img.Bounds().Max.X - 1) * (tmp_img.orig_img.Bounds().Max.Y - 1) * 3
	tmp_img.data = make([]byte, 0, size)

	// block out space for the size argument
	tmp_img.data = append(tmp_img.data, 0, 0, 0, 0)

	// save output
	tmp_img.output = new_img

	// mark image as open
	tmp_img.is_open = true

	// all was successful, return the new pointer
	img = tmp_img
	return
}

/*
Write data into an image. This will encode the data into the least significant
bits of the color values of each pixel.

After each pixel has had data encoded into it, subsequent calls to Write will
return 0 bytes written, as well as a ImageFullError. The n value should be
considered before the EOF error

Note that the image is not actually created until the call to Close, even if
the medium is full.
*/
func (img *StegoImgWriter) Write(p []byte) (n int, err error) {
	if !img.is_open {
		return n, ImageClosedError
	}
	for i := 0; i < len(p); i++ {
		if len(img.data) < cap(img.data) {
			img.data = append(img.data, p[i])
			n++
		} else {
			return n, ImageFullError
		}
	}
	return n, nil
}

/*
Close finishes encoding data into the image. This will cause the image data to
be written to the new_img file provided upon creation.
*/
func (img *StegoImgWriter) Close() error {

	// calculate the total size of the file
	var size uint32 = uint32(len(img.data) - 4)

	// set the size into the data
	binary.BigEndian.PutUint32(img.data[:4], size)

	// the data byte to write
	pos := 0

	// for each row in the image
	for y := 0; y < img.orig_img.Bounds().Max.Y; y++ {

		// for each column in the row
		for x := 0; x < img.orig_img.Bounds().Max.X; x++ {

			// get the color
			col := img.orig_img.At(x, y)

			//get the values
			var vals [3]uint32
			vals[0], vals[1], vals[2], _ = col.RGBA()

			// for each color value in the pixel
			for v := 0; v < 3; v++ {

				// check if there's data to encode
				if pos < len(img.data) && pos < len(img.data) {

					// Encode the byte into the value
					vals[v] = vals[v] & 0xFF00
					vals[v] = vals[v] | uint32(img.data[pos])
					pos++

				}

				// else no data to encode, nothing to do

			} // end of each pixel

			// write the pixel to the new image
			img.new_img.SetNRGBA64(x, y, color.NRGBA64{uint16(vals[0]), uint16(vals[1]), uint16(vals[2]), 0xFFFF})

		} // end of each column

	} // end of each row

	img.is_open = false

	// determine the format to encode into
	switch img.format {
	case "png":
		png.Encode(img.output, img.new_img)
	case "jpeg":
		jpeg.Encode(img.output, img.new_img, &jpeg.Options{100})
	case "gif":
		gif.Encode(img.output, img.new_img, &gif.Options{256, nil, nil})
	case "bmp":
		bmp.Encode(img.output, img.new_img)
	default:
		return errors.New("Somehow got an unaccepted format of " + img.format)
	}

	return nil

}

/*
Returns the amount of space left for data in the image
*/
func (img *StegoImgWriter) SpaceLeft() int {
	return cap(img.data) - len(img.data)
}
