package youtube

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/htim/youpod/core"
	"github.com/rs/xid"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//wrapper around youtube-dl cmd
type Service struct {
	outputDir string
}

func NewService(outputDir string) (*Service, error) {
	cmd := exec.Command("youtube-dl", "--version")

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "cannot get youtube-dl version")
	}

	log.Debugf("found youtube-dl:%s", buf.String())

	return &Service{
		outputDir: outputDir,
	}, nil
}

func (d *Service) Download(owner core.User, link string) (core.File, error) {

	id := xid.New().String()

	log.Debugf("downloading %s", link)

	var stdout, stderr bytes.Buffer

	output := fmt.Sprintf("%s/%s.%%(ext)s", d.outputDir, id)

	cmd := exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "-o", output, "--write-info-json", link)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return core.File{}, errors.Wrapf(err, "cannot download video: %s", stderr.String())
	}

	if stderr.Len() > 0 {
		errString := stderr.String()
		if strings.Contains(errString, "ERROR") {
			return core.File{}, errors.Errorf("cannot download video: %s", errString)
		}
	}

	log.Debugf("downloading %s completed: %s", link, stdout.String())

	infoJson, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.info.json", d.outputDir, id))
	if err != nil {
		return core.File{}, errors.Wrap(err, "cannot read info.json file")
	}

	var info info
	if err = json.Unmarshal(infoJson, &info); err != nil {
		return core.File{}, errors.Wrap(err, "cannot unmarshal info.json file")
	}

	f, err := os.Open(fmt.Sprintf("%s/%s.mp3", d.outputDir, id))
	if err != nil {
		return core.File{}, errors.Wrap(err, "cannot open downloaded file")
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return core.File{}, errors.Wrap(err, "cannot get file info")
	}

	picture, err := thumbnailBase64(info.Thumbnail)
	if err != nil {
		log.WithError(err).WithField("thumbnail", info.Thumbnail).Error("cannot prepare file thumbnail")
	}

	return core.File{
		Metadata: core.Metadata{
			TmpFileID: id,
			Name:      info.Fulltitle,
			Author:    info.Uploader,
			Size:      fileInfo.Size(),
			Picture:   picture,
			CreatedAt: time.Now(),
		},
		Content: f,
	}, nil
}

func (d *Service) Cleanup(f core.File) {
	if err := f.Content.Close(); err != nil {
		log.WithError(err).Debug("file is already closed")
	}
	mp3 := fmt.Sprintf("%s/%s.%s", d.outputDir, f.TmpFileID, "mp3")
	infoJson := fmt.Sprintf("%s/%s.%s", d.outputDir, f.TmpFileID, "info.json")
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
	Thumbnail   string `json:"thumbnail"`
}

func thumbnailBase64(url string) (string, error) {
	thumbnail, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "cannot download")
	}
	defer func() {
		if err := thumbnail.Body.Close(); err != nil {
			log.WithError(err).WithField("thumbnail", url).Error("cannot close thumbnail response body")
		}
	}()

	img, err := jpeg.Decode(thumbnail.Body)
	if err != nil {
		return "", errors.Wrap(err, "cannot decode image response body")
	}

	min := img.Bounds().Dx()
	if img.Bounds().Dy() < min {
		min = img.Bounds().Dy()
	}

	resized := imaging.CropCenter(img, min, min)
	buf := &bytes.Buffer{}
	if err = jpeg.Encode(buf, resized, nil); err != nil {
		return "", errors.Wrap(err, "cannot encode resized image")
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
