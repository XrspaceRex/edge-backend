package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/file"
	"gofr.dev/pkg/gofr/http/response"
)

func main() {

	app := gofr.New()

	app.AddHTTPService("edge", "http://192.168.0.175:8080")

	app.POST("/api/v1/invocations", func(ctx *gofr.Context) (interface{}, error) {

		var d Data

		// bind the multipart data into the variable d
		err := ctx.Bind(&d)
		if err != nil {
			return nil, err
		}

		if d.Music == nil {
			return nil, fmt.Errorf("file `music` invalid")
		}
		if d.AvatarOffset == nil {
			return nil, fmt.Errorf("file `avatar_offset` invalid")
		}

		music1, err := d.Music.Open()
		if err != nil {
			return nil, err
		}
		musicBytes, _ := io.ReadAll(music1)
		defer music1.Close()

		avatar1, err := d.AvatarOffset.Open()
		if err != nil {
			return nil, err
		}
		avatarBytes, _ := io.ReadAll(avatar1)
		avatar1.Close()

		buf := &bytes.Buffer{}
		gen := generateMultiPartBody(buf)
		writer, err := gen("music", musicBytes)
		if err != nil {
			return nil, err
		}
		if _, err := gen("avatar_offset", avatarBytes); err != nil {
			return nil, err
		}

		writer.Close()

		resp, err := http.Post("http://192.168.0.175:8080/invocations", writer.FormDataContentType(), buf)
		if err != nil {
			return nil, err
		}

		result, err := io.ReadAll(resp.Body)

		x := response.File{
			Content:     result,
			ContentType: "text/plain",
		}

		return x, err

	})

	app.Run()
}

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

func generateMultiPartBody(buf *bytes.Buffer) func(name string, content []byte) (*multipart.Writer, error) {

	writer := multipart.NewWriter(buf)

	return func(name string, content []byte) (*multipart.Writer, error) {
		musicPart, err := writer.CreateFormFile(name, name+".wav")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(musicPart, bytes.NewBuffer(content))
		if err != nil {
			return nil, fmt.Errorf("Failed to write file to form: %v", err)
		}

		return writer, nil
	}
}
