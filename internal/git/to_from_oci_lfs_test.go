package git

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/oci"
)

// lfsTest is an extension of test to aid in validating git lfs files.
type lfsTest struct {
	t                test
	expectedLFSFiles int
}

var lfsTests = []lfsTest{
	{
		t: test{
			name: "One LFS File",
			args: args{
				argRevList:          []string{"main"},
				expectedTagList:     []string{},
				expectedHeadList:    []string{"main"},
				expectedRebuildRefs: []string{"refs/heads/main"},
				tag:                 "sync",
			},
			wantErr: false,
		},
		expectedLFSFiles: 1,
	},
	{
		t: test{
			name: "Two LFS Files",
			args: args{
				argRevList:          []string{"Feature1"},
				expectedTagList:     []string{},
				expectedHeadList:    []string{"main", "Feature1"},
				expectedRebuildRefs: []string{"refs/heads/main", "refs/heads/Feature1"},
				tag:                 "sync",
			},
			wantErr: false,
		},
		expectedLFSFiles: 2,
	},
}

// Test_ToFromOCILFS runs all lfsTests and verifies the results of the LFS portions of ToOCI and FromOCI.
// It does NOT validate the results of the commit manifest, commit config, and bundles.
func Test_ToFromOCILFS(t *testing.T) { //nolint
	ctx := context.Background()
	ctx = logger.NewContext(ctx, tlog.Logger(t, -2))

	lfsSrc, lfsSrcHandler, srcServer, lfsDst, lfsDstHandler, dstServer := setupLFSServerHandlers(t, ctx)
	defer lfsSrcHandler.cleanup() //nolint
	defer srcServer.Close()
	defer lfsDstHandler.cleanup() //nolint
	defer dstServer.Close()

	target := memory.New() // oci target

	// run LFS tests
	for _, tt := range lfsTests {

		// build a map of expectations which corresponds to argRevList
		reachableLFSFiles, err := lfsSrcHandler.cmdHelper.ListReachableLFSFiles(tt.t.args.argRevList...)
		if err != nil {
			t.Errorf("resolving reachable lfs files: %v", err)
		}
		expectedOIDmap := make(map[string]bool)
		for _, f := range reachableLFSFiles {
			expectedOIDmap[filepath.Base(f)] = false
		}

		t.Run(tt.t.name+":ToOCILFS", func(t *testing.T) {
			toOCIFStorePath := t.TempDir()
			toOCIFStore, err := file.New(toOCIFStorePath)
			if err != nil {
				t.Fatalf("initializing to ocicache file store: %v", err)
			}
			defer toOCIFStore.Close()
			toOCICmdOpts := cmd.Options{
				LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: srcServer.URL},
			}

			syncOpts := SyncOptions{IntermediateDir: toOCIFStorePath, IntermediateStore: toOCIFStore}
			toOCITester, err := NewToOCI(ctx, target, tt.t.args.tag, lfsSrc, tt.t.args.argRevList, syncOpts, &toOCICmdOpts)
			if err != nil {
				t.Errorf("creating ToOCI: %v", err)
			}
			defer toOCITester.Cleanup() //nolint

			commitManDesc, err := toOCITester.Run(ctx)
			if err != nil {
				t.Errorf("ToOCI() error = %v, wantErr %v", err, tt.t.wantErr)
			}

			// fetch it back as our version is not updated when it's sent
			// TODO: Should we update our version of the manifest and config? Safer as a public api, but unnecessary for how we use it in ace-dt.
			_, err = toOCITester.FetchLFSManifestConfig(ctx, commitManDesc, false)
			if err != nil {
				t.Errorf("fetching lfs manifest and config: %v", err)
			}

			err = validateLFSManifestConfig(toOCITester.lfs.manifest, toOCITester.lfs.config, expectedOIDmap)
			if err != nil {
				t.Errorf("validating lfs manifest and config: %v", err)
			}

			if err := toOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up toOCITester handler: %v", err)
			}
		})

		// clear expectedOIDmap
		for oid := range expectedOIDmap {
			expectedOIDmap[oid] = false
		}

		t.Run(tt.t.name+"FromOCILFS", func(t *testing.T) {
			fromOCIFStorePath := t.TempDir()
			fromOCIFStore, err := file.New(fromOCIFStorePath)
			if err != nil {
				t.Fatalf("initializing from oci cache file store: %v", err)
			}
			defer fromOCIFStore.Close()
			fromOCICmdOpts := cmd.Options{
				LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: dstServer.URL},
			}

			syncOpts := SyncOptions{IntermediateDir: fromOCIFStorePath, IntermediateStore: fromOCIFStore}
			fromOCITester, err := NewFromOCI(ctx, target, tt.t.args.tag, lfsDst, syncOpts, &fromOCICmdOpts)
			if err != nil {
				t.Errorf("creating FromOCI: %v", err)
			}
			defer fromOCITester.Cleanup() //nolint

			updatedRefs, err := fromOCITester.Run(ctx)
			if err != nil {
				t.Fatalf("from oci: %v", err)
			}
			t.Logf("updated refs: %s", updatedRefs)

			err = validateLFSRebuild(ctx, lfsDstHandler, expectedOIDmap)
			if err != nil {
				t.Errorf("validating rebuilt lfs repo: %v", err)
			}

			if err := fromOCITester.Cleanup(); err != nil {
				t.Errorf("cleaning up fromOCITester handler: %v", err)
			}
		})
	}

	if err := lfsSrcHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up source handler: %v", err)
	}

	if err := lfsDstHandler.Cleanup(); err != nil {
		t.Errorf("cleaning up destination handler: %v", err)
	}

	t.Log("test complete")
}

