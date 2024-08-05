package mirror

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
)

type BatchSerialize struct {
	*Action
}

func (action *BatchSerialize) Run(ctx context.Context, gatherList, syncDir string) error {
	// log := logger.FromContext(ctx)
	// cfg := action.Config.Get(ctx)
	// navigate to syncDir and
	// if trackerFile exists, open it.
	f, err := os.Open(gatherList)
	if err != nil {
		return fmt.Errorf("opening gather list file: %w", err)
	}
	defer f.Close()
	// we need to open the file for reading and get the previous digests that were synced
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("reading gather list file: %w", err)
	}
	// create a map for easy retrieval
	images := make(map[string]string, len(records)-1)
	// iterate through records, skip the first line, load the data into the map.
	for i := 1; i <= len(records); i++ {
		// name, images
		record := records[i]
		images[record[0]] = images[record[1]]
	}
	// create a tracker map, imageName:sync number
	var trackerMap map[string]int
	// TODO look through trackerfile to find the last tag for each image name and increment by 1 (and make sure it doesn't exist)
	// trackerfile name is assumed? can be overridden with a flag?
	trackerFile, err := os.Open(filepath.Join(syncDir, "record_keeping.csv"))
	if err != nil {
		return fmt.Errorf("opening record-keeping file: %w", err)
	}
	// trackerFile is also a csv with the format SYNC_NAME, IMAGE, DIGEST. We are reading now but will need to write to it at the end of each serialize operation.
	tr := csv.NewReader(trackerFile)
	trackerRecords, err := tr.ReadAll()
	if err != nil {
		return fmt.Errorf("reading the record-keeping file: %w", err)
	}
	// iterate through records, skip the first line, load the data into the map.
	for i := 1; i <= len(trackerRecords); i++ {
		// name, images
		record := records[i]
		// we want to split the record so that we get the name and the int
		name := strings.Split(record[0], "-")
		if len(name) != 2 {
			return fmt.Errorf("malformed tracker file name %s", record[0])
		}
		syncNumber, err := strconv.Atoi(name[1])
		if err != nil {
			return fmt.Errorf("splitting on record name %s: %w", record[0], err)
		}
		v, ok := trackerMap[name[0]]
		if ok {
			// we need to assert that the new sync number is greater than the existing and then replace
			if syncNumber < v {
				continue
			}
		}
		trackerMap[name[0]] = syncNumber
	}
	for imgName, image := range images {
		// create the image target
		repo, err := action.Config.Repository(ctx, image)
		if err != nil {
			return err
		}
		// for each image, create the serialize options
		opts := mirror.SerializeOptions{
			BufferOpts:          mirror.BlockBufOptions{}, // I think this should be empty, this feature shouldn't be used with a tape drive right?
			ExistingCheckpoints: nil,
			ExistingImages:      []string{}, //TODO
			Recursive:           action.Recursive,
			RepoFunc:            action.Config.Repository,
			SourceRepo:          repo,
			SourceReference:     image,
		}
		// new image name
		newSyncNumber := trackerMap[imgName] + 1
		// convert to string
		fileName := strings.Join([]string{imgName, strconv.Itoa(newSyncNumber)}, "-")
		fileName = filepath.Join(syncDir, fileName)
		if err := mirror.Serialize(ctx, fileName, "", action.DataTool.Version(), opts); err != nil {
			return err
		}
	}

	return nil
}
