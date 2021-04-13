// Copyright 2020 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/tetratelabs/getenvoy/api"
)

const (
	archiveFormat = ".tar.gz"
)

// ServerOpts represents configuration options for NewServer.
type ServerOpts struct {
	Manifest       *api.Manifest
	GetArtifactDir func(uri string) (string, error)
	OnError        func(error)
}

// Server represents an HTTP server serving GetEnvoy manifest.
type Server interface {
	GetManifestURL() string
	Close()
}

// NewServer returns a new HTTP server serving GetEnvoy manifest.
func NewServer(opts *ServerOpts) Server {
	s := &server{opts: opts}
	s.http = httptest.NewServer(s)

	manifest, _ := proto.Clone(opts.Manifest).(*api.Manifest)
	s.rewriteManifestURLs(manifest)
	s.manifest = manifest

	return s
}

// server represents an HTTP server serving GetEnvoy manifest.
type server struct {
	opts     *ServerOpts
	manifest *api.Manifest
	http     *httptest.Server
}

func (s *server) GetManifestURL() string {
	return s.http.URL + "/getenvoy/manifest.json"
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		switch {
		case r.RequestURI == "/getenvoy/manifest.json":
			payload, err := s.GetManifest()
			if err != nil {
				return err
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(payload)
			if err != nil {
				return err
			}
		case strings.HasPrefix(r.RequestURI, "/getenvoy/builds/"):
			uri := r.RequestURI[len("/getenvoy/builds/"):]
			payload, err := s.GetArtifact(uri)
			if err != nil {
				return err
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(payload)
			if err != nil {
				return err
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		return nil
	}
	if err := handler(w, r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		s.opts.OnError(err)
	}
}

func (s *server) GetManifest() ([]byte, error) {
	data, err := protojson.Marshal(s.manifest)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *server) GetArtifact(uri string) ([]byte, error) {
	if !strings.HasSuffix(uri, archiveFormat) {
		return nil, fmt.Errorf("unexpected uri %q: expected archive suffix %q", uri, archiveFormat)
	}
	uri = uri[:len(uri)-len(archiveFormat)]
	dir, err := s.opts.GetArtifactDir(uri)
	if err != nil {
		return nil, err
	}
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck
	tarFile := filepath.Join(tmpDir, "archive"+archiveFormat)
	err = archiver.Archive([]string{dir}, tarFile)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Clean(tarFile))
}

func (s *server) rewriteManifestURLs(manifest *api.Manifest) {
	for _, flavor := range manifest.Flavors {
		for _, version := range flavor.Versions {
			for _, build := range version.Builds {
				build.DownloadLocationUrl = fmt.Sprintf("%s/getenvoy/builds/%s%s", s.http.URL, build.DownloadLocationUrl, archiveFormat)
			}
		}
	}
}

func (s *server) Close() {
	s.http.Close()
}
