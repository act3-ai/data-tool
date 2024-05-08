package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"git.act3-ace.com/ace/data/schema/pkg/util"
)

// ErrNonPortablePath is used to denote a filename/path that contains invalid characters.
var ErrNonPortablePath = errors.New("non-portable path")

// TarArchiverOption is used for options for the archival process.
type TarArchiverOption func(ar *TarArchiver) error

// WithExecutableFiles perserves executable file within the archive (the default is to remove executable permissions on all files).
func WithExecutableFiles() TarArchiverOption {
	return func(ar *TarArchiver) error {
		ar.preserveExecutableFiles = true
		return nil
	}
}

// WithPortablePaths ensures that all paths within the archive are portable paths (for cross platform compatibility)
// We do not want to have file within archives having bad file names that do not work on Windows.
func WithPortablePaths() TarArchiverOption {
	return func(ar *TarArchiver) error {
		ar.portablePaths = true
		return nil
	}
}

// TarArchiver archives a directory to a stream.
type TarArchiver struct {
	// tw is a writer
	tw *tar.Writer

	// preserveExecutableFiles will set executable bits when the file is executable.  Only applies to files.
	preserveExecutableFiles bool

	// portablePaths ensures all paths are portable POSIX paths
	portablePaths bool
}

// Close closes the writer if present, and clears any existing stream objects.
func (ar *TarArchiver) Close() error {
	if ar.tw != nil {
		if err := ar.tw.Close(); err != nil {
			return fmt.Errorf("failed to close tar archiver: %w", err)
		}
	}
	ar.tw = nil
	return nil
}

// NewArchiver initializes a new archive and prepares it for writing.
func NewArchiver(ostream io.Writer, options ...TarArchiverOption) (*TarArchiver, error) {
	ar := &TarArchiver{
		tw: tar.NewWriter(ostream),
	}

	for _, o := range options {
		if err := o(ar); err != nil {
			return nil, err
		}
	}

	return ar, nil
}

// WriteEntry adds file or directory data to the archive.
// finfo is expected to be the FileInfo for the actual file or directory (not a symlink to a file or directory).
func (ar *TarArchiver) WriteEntry(fsys fs.FS, path string, finfo fs.FileInfo) error {
	if ar.tw == nil {
		return fmt.Errorf("tar archive was not created for writing first")
	}

	if ar.portablePaths {
		if !util.IsPortablePath(path) {
			return fmt.Errorf("path \"%s\" contains invalid characters: %w", path, ErrNonPortablePath)
		}
	}

	fm := finfo.Mode()

	hdr := &tar.Header{
		Name:    path,
		ModTime: time.Unix(0, 0).UTC(),
	}

	switch {
	case fm.IsRegular():
		hdr.Typeflag = tar.TypeReg
		hdr.Size = finfo.Size()
		if ar.preserveExecutableFiles && fm.Perm()&0111 != 0 {
			hdr.Mode = 0777
		} else {
			hdr.Mode = 0666
		}
	case finfo.IsDir():
		hdr.Typeflag = tar.TypeDir
		hdr.Name += "/"
		hdr.Mode = 0777
	default:
		return fmt.Errorf("file mode %v not supported", fm)
	}

	if err := ar.tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("writing header for path %s: %w", hdr.Name, err)
	}

	if !finfo.IsDir() {
		// write the file contents
		rdr, err := fsys.Open(path)
		if err != nil {
			return fmt.Errorf("%s: no way to read file contents: %w", path, err)
		}
		defer rdr.Close()

		_, err = io.Copy(ar.tw, rdr)
		if err != nil {
			return fmt.Errorf("%s: copying contents: %w", path, err)
		}
		if err := rdr.Close(); err != nil {
			return fmt.Errorf("failed to close source %s: %w", path, err)
		}
	}

	return nil
}

// TarFileData provides methods for accessing information about
// or contents of a file within an archive.
type TarFileData struct {
	fs.FileInfo

	// The original header info; depends on
	// type of archive -- could be nil, too.
	Header *tar.Header

	// Allow the file contents to be read (and closed)
	io.ReadCloser
}

// TarExtractor is used for reading a tar archive from a stream and extracting to disk.
type TarExtractor struct {
	// tr is the reader for the tar archive
	tr *tar.Reader

	// OverwriteExisting files when extracting
	OverwriteExisting bool

	// MakeParentDirectories will make all parents instead of just what should be necessary (set to true for backwards compatibility)
	MakeParentDirectories bool
}

// Open opens an archive stream for reading.
func (ar *TarExtractor) Open(in io.Reader) error {
	if ar.tr != nil {
		return fmt.Errorf("tar archive is already open for reading")
	}
	ar.tr = tar.NewReader(in)
	return nil
}

// Read reads the next file from t, which must have
// already been opened for reading. If there are no
// more files, the error.Is io.EOF. The File must
// be closed when finished reading from it.
func (ar *TarExtractor) Read() (*TarFileData, error) {
	if ar.tr == nil {
		return nil, fmt.Errorf("tar archive is not open")
	}

	hdr, err := ar.tr.Next()
	if err != nil {
		return nil, fmt.Errorf("error reading tar archive: %w", err)
	}

	file := &TarFileData{
		FileInfo:   hdr.FileInfo(),
		Header:     hdr,
		ReadCloser: io.NopCloser(ar.tr),
	}

	return file, nil
}

// UnarchiveRead implements tar unarchiving from an io.reader stream to
// a destination path.
func (ar *TarExtractor) UnarchiveRead(sourceReader io.Reader, destPath string) error {
	err := ar.Open(sourceReader)
	if err != nil {
		return fmt.Errorf("opening tar archive for reading: %w", err)
	}

	for {
		err := ar.untarNext(destPath)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading file in tar archive: %w", err)
		}
	}

	return nil
}

func (ar *TarExtractor) untarNext(destination string) error {
	f, err := ar.Read()
	if err != nil {
		return err
	}
	header := f.Header
	return ar.untarEntry(f, destination, header)
}

func (ar *TarExtractor) untarEntry(f *TarFileData, destination string, hdr *tar.Header) error {
	to := filepath.Join(destination, hdr.Name)

	switch hdr.Typeflag {
	case tar.TypeDir:
		// create the directory (the parent(s) should already exist)
		var err error
		if ar.MakeParentDirectories {
			err = os.MkdirAll(to, 0777)
		} else {
			err = os.Mkdir(to, 0777)
		}
		if err != nil {
			return fmt.Errorf("tar extraction: %w", err)
		}
		return nil
	// we only store regular files in the archive
	case tar.TypeReg:
		return ar.writeNewFile(to, f)
	default:
		return fmt.Errorf("%s has an unknown TAR type %c", hdr.Name, hdr.Typeflag)
	}
}

func (ar *TarExtractor) writeNewFile(fpath string, in io.Reader) error {
	if !ar.OverwriteExisting {
		_, err := os.Stat(fpath)
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("file %s already exists but we are not overwriting files: %w", fpath, err)
		}
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %w", fpath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %w", fpath, err)
	}
	return nil
}
