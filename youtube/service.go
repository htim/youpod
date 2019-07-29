package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/htim/youpod"
	"github.com/rs/xid"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//wrapper around youtube-dl cmd
type Service struct {
}

func NewService() (*Service, error) {
	cmd := exec.Command("youtube-dl", "--version")

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "cannot get youtube-dl version")
	}

	log.Debugf("found youtube-dl:%s", buf.String())

	return &Service{}, nil
}

func (d *Service) Download(owner youpod.User, link string) (youpod.File, error) {

	id := xid.New().String()

	log.Debugf("downloading %s", link)

	var stdout, stderr bytes.Buffer

	output := fmt.Sprintf("%s.%%(ext)s", id)

	cmd := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "-o", output, "--write-info-json", link)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return youpod.File{}, errors.Wrapf(err, "cannot download video: %s", stderr.String())
	}

	if stderr.Len() > 0 {
		return youpod.File{}, errors.Errorf("cannot download video: %s", stderr.String())
	}

	log.Debugf("downloading %s completed: %s", link, stdout.String())

	infoJson, err := ioutil.ReadFile(fmt.Sprintf("%s.info.json", id))
	if err != nil {
		return youpod.File{}, errors.Wrap(err, "cannot read info.json file")
	}

	var info info
	if err = json.Unmarshal(infoJson, &info); err != nil {
		return youpod.File{}, errors.Wrap(err, "cannot unmarshal info.json file")
	}

	f, err := os.Open(fmt.Sprintf("%s.mp3", id))
	if err != nil {
		return youpod.File{}, errors.Wrap(err, "cannot open downloaded file")
	}

	return youpod.File{
		FileMetadata: youpod.FileMetadata{
			ID:   id,
			Name: info.Fulltitle,
		},
		Content: f,
	}, nil
}

func (d *Service) Cleanup(f youpod.File) {
	if err := f.Content.Close(); err != nil {
		log.WithError(err).Debug("file is already closed")
	}
	mp3 := fmt.Sprintf("./%s.%s", f.ID, "mp3")
	infoJson := fmt.Sprintf("./%s.%s", f.ID, "info.json")
	if err := os.Remove(mp3); err != nil {
		log.WithError(err).Debugf("cannot remove file: %s", mp3)
	}
	if err := os.Remove(infoJson); err != nil {
		log.WithError(err).Debugf("cannot remove file: %s", infoJson)
	}
}

type info struct {
	Fulltitle   string `json:"fulltitle"`
	Description string `json:"description"`
	Uploader    string `json:"uploader"`
}