// setupLFSServerHandlers sets up src and dst LFS servers as well as handlers to access
// their content. It is the caller's responsibility to close the servers and cleanup the handlers.
func setupLFSServerHandlers(t *testing.T, ctx context.Context) (lfsSrc string, lfsSrcHandler *ToOCI, srcServer *httptest.Server, //nolint
	lfsDst string, lfsDstHandler *FromOCI, dstServer *httptest.Server,
) {
	// setup source repository backed by its own lfs server.
	lfsSrc = t.TempDir()

	// setup src server
	srcServer = SetupLFSServer(t, lfsSrc, "Source")
	t.Logf("Setup git LFS server at %s", srcServer.URL)

	// lfsSrcHandler gives us access to the "srcGitRemote", which we can use to verify the destination repo (lfsDst) is the same as the source (lfsSrc).
	srcSyncOpts := SyncOptions{IntermediateDir: lfsSrc}
	srcCmdOpts := cmd.Options{LFSOptions: &cmd.LFSOptions{WithLFS: true}}
	lfsSrcHandler, err := NewToOCI(ctx, nil, "", "", nil, srcSyncOpts, &srcCmdOpts)
	if err != nil {
		t.Fatalf("creating lfs src handler: %v", err)
	}

	// populate source repository
	if err := createLFSRepo(lfsSrcHandler.cmdHelper); err != nil {
		t.Fatalf("setting up lfs testing repo: %v", err)
	}
	// end source setup

	// setup empty destination repository backed by its own lfs server
	lfsDst = t.TempDir()
	dstServer = SetupLFSServer(t, lfsDst, "Destination")
	t.Logf("Setup git LFS server at %s", dstServer.URL)

	// lfsDstHandler gives us access to the "dstGitRemote", which we can use to verify the destination repo (lfsDst) is the same as the source (lfsSrc)
	dstSyncOpts := SyncOptions{IntermediateDir: lfsDst}
	dstCmdOpts := cmd.Options{LFSOptions: &cmd.LFSOptions{WithLFS: true, ServerURL: dstServer.URL}}
	lfsDstHandler, err = NewFromOCI(ctx, nil, "", "", dstSyncOpts, &dstCmdOpts)
	if err != nil {
		t.Errorf("creating lfs dst handler: %v", err)
	}

	// prepare the destination repository, but don't populate it with anything
	if err := lfsDstHandler.cmdHelper.InitializeRepo(); err != nil {
		t.Errorf("initializing destination repository: %v", err)
	}
	if err := lfsDstHandler.cmdHelper.ConfigureLFS(); err != nil {
		t.Errorf("prepping lfs repo for handling lfs files: %v", err)
	}
	// end destination setup

	return
}

