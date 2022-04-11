package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
)

const (
	firstQuarter                = 192
	secondQuarter               = 48
	thirdQuarter                = 12
	fourthQuarter               = 3
	dataSizeHeaderReservedBytes = 20
)

func Encode(pic io.Reader, data io.Reader, result io.Writer) error {
	RGBAImage, format, err := getImageAsRGBA(pic)
	if err != nil {
		return err
	}

	isPng := false
	if format == "png" {
		isPng = true
	}
	dataBytes := make(chan byte, 128)
	errChan := make(chan error)

	go readData(data, dataBytes, errChan)

	dx := RGBAImage.Bounds().Dx()
	dy := RGBAImage.Bounds().Dy()

	hasMoreBytes := true

	var count int
	var dataCount uint32

	for x := 0; x < dx && hasMoreBytes; x++ {
		for y := 0; y < dy && hasMoreBytes; y++ {
			if count >= dataSizeHeaderReservedBytes {
				c := RGBAImage.RGBAAt(x, y)
				flag := false
				if isPng && c.A == byte(0x00) {
					flag = true
				}
				hasMoreBytes, err = setColorSegment(flag, &c.R, dataBytes, errChan)
				if err != nil {
					return err
				}
				if hasMoreBytes {
					dataCount++
				}
				hasMoreBytes, err = setColorSegment(flag, &c.G, dataBytes, errChan)
				if err != nil {
					return err
				}
				if hasMoreBytes {
					dataCount++
				}
				hasMoreBytes, err = setColorSegment(flag, &c.B, dataBytes, errChan)
				if err != nil {
					return err
				}
				if hasMoreBytes {
					dataCount++
				}
				RGBAImage.SetRGBA(x, y, c)
			} else {
				count += 4
			}
		}
	}

	if dataCount == 0 {
		return errors.New("the image isn't supported steganographic!")
	}
	select {
	case _, ok := <-dataBytes:
		if ok {
			return errors.New("data exceeds image capacity!")
		}
	default:
	}

	setDataSizeHeader(RGBAImage, quartersOfBytesOf(dataCount))

	switch format {
	case "png", "jpeg":
		return png.Encode(result, RGBAImage)
	default:
		return errors.New("unsupported image format!")
	}
}

func Decode(pic io.Reader) ([]byte, error) {
	resultBytes := make([]byte, 0, 40960)
	RGBAImage, _, err := getImageAsRGBA(pic)
	if err != nil {
		return []byte{}, err
	}

	dx := RGBAImage.Bounds().Dx()
	dy := RGBAImage.Bounds().Dy()

	dataBytes := make([]byte, 0, 40960)
	dataCount := extractDataCount(RGBAImage)

	var count int

	for x := 0; x < dx && dataCount > 0; x++ {
		for y := 0; y < dy && dataCount > 0; y++ {
			if count >= dataSizeHeaderReservedBytes {
				c := RGBAImage.RGBAAt(x, y)
				dataBytes = append(dataBytes, getLastTwoBits(c.R), getLastTwoBits(c.G), getLastTwoBits(c.B))
				dataCount -= 3
			} else {
				count += 4
			}
		}
	}
	if dataCount > 0 {
		return []byte{}, fmt.Errorf("the image isn't steganographic!")
	}
	if dataCount < 0 {
		dataBytes = dataBytes[:len(dataBytes)+dataCount]
	}

	dataBytes = align(dataBytes)

	for i := 0; i < len(dataBytes); i += 4 {
		resultBytes = append(resultBytes, constructByteOfQuartersAsSlice(dataBytes[i:i+4]))
	}

	return resultBytes, nil
}

func quartersOfBytesOf(counter uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, counter)
	quarters := make([]byte, 16)
	for i := 0; i < 16; i += 4 {
		quarters[i] = quartersOfByte(bs[i/4])[0]
		quarters[i+1] = quartersOfByte(bs[i/4])[1]
		quarters[i+2] = quartersOfByte(bs[i/4])[2]
		quarters[i+3] = quartersOfByte(bs[i/4])[3]
	}

	return quarters
}

