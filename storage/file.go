package storage

import (
	"errors"
	"fmt"
	"github.com/crazybolillo/reis/codec/ulaw"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"os"
	"path/filepath"
)

type FileBackend struct {
	basePath string
}

type FileRecord struct {
	fd      *os.File
	encoder *wav.Encoder
}

func (f *FileRecord) Write(p []byte) (n int, err error) {
	expanded := ulaw.Expand(p)
	payload := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: f.encoder.NumChans,
			SampleRate:  f.encoder.SampleRate,
		},
		Data:           expanded,
		SourceBitDepth: f.encoder.BitDepth,
	}

	return len(p), f.encoder.Write(&payload)
}

func (f *FileRecord) Close() error {
	err := f.encoder.Close()
	err = errors.Join(f.fd.Close(), err)

	return err
}

func NewFileBackend(basePath string) (*FileBackend, error) {
	stat, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}

	return &FileBackend{basePath: basePath}, nil
}

func (f *FileBackend) New(name string) (Record, error) {
	path, err := filepath.Abs(filepath.Join(f.basePath, name+".wav"))
	if err != nil {
		return nil, err
	}

	fd, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	encoder := wav.NewEncoder(fd, 8000, 16, 1, 1)

	return &FileRecord{
		fd:      fd,
		encoder: encoder,
	}, nil
}
