package python

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2/registry/remote"
)

// Upload will upload the package to pypi with contents cr and metadata from entry.
func Upload(ctx context.Context, client remote.Client, getContent func() (io.ReadCloser, error), pypi string, entry DistributionEntry) error {
	// This function takes a function getContent instead of a io.ReadCLoser directly because we need to support retry

	/*
		curl --request POST \
			--form 'content=@path/to/my.pypi.package-0.0.1.tar.gz' \
			--form 'name=my.pypi.package'
			--form 'version=1.3.7'
			--user <username>:<personal_access_token> \
			"https://gitlab.example.com/api/v4/projects/1/packages/pypi?requires_python=3.7"
	*/

	dist, err := NewDistribution(entry.Filename)
	if err != nil {
		return err
	}

	u, err := url.Parse(pypi)
	if err != nil {
		return fmt.Errorf("parsing PyPI URL: %w", err)
	}

	fields := map[string]string{}

	fields[":action"] = "file_upload" // required by artifactory
	fields["protocol_version"] = "1"
	fields["name"] = dist.Project()
	fields["version"] = dist.Version()

	if d := entry.Digest; d.Algorithm() == digest.SHA256 {
		fields["sha256_digest"] = d.Encoded()
	}

	if entry.RequiresPython != "" {
		// It seems that the comma delimited list of python version is not parsed
		fields["requires_python"] = entry.RequiresPython
	}

	// TODO there are many other fields that can be set.
	// It seems like they are extracted from the encoded metadata

	getBody := func() (io.ReadCloser, string, error) {
		r, w := io.Pipe()
		mp := multipart.NewWriter(w)

		go func() {
			for k, v := range fields {
				if err := mp.WriteField(k, v); err != nil {
					w.CloseWithError(err)
					return
				}
			}

			cw, err := mp.CreateFormFile("content", entry.Filename)
			if err != nil {
				w.CloseWithError(err)
				return
			}

			cr, err := getContent()
			if err != nil {
				w.CloseWithError(err)
				return
			}
			defer cr.Close()

			_, err = io.Copy(cw, cr)
			if err != nil {
				w.CloseWithError(err)
				return
			}

			err = mp.Close()
			if err != nil {
				w.CloseWithError(err)
				return
			}

			w.Close()
		}()

		return r, mp.FormDataContentType(), nil
	}

	rc, ct, err := getBody()
	if err != nil {
		return err
	}
	defer rc.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), rc)
	if err != nil {
		return fmt.Errorf("creating PyPI request: %w", err)
	}

	// support retry
	req.GetBody = func() (io.ReadCloser, error) {
		rc, _, err := getBody()
		return rc, err
	}

	req.Header.Add("Content-Type", ct)

	// make the actual request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching python package: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code (%d) during POST of package %s", resp.StatusCode, entry.Filename)
	}

	return nil
}