// SetupLFSServer starts a lfs server capable of handling basic lfs client requests according to the batch api.
func SetupLFSServer(t *testing.T, lfsStorage, serverName string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		t.Logf("[%s Server] Received Request: %s", serverName, dump)

		switch r.Method {
		case http.MethodGet: // receive request for lfs file

			oidRelativePath := resolveLFSOIDPath(strings.TrimPrefix(r.URL.Path, "/")) // TODO: this trim may not work on non-UNIX
			oidPath := filepath.Join(lfsStorage, ".git", oidRelativePath)
			if err = copyLFSToResponse(oidPath, w); err != nil {
				t.Errorf("[%s Server] copying lfs file to http response: %v", serverName, err)
			}

		case http.MethodPut: // receive lfs files

			oid := strings.TrimPrefix(r.URL.Path, "/")
			oidPath := filepath.Join(lfsStorage, ".git", resolveLFSOIDPath(oid))

			_, err := os.Stat(oidPath)
			switch {
			case errors.Is(err, fs.ErrNotExist):
				t.Logf("[%s Server]  oid does not yet exist, creating now", serverName)

				if err := copyLFSFromResponse(oidPath, r); err != nil {
					errStr := fmt.Sprintf("[%s Server] copying lfs file from request: %v", serverName, err)
					t.Errorf(errStr)
					http.Error(w, errStr, http.StatusInternalServerError)
				}

			case err != nil:
				errStr := fmt.Sprintf("[%s Server] resolving oid status: %v", serverName, err)
				t.Errorf(errStr)
				http.Error(w, errStr, http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusOK)

		case http.MethodPost: // receive request for action to take

			// unmarshal the body
			var bodyBytes bytes.Buffer
			_, err = io.Copy(&bodyBytes, r.Body)
			if err != nil {
				t.Errorf("[%s Server] copying request body for handling: %v", serverName, err)
			}

			bReq := batchRequest{}
			err = json.Unmarshal(bodyBytes.Bytes(), &bReq)
			if err != nil {
				t.Errorf("[%s Server] decoding request body: %v", serverName, err)
			}

			// we construct a response based on the request, adding extra actions after copying
			bResp := BatchResponse{
				TransferAdapterName: "basic",
				Objects:             bReq.Objects,
				HashAlgorithm:       bReq.HashAlgorithm,
			}

			// handle request for how to perform an action
			t.Logf("[%s Server] Request operation is %s, handling accordingly", serverName, bReq.Operation)
			switch bReq.Operation {
			case downloadAction:

				setupActions(downloadAction, r.Host, bReq.Objects)

			case uploadAction:

				setupActions(uploadAction, r.Host, bReq.Objects)

			default:
				t.Logf("[%s Server] Unknown request operation: %s", serverName, bReq.Operation)
			}

			// respond to the client
			bRespBytes, err := json.Marshal(bResp)
			if err != nil {
				t.Errorf("[%s Server] encoding response json: %v", serverName, err)
			}

			w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
			_, err = w.Write(bRespBytes) // calls w.WriteHeader with ok status before writing body
			if err != nil {
				t.Errorf("[%s Server] writing response: %v", serverName, err)
			}

		default:
			http.Error(w, fmt.Sprintf("[%s Server] unexpected request type: %s", serverName, r.Method), http.StatusInternalServerError)
		}
	}))
}

// setupActions sets up batch api Transfer objects for the intended actions.
func setupActions(actionType, host string, objs []*Transfer) {
	for _, obj := range objs {
		obj.Authenticated = true
		obj.Actions = make(ActionSet)
		href := url.URL{
			Scheme: "http",
			Host:   host,
			Path:   obj.Oid,
		}
		action := Action{Href: href.String()}
		obj.Actions[actionType] = &action
	}
}

// copyLFSToResponse copies an lfs file to an http response.
func copyLFSToResponse(oidPath string, w http.ResponseWriter) error {
	oidInfo, err := os.Stat(oidPath)
	if err != nil {
		return fmt.Errorf("getting stats of oid file: %w", err)
	}

	oidFile, err := os.Open(oidPath)
	if err != nil {
		return fmt.Errorf("opening oid file: %w", err)
	}
	defer oidFile.Close()

	// respond to the client with the file contents
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", oidInfo.Size()))

	n, err := io.Copy(w, oidFile)
	if err != nil || n != oidInfo.Size() {
		return fmt.Errorf("copying file to response body: %w", err)
	}

	return nil
}

// copyLFSFromResponse copies the body of a request to an lfs file.
func copyLFSFromResponse(oidPath string, r *http.Request) error {
	if err := os.MkdirAll(filepath.Dir(oidPath), 0o777); err != nil {
		return fmt.Errorf("creating path to new oid location: %w", err)
	}

	oidFile, err := os.Create(oidPath)
	if err != nil {
		return fmt.Errorf("creating oid file: %w", err)
	}

	n, err := io.Copy(oidFile, r.Body)
	if err != nil {
		return fmt.Errorf("copying to oid file: %w", err)
	}
	if n != r.ContentLength {
		return fmt.Errorf("copied %d of %d bytes to oid file", n, r.ContentLength)
	}
	if err := oidFile.Close(); err != nil {
		return fmt.Errorf("closing oid file: %w", err)
	}

	return nil
}

