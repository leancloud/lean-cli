package upload

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func getSeekerSize(seeker io.Seeker) (int64, error) {
	size, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// Upload upload specific file to LeanCloud
func Upload(name string, mimeType string, reader io.ReadSeeker, opts *Options) (*File, error) {
	size, err := getSeekerSize(reader)
	if err != nil {
		return nil, err
	}

	tokens, err := getFileTokens(name, mimeType, size, opts)
	if err != nil {
		return nil, err
	}

	switch tokens.Provider {
	case "qiniu":
		return uploadToQiniu(name, tokens, reader, opts)
	case "s3":
		return uploadToS3(name, tokens, reader, opts)
	case "qcloud":
		return uploadToCOS(name, tokens, reader, opts)
	default:
		return nil, errors.New("Unknown file provider: " + tokens.Provider)
	}
}

func uploadToQiniu(name string, tokens *fileTokens, reader io.ReadSeeker, opts *Options) (*File, error) {
	out, in := io.Pipe()
	part := multipart.NewWriter(in)
	done := make(chan error)

	go func() {
		if err := part.WriteField("key", tokens.Key); err != nil {
			in.Close()
			done <- err
			return
		}
		if err := part.WriteField("token", tokens.Token); err != nil {
			in.Close()
			done <- err
			return
		}
		writer, err := part.CreateFormFile("file", name)
		if err != nil {
			in.Close()
			done <- err
			return
		}
		_, err = io.Copy(writer, reader)
		if err != nil {
			in.Close()
			done <- err
			return
		}
		if err := part.Close(); err != nil {
			in.Close()
			done <- err
			return
		}
		in.Close()
		done <- nil
	}()

	request, err := http.NewRequest("POST", "https://up.qbox.me/", out)
	request.Header.Set("Content-Type", part.FormDataContentType())
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	err = <-done
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(string(content))
	}

	return &File{
		ObjectID: tokens.ObjectID,
		URL:      tokens.URL,
	}, nil
}

func uploadToS3(name string, tokens *fileTokens, reader io.ReadSeeker, opts *Options) (*File, error) {
	request, err := http.NewRequest("PUT", tokens.UploadURL, reader)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", tokens.MimeType)
	request.Header.Set("Cache-Control", "public, max-age=31536000")
	request.ContentLength = tokens.Size

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	return &File{
		ObjectID: tokens.ObjectID,
		URL:      tokens.URL,
	}, nil
}

func uploadToCOS(name string, tokens *fileTokens, reader io.ReadSeeker, opts *Options) (*File, error) {
	out, in := io.Pipe()
	part := multipart.NewWriter(in)
	done := make(chan error)

	go func() {
		if err := part.WriteField("op", "upload"); err != nil {
			in.Close()
			done <- err
			return
		}
		writer, err := part.CreateFormFile("fileContent", name)
		if err != nil {
			in.Close()
			done <- err
			return
		}
		_, err = io.Copy(writer, reader)
		if err != nil {
			in.Close()
			done <- err
			return
		}
		if err := part.Close(); err != nil {
			in.Close()
			done <- err
			return
		}
		in.Close()
		done <- nil
	}()

	body, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", tokens.UploadURL+"?sign="+url.QueryEscape(tokens.Token), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", part.FormDataContentType())
	request.ContentLength = int64(len(body))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	err = <-done
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(string(content))
	}

	return &File{
		ObjectID: tokens.ObjectID,
		URL:      tokens.URL,
	}, nil
}

// UploadFileVerbose will open an file and upload it
func UploadFileVerbose(name string, mimeType string, path string, opts *Options) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return Upload(name, mimeType, f, opts)
}

// UploadFile will open an file and upload it. the file name and mime type is autodetected
func UploadFile(path string, opts *Options) (*File, error) {
	_, name := filepath.Split(path)
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	return UploadFileVerbose(name, mimeType, path, opts)
}
