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
	TrackerFile string
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
	var counter int
	// create a tracker map, imageName:slice of existing images.
	// iterate the counter for serialize command to create the new file.
	trackerMap := map[string][]string{}
	// TODO: trackerfile name is assumed. add option to overwrite with a flag.
	trackerFile, err := os.Open(action.TrackerFile)
	if err != nil {
		return fmt.Errorf("opening record-keeping file: %w", err)
	}
	// trackerFile is also a csv with the format SYNC_NAME, IMAGE, DIGEST. We are reading now but will need to write to it at the end of each serialize operation.
	tr := csv.NewReader(trackerFile)
	trackerRecords, err := tr.ReadAll()
	if err != nil {
		return fmt.Errorf("reading the record-keeping file: %w", err)
	}
	// if err := trackerFile.Close(); err != nil {
	// 	return fmt.Errorf("closing record-keeping file: %w", err)
	// }
	defer trackerFile.Close()
	// iterate through records, skip the first line, load the data into the map.
	for i := 1; i <= len(trackerRecords); i++ {
		record := records[i]
		// record[0] | record[1] | record[2]
		// name     | images[]  |  digest
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
		_, ok := trackerMap[name[1]]
		if ok {
			// we need to assert that the new sync number is greater than the existing and then replace
			if syncNumber < counter {
				continue
			}
		}
		counter = syncNumber
		trackerMap[name[1]] = append(trackerMap[name[1]], record[1])
		// trackerMap[name[1]] = syncNumber
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
			ExistingImages:      trackerMap[imgName],
			Recursive:           action.Recursive,
			RepoFunc:            action.Config.Repository,
			SourceRepo:          repo,
			SourceReference:     image,
		}
		// new image name
		newSyncNumber := counter + 1
		// convert to string
		fileName := strings.Join([]string{strconv.Itoa(newSyncNumber), imgName}, "-")
		fileName = filepath.Join(syncDir, "data", fileName)
		if err := mirror.Serialize(ctx, fileName, "", action.DataTool.Version(), opts); err != nil {
			return err
		}
		// iterate the counter if serialize is successful
		counter++
		// get the reference digest
		digest, err := repo.Reference.Digest()
		if err != nil {
			return fmt.Errorf("getting repository digest: %w", err)
		}
		// add it to the tracker file
		tw := csv.NewWriter(trackerFile)
		tw.Write([]string{fileName, image, digest.String()})
	}

	return nil
}
