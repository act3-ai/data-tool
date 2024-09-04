package mirror

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// BatchSerialize represents the mirror batch-serialize action.
type BatchSerialize struct {
	*Action
	TrackerFile      string
	Compression      string
	WithManifestJSON bool
}

// Run runs the mirror batch-serialize action.
func (action *BatchSerialize) Run(ctx context.Context, gatherList, syncDir string) error {
	log := logger.FromContext(ctx)
	// navigate to syncDir and
	// if trackerFile exists, open it.
	f, err := os.Open(gatherList)
	if err != nil {
		return fmt.Errorf("opening gather list file: %w", err)
	}
	defer f.Close()
	// we need to read the file and get the previous images that were synced
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.Comment = '#'
	r.TrimLeadingSpace = true
	// create a map for easy retrieval
	images := map[string]string{}
	var entries int
	// iterate through records, skip the first line, load the data into the map.
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading gather list file: %w", err)
		}
		if entries != 0 {
			// skip the first line with the headers
			// name, images
			images[record[0]] = record[1]
		}
		entries++
	}

	var counter int
	// create a tracker map, imageName:slice of existing images.
	// iterate the counter for serialize command to create the new file.

	trackerMap := make(map[string]map[string]string)

	trackerFile, err := os.OpenFile(filepath.Join(syncDir, action.TrackerFile), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("opening the record-keeping file: %w", err)
	}
	defer trackerFile.Close()

	// create the writer, but we want to read the contents into memory first before writing
	tw := csv.NewWriter(trackerFile)
	defer tw.Flush()

	// trackerFile is also a csv with the format SYNC_NAME, IMAGE, DIGEST. We are reading now but will need to write to it at the end of each serialize operation.
	tr := csv.NewReader(trackerFile)
	tr.FieldsPerRecord = -1
	tr.Comment = '#'
	tr.TrimLeadingSpace = true
	trackerRecords, err := tr.ReadAll()
	if err != nil {
		return fmt.Errorf("reading the record-keeping file: %w", err)
	}

	if len(trackerRecords) == 0 {
		err := tw.Write([]string{"sync_name", "artifact", "digest"})
		if err != nil {
			return fmt.Errorf("writing csv header: %w", err)
		}
	}

	// iterate through records, skip the first line, load the data into the map.
	for i := 1; i <= len(trackerRecords)-1; i++ {
		record := trackerRecords[i]
		// record[0] | record[1] | record[2]
		// name     | image      |  digest
		// we want to split the tar file name so that we get the name and the int
		// expected tar file name examples: 0-name1.tar, 1-name2.tar, etc...
		name := strings.Split(record[0], "-")
		if len(name) != 2 {
			return fmt.Errorf("unexpected tar file name %s", record[0])
		}
		syncNumber, err := strconv.Atoi(name[0])
		if err != nil {
			return fmt.Errorf("splitting on record name %s: %w", record[0], err)
		}
		imgName := strings.Split(name[1], ".")[0]
		_, ok := trackerMap[imgName]
		if ok {
			// we need to assert that the new sync number is greater than the existing and then replace
			if syncNumber < counter {
				continue
			}
		} else {
			trackerMap[imgName] = make(map[string]string)
		}
		counter = syncNumber
		trackerMap[imgName][record[1]] = record[0]

	}
	for imgName, image := range images {
		// was this image previously serialized?
		existingFile, ok := trackerMap[imgName][image]
		if ok {
			log.InfoContext(ctx, "Artifact exists in sync directory", "fileName", existingFile)
			continue
		}
		// get the existing images list
		existingImages := generateExistingImagesList(trackerMap[imgName])
		// create the image target
		repo, err := action.Config.Repository(ctx, image)
		if err != nil {
			return err
		}
		// for each image, create the serialize options
		opts := mirror.SerializeOptions{
			BufferOpts:          mirror.BlockBufOptions{}, // I think this should be empty, this feature shouldn't be used with a tape drive right?
			ExistingCheckpoints: nil,
			ExistingImages:      existingImages,
			Recursive:           action.Recursive,
			RepoFunc:            action.Config.Repository,
			SourceRepo:          repo,
			SourceReference:     image,
			Compression:         action.Compression,
			WithManifestJSON:    action.WithManifestJSON,
		}
		// new image name
		newSyncNumber := counter + 1
		// convert to string
		fileName := fmt.Sprintf("%06d-%s", newSyncNumber, imgName)
		var extension string
		switch action.Compression {
		case "zstd":
			extension = "tar.zst"
		case "gzip":
			extension = "tar.gz"
		default:
			extension = "tar"
		}
		fileName = filepath.Join(syncDir, strings.Join([]string{fileName, extension}, "."))
		log.InfoContext(ctx, "serializing artifact to file:", "artifactName", imgName, "file", fileName)
		if err := mirror.Serialize(ctx, fileName, "", action.DataTool.Version(), opts); err != nil {
			return err
		}
		// iterate the counter if serialize is successful
		counter++
		// get the reference digest
		desc, err := repo.Resolve(ctx, image)
		if err != nil {
			return fmt.Errorf("getting repository descriptor: %w", err)
		}
		// add it to the tracker file
		if err = tw.Write([]string{filepath.Base(fileName), image, desc.Digest.String()}); err != nil {
			return fmt.Errorf("writing to record keeping file: %w", err)
		}
	}

	return nil
}

func generateExistingImagesList(images map[string]string) []string {
	existingImages := make([]string, 0, len(images))
	for image := range images {
		existingImages = append(existingImages, image)
	}
	return existingImages
}
