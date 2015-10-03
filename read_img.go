package stegoimg

import (
	"encoding/binary"
	_ "golang.org/x/image/bmp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"fmt"
)

/*
A StegoImgReader wraps a go image, and can be used to
read the steganographically encoded content of the image.
*/
type StegoImgReader struct {
	data []byte
	pos  int
}

/*
Creates a StegoImgReader from the given image file.

This will assume the file being passed is a stego encoded image
and makes no attempts to vaidate the data it reads.

The img_file reader need not remain open after the call returns
*/
func NewStegoImgReader(img_file io.Reader) (img *StegoImgReader, e error) {

	// get the initial image decoded
	i, _, e := image.Decode(img_file)

	if e != nil {
		return nil, e
	}

	// make the reader
	img = new(StegoImgReader)

	// allocate 4B of data for the size
	img.data = make([]byte, 4)
	got_size := false
	pos := 0

	// read data
ReadLoop:
	// for each row in the image
	for y := 0; y < i.Bounds().Max.Y; y++ {

		// for each column in the row
		for x := 0; x < i.Bounds().Max.X; x++ {

			// get the color
			col := i.At(x, y)

			//get the values
			var vals [3]uint32
			vals[0], vals[1], vals[2], _ = col.RGBA()

			// for each color value in the pixel
			for v := 0; v < 3; v++ {

				// check if there's data to encode
				if pos < len(img.data) && pos < len(img.data) {

					// Decode the byte from the value
					img.data[pos] = byte(vals[v] & 0x00FF)
					pos++

				} else {
					break ReadLoop
				}

				// if this is the first readthrough, get the size
				if pos == 4 && !got_size {
					size := binary.BigEndian.Uint32(img.data)
					img.data = make([]byte, size)
					pos = 0
					got_size = true
					fmt.Printf("Reader: File size of %v bytes\n", size)
				}

				// else no data to encode, nothing to do

			} // end of each pixel

		} // end of each column

	} // end of each row

	return

}

func (img *StegoImgReader) Read(p []byte) (n int, e error) {

	to_read := len(p)
	if to_read > len(img.data)-img.pos {
		to_read = len(img.data) - img.pos
	}
	fmt.Printf("Pos %v\n", img.pos)
	fmt.Printf("Read requested for %v of %v, reading %v\n", len(p), len(img.data), to_read)

	for n = 0; n < to_read; n++ {
		p[n] = img.data[img.pos]
		img.pos++
	}

	if img.pos == len(img.data) {
		return n, io.EOF
	}

	return n, nil
}
