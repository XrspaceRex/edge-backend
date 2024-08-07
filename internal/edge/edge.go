package edge

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/go-audio/wav"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/file"
)

// Data is the struct that we are trying to bind files to
type Data struct {
	// Name represents the non-file field in the struct
	Name string `form:"name"`

	// The Compressed field is of type zip,
	// the tag `upload` signifies the key for the form where the file is uploaded
	// if the tag is not present, the field name would be taken as a key.
	Compressed file.Zip `file:"upload"`

	// The FileHeader determines the generic file format that we can get
	// from the multipart form that gets parsed by the incoming HTTP request
	Music        *multipart.FileHeader `file:"music"`
	AvatarOffset *multipart.FileHeader `file:"avatar_offset"`
}

func Edge(ctx *gofr.Context) (interface{}, error) {
	var d Data

	// bind the multipart data into the variable d
	err := ctx.Bind(&d)
	if err != nil {
		return nil, err
	}

	music1, err := d.Music.Open()
	if err != nil {
		return nil, err
	}

	defer music1.Close()

	avatar1, err := d.AvatarOffset.Open()
	if err != nil {
		panic(err)
	}
	avatarBytes, _ := io.ReadAll(avatar1)
	avatar1.Close()

	wd := wav.NewDecoder(music1)

	for {
		chunk, err := wd.NextChunk()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		buf := &bytes.Buffer{}

		cb := []byte{}
		chunk.Read(cb)
		gen := generateMultiPartBody(buf)
		writer := gen("music", cb)
		gen("avatar_offset", avatarBytes)

		writer.Close()

		// resp, err := http.Post("http://192.168.0.175:8080/invocations", writer.FormDataContentType(), buf)
		// if err != nil {
		// 	panic(err)
		// }

		// b, _ := io.ReadAll(resp.Body)
	}

	// return the number of compressed files received
	return fmt.Sprintf("music length: %d, avatar length: %d", 1, len(avatarBytes)), nil
}

func generateMultiPartBody(buf *bytes.Buffer) func(name string, content []byte) *multipart.Writer {

	writer := multipart.NewWriter(buf)

	return func(name string, content []byte) *multipart.Writer {
		musicPart, err := writer.CreateFormFile(name, name+".wav")
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(musicPart, bytes.NewBuffer(content))
		if err != nil {
			fmt.Errorf("Failed to write file to form: %v", err)
		}

		return writer
	}
}