func setDataSizeHeader(RGBAImage *image.RGBA, dataCountBytes []byte) {
	dx := RGBAImage.Bounds().Dx()
	dy := RGBAImage.Bounds().Dy()

	count := 0

	for x := 0; x < dx && count < (dataSizeHeaderReservedBytes/4)*3; x++ {
		for y := 0; y < dy && count < (dataSizeHeaderReservedBytes/4)*3; y++ {
			c := RGBAImage.RGBAAt(x, y)
			c.R = setLastTwoBits(c.R, dataCountBytes[count])
			c.G = setLastTwoBits(c.G, dataCountBytes[count+1])
			c.B = setLastTwoBits(c.B, dataCountBytes[count+2])
			RGBAImage.SetRGBA(x, y, c)

			count += 3

		}
	}
}

func setColorSegment(flag bool, colorSegment *byte, data <-chan byte, errChan <-chan error) (hasMoreBytes bool, err error) {
	select {
	case byte, ok := <-data:
		if !ok || flag {
			return false, nil
		}
		*colorSegment = setLastTwoBits(*colorSegment, byte)
		return true, nil

	case err := <-errChan:
		return false, err

	}
}

func readData(reader io.Reader, bytes chan<- byte, errChan chan<- error) {
	b := make([]byte, 1)
	for {
		if _, err := reader.Read(b); err != nil {
			if err == io.EOF {
				break
			}
			errChan <- err
			return
		}
		for _, b := range quartersOfByte(b[0]) {
			bytes <- b
		}

	}
	close(bytes)
}

func getImageAsRGBA(reader io.Reader) (*image.RGBA, string, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, format, err
	}

	RGBAImage := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(RGBAImage, RGBAImage.Bounds(), img, img.Bounds().Min, draw.Src)

	return RGBAImage, format, nil
}

func quartersOfByte(b byte) [4]byte {
	return [4]byte{b & firstQuarter >> 6, b & secondQuarter >> 4, b & thirdQuarter >> 2, b & fourthQuarter}
}

func clearLastTwoBits(b byte) byte {
	return b & byte(252)
}

func setLastTwoBits(b byte, value byte) byte {
	return clearLastTwoBits(b) | value
}

func getLastTwoBits(b byte) byte {
	return b & fourthQuarter
}

func constructByteOfQuarters(first, second, third, fourth byte) byte {
	return (((first << 6) | (second << 4)) | third<<2) | fourth
}

func constructByteOfQuartersAsSlice(b []byte) byte {
	return constructByteOfQuarters(b[0], b[1], b[2], b[3])
}

func align(dataBytes []byte) []byte {
	switch len(dataBytes) % 4 {
	case 1:
		dataBytes = append(dataBytes, byte(0), byte(0), byte(0))
	case 2:
		dataBytes = append(dataBytes, byte(0), byte(0))
	case 3:
		dataBytes = append(dataBytes, byte(0))
	}
	return dataBytes
}

func extractDataCount(RGBAImage *image.RGBA) int {
	dataCountBytes := make([]byte, 0, 16)

	dx := RGBAImage.Bounds().Dx()
	dy := RGBAImage.Bounds().Dy()

	count := 0

	for x := 0; x < dx && count < dataSizeHeaderReservedBytes; x++ {
		for y := 0; y < dy && count < dataSizeHeaderReservedBytes; y++ {
			c := RGBAImage.RGBAAt(x, y)
			dataCountBytes = append(dataCountBytes, getLastTwoBits(c.R), getLastTwoBits(c.G), getLastTwoBits(c.B))
			count += 4
		}
	}

	dataCountBytes = append(dataCountBytes, byte(0))

	var bs = []byte{constructByteOfQuartersAsSlice(dataCountBytes[:4]),
		constructByteOfQuartersAsSlice(dataCountBytes[4:8]),
		constructByteOfQuartersAsSlice(dataCountBytes[8:12]),
		constructByteOfQuartersAsSlice(dataCountBytes[12:])}

	return int(binary.LittleEndian.Uint32(bs))
}