// From git-lfs transfer batch api: https://github.com/git-lfs/git-lfs/blob/main/tq/api.go
// not supposed to be used outside of git-lfs package.
type BatchResponse struct {
	Objects             []*Transfer `json:"objects"`
	TransferAdapterName string      `json:"transfer"`
	HashAlgorithm       string      `json:"hash_algo"`
	// endpoint            lfshttp.Endpoint
}

type Transfer struct {
	// Name          string `json:"name,omitempty"`
	Oid           string    `json:"oid,omitempty"`
	Size          int64     `json:"size"`
	Authenticated bool      `json:"authenticated,omitempty"`
	Actions       ActionSet `json:"actions,omitempty"`
	// Links         ActionSet    `json:"_links,omitempty"`
	// Error         *ObjectError `json:"error,omitempty"`
	// Path    string `json:"path,omitempty"`
	// Missing bool   `json:"-"`
}

type ActionSet map[string]*Action

type Action struct {
	Href string `json:"href"`
	// Header    map[string]string `json:"header,omitempty"`
	// ExpiresAt time.Time         `json:"expires_at,omitempty"`
	// ExpiresIn int               `json:"expires_in,omitempty"`
	// Id        string            `json:"-"`
	// Token     string            `json:"-"`

	// createdAt time.Time
}

const (
	uploadAction   = "upload"
	downloadAction = "download"
)

type batchRef struct {
	Name string `json:"name,omitempty"`
}

type batchRequest struct {
	Operation            string      `json:"operation"`
	Objects              []*Transfer `json:"objects"`
	TransferAdapterNames []string    `json:"transfers,omitempty"`
	Ref                  *batchRef   `json:"ref"`
	HashAlgorithm        string      `json:"hash_algo"`
}

// validateLFSManifestConfig validates the LFS manifest by ensuring the manifest layers match exactly what's expected.
// Note: The config is empty, so it's not validated.
func validateLFSManifestConfig(lfsManifest ocispec.Manifest, config oci.LFSConfig, expectedOIDMap map[string]bool) error {

	configErrs := make([]error, 0)

	// ensure layers match the config and they were expected
	for _, layer := range lfsManifest.Layers {

		oid := layer.Annotations[ocispec.AnnotationTitle]

		// was this layer, holding an oid, expected?
		if _, ok := expectedOIDMap[oid]; !ok {
			configErrs = append(configErrs, fmt.Errorf("unexpected oid found in config: %s", oid))
		} else {
			expectedOIDMap[oid] = true
		}

	}

	// did we find all expected oid's?
	for expectedOID, found := range expectedOIDMap {
		if !found {
			configErrs = append(configErrs, fmt.Errorf("expected oid %s not found in config", expectedOID))
		}
	}

	return errors.Join(configErrs...)
}

// validateLFSRebuild validates the destination git repository by walking the lfs/objects directory, ensuring it matches what's expected.
func validateLFSRebuild(ctx context.Context, lfsDstHandler *FromOCI, expectedOIDs map[string]bool) error {
	var rebuildErrs []error

	// fetch all lfs files, as these should be pushed to the server but not exist in the destination as this is the purpose of lfs.
	err := lfsDstHandler.cmdHelper.LFS.Fetch(lfsDstHandler.syncOpts.IntermediateDir, "--all")
	if err != nil {
		return fmt.Errorf("fetching all lfs files: %w", err)
	}
	relativelfsObjsPath := filepath.Join(lfsDstHandler.syncOpts.IntermediateDir, ".git", cmd.LFSObjsPath)
	lfsObjsFS := os.DirFS(relativelfsObjsPath)

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := fs.Stat(lfsObjsFS, path)
		if err != nil {
			return fmt.Errorf("unable to get the actual file info: %w", err)
		}

		// we don't care about directories, just the files in them
		if info.IsDir() {
			return nil
		}

		if _, ok := expectedOIDs[info.Name()]; !ok {
			return fmt.Errorf("unexpected LFS file with oid '%s' found", info.Name())
		}

		expectedOIDs[info.Name()] = true

		return nil
	}

	// begin walk
	if err := fs.WalkDir(lfsObjsFS, ".", walkFn); err != nil { // TODO: loading entire obj dir into memory may be problematic
		return fmt.Errorf("walking LFS destination obj directory: %w", err)
	}

	// did we find all expected OID's?
	for expectedOID, found := range expectedOIDs {
		if !found {
			rebuildErrs = append(rebuildErrs, fmt.Errorf("expected LFS file with oid '%s' not found", expectedOID))
		}
	}

	return errors.Join(rebuildErrs...)
}

func resolveLFSOIDPath(oid string) string {
	return filepath.Join(cmd.LFSObjsPath, oid[0:2], oid[2:4], oid) // oid "abcdef" -> ab/cd/abcdef
}
