package bottle

import (
	"errors"
)

var (

	// ErrDirNotEmpty is return when a directory that is found not to be empty.
	ErrDirNotEmpty = errors.New("directory is not empty")

	// ErrPartInsecureArchive is an error that indicates that a part contains a directory archive but does not include
	// a trailing slash in the name. This can enable aliasing of files and directories by an attacker or otherwise, as
	// mentioned in https://git.act3-ace.com/ace/data/tool/-/issues/314
	ErrPartInsecureArchive = errors.New("part is named as regular file, but is an archived directory media type")
)
