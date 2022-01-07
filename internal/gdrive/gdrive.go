package gdrive

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"pupadrive.go/internal/configs"
	"pupadrive.go/internal/logger"
)

type Drive struct {
	Client *drive.Service
}

func (d *Drive) UploadReader(r io.Reader, fileName string, parentFolder string, progress func(current, total int64)) (string, error) {
	res, err := d.Client.Files.Create(&drive.File{Parents: []string{parentFolder}, Name: fileName}).
		Media(r).SupportsAllDrives(true).ProgressUpdater(progress).Do()

	return res.Id, err
}

func (d *Drive) UploadFile(filePath string, parentFolder string, progress func(current, total int64)) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}

	defer file.Close()
	fileName := filepath.Base(file.Name())

	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}

	return d.UploadReader(file, fileName, parentFolder, progress)
}

func (d *Drive) CreateFolder(name string, parent string) (string, error) {
	if parent == "" {
		parent = configs.GetTeamDriveID()
	}
	resp, err := d.Client.Files.Create(&drive.File{Name: name, MimeType: "application/vnd.google-apps.folder", Parents: []string{parent}}).SupportsAllDrives(true).Do()
	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}
	return resp.Id, nil
}

type Progress struct {
	current, total int64
}

func (d *Drive) UploadFolder(dirPath string, parent string, progress func(current, total int64)) error {
	p := Progress{
		current: 0,
		total:   0,
	}
	return d.uploadFolderRecursive(dirPath, parent, progress, &p)
}

func (d *Drive) uploadFolderRecursive(dirPath string, parent string, progress func(current, total int64), p *Progress) error {
	err := filepath.WalkDir(dirPath, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			logger.ErrorLogger.Println(err)
			return err
		}

		if dirPath == path {
			return nil
		}

		if e.IsDir() {
			id, err := d.CreateFolder(e.Name(), parent)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}

			err = d.uploadFolderRecursive(fmt.Sprintf("%s/%s", path, e.Name()), id, progress, p)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}

		} else {
			fileCurrent, fileTotal := int64(0), int64(0)
			d.UploadFile(path, parent, func(c, t int64) {
				p.current += c - fileCurrent
				p.total += t - fileTotal
				fileCurrent, fileTotal = c, t
				progress(p.current, p.total)
			})
		}

		return nil
	})
	logger.ErrorLogger.Println(err)
	return err
}

func (d *Drive) CheckFolderExist(name string, parent string) (string, error) {
	if parent == "" {
		parent = configs.GetTeamDriveID()
	}
	query := fmt.Sprintf("'%s' in parents and name='%s' and mimeType='application/vnd.google-apps.folder' and trashed = false", parent, name)
	exists, err := d.Client.Files.List().Q(query).DriveId(configs.GetTeamDriveID()).IncludeItemsFromAllDrives(true).Corpora("drive").
		Spaces("drive").SupportsAllDrives(true).Do()
	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}
	if len(exists.Files) > 0 {
		return exists.Files[0].Id, nil
	}
	return "", nil

}

func (d *Drive) CheckFileExist(name string) (string, error) {

	query := fmt.Sprintf("name='%s' and trashed = false", name)
	exists, err := d.Client.Files.List().Q(query).DriveId(configs.GetTeamDriveID()).IncludeItemsFromAllDrives(true).Corpora("drive").
		Spaces("drive").SupportsAllDrives(true).Do()

	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", err
	}

	if len(exists.Files) > 0 {
		return exists.Files[0].Id, nil
	}
	return "", nil

}
