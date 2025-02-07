/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"bytes"
	"os"
	"testing"

	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/GoogleContainerTools/kaniko/pkg/fakes"
	"github.com/google/go-cmp/cmp"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	image = "foo:latest"
)

func Test_Warmer_Warm_not_in_cache(t *testing.T) {
	tarBuf := new(bytes.Buffer)
	manifestBuf := new(bytes.Buffer)

	cw := &Warmer{
		Remote: func(_ string, _ config.RegistryOptions, _ string) (v1.Image, error) {
			return fakes.FakeImage{}, nil
		},
		Local: func(_ *config.CacheOptions, _ string) (v1.Image, error) {
			return nil, NotFoundErr{}
		},
		TarWriter:      tarBuf,
		ManifestWriter: manifestBuf,
	}

	opts := &config.WarmerOptions{}

	_, err := cw.Warm(Image{
		Reference: image,
		Platform:  "",
	}, opts)
	if err != nil {
		t.Errorf("expected error to be nil but was %v", err)
		t.FailNow()
	}

	if len(tarBuf.Bytes()) == 0 {
		t.Error("expected image to be written but buffer was empty")
	}
}

func Test_Warmer_Warm_in_cache_not_expired(t *testing.T) {
	tarBuf := new(bytes.Buffer)
	manifestBuf := new(bytes.Buffer)

	cw := &Warmer{
		Remote: func(_ string, _ config.RegistryOptions, _ string) (v1.Image, error) {
			return fakes.FakeImage{}, nil
		},
		Local: func(_ *config.CacheOptions, _ string) (v1.Image, error) {
			return fakes.FakeImage{}, nil
		},
		TarWriter:      tarBuf,
		ManifestWriter: manifestBuf,
	}

	opts := &config.WarmerOptions{}

	_, err := cw.Warm(Image{
		Reference: image,
		Platform:  "",
	}, opts)
	if !IsAlreadyCached(err) {
		t.Errorf("expected error to be already cached err but was %v", err)
		t.FailNow()
	}

	if len(tarBuf.Bytes()) != 0 {
		t.Errorf("expected nothing to be written")
	}
}

func Test_Warmer_Warm_in_cache_expired(t *testing.T) {
	tarBuf := new(bytes.Buffer)
	manifestBuf := new(bytes.Buffer)

	cw := &Warmer{
		Remote: func(_ string, _ config.RegistryOptions, _ string) (v1.Image, error) {
			return fakes.FakeImage{}, nil
		},
		Local: func(_ *config.CacheOptions, _ string) (v1.Image, error) {
			return fakes.FakeImage{}, ExpiredErr{}
		},
		TarWriter:      tarBuf,
		ManifestWriter: manifestBuf,
	}

	opts := &config.WarmerOptions{}

	_, err := cw.Warm(Image{
		Reference: image,
		Platform:  "",
	}, opts)
	if !IsAlreadyCached(err) {
		t.Errorf("expected error to be already cached err but was %v", err)
		t.FailNow()
	}

	if len(tarBuf.Bytes()) != 0 {
		t.Errorf("expected nothing to be written")
	}
}

func TestParseDockerfile_SingleStageDockerfile(t *testing.T) {
	dockerfile := `FROM alpine:latest
LABEL maintainer="alexezio"
`
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := &config.WarmerOptions{DockerfilePath: tmpfile.Name()}
	images, err := ParseDockerfile(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	expected := Image{
		Reference: "alpine:latest",
		Platform:  "",
	}
	if !cmp.Equal(images[0], expected) {
		t.Fatalf("expected '%v', got '%v'", expected, images[0])
	}
}

func TestParseDockerfile_MultiStageDockerfile(t *testing.T) {
	dockerfile := `FROM golang:1.20 as BUILDER
LABEL maintainer="alexezio"

FROM alpine:latest as RUNNER
LABEL maintainer="alexezio"
`
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := &config.WarmerOptions{DockerfilePath: tmpfile.Name()}
	images, err := ParseDockerfile(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 image, got %d", len(images))
	}
	expected := []Image{
		{
			Reference: "golang:1.20",
			Platform:  "",
		},
		{
			Reference: "alpine:latest",
			Platform:  "",
		},
	}
	if !cmp.Equal(expected[0], images[0]) {
		t.Fatalf("expected '%v', got '%v'", expected[0], images[0])
	}

	if !cmp.Equal(expected[1], images[1]) {
		t.Fatalf("expected '%v', got '%v'", expected[1], images[1])
	}
}

func TestParseDockerfile_ArgsDockerfile(t *testing.T) {
	dockerfile := `ARG version=latest
FROM golang:${version}
`
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := &config.WarmerOptions{DockerfilePath: tmpfile.Name(), BuildArgs: []string{"version=1.20"}}
	images, err := ParseDockerfile(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	expected := Image{
		Reference: "golang:1.20",
		Platform:  "",
	}
	if !cmp.Equal(images[0], expected) {
		t.Fatalf("expected '%v', got '%v'", expected, images[0])
	}
}

func TestParseDockerfile_MissingsDockerfile(t *testing.T) {
	opts := &config.WarmerOptions{DockerfilePath: "dummy-nowhere"}
	baseNames, err := ParseDockerfile(opts)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if len(baseNames) != 0 {
		t.Fatalf("expected no base names, got %d", len(baseNames))
	}
}

func TestParseDockerfile_InvalidsDockerfile(t *testing.T) {
	dockerfile := "This is a invalid dockerfile"
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	opts := &config.WarmerOptions{DockerfilePath: tmpfile.Name()}
	baseNames, err := ParseDockerfile(opts)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if len(baseNames) != 0 {
		t.Fatalf("expected no base names, got %d", len(baseNames))
	}
}
