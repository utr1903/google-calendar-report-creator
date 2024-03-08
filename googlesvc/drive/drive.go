package drive

import (
	"context"
	"log"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const FOLDER_NAME = "MyEffortReports"
const FOLDER_MIME_TYPE = "application/vnd.google-apps.folder"

type GoogleDriveService struct {
	client *http.Client
}

func New(
	client *http.Client,
) *GoogleDriveService {
	return &GoogleDriveService{
		client: client,
	}
}

func (s *GoogleDriveService) MoveFile(
	ctx context.Context,
	fileId string,
) error {

	// Create Drive service
	svc, err := s.createDriveService(ctx)
	if err != nil {
		return err
	}

	// Get the file from Drive
	file, err := s.getFile(svc, fileId)
	if err != nil {
		return err
	}

	// Create necessary folder to move if not exists
	folderId, err := s.createFolderIfNotExists(svc)
	if err != nil {
		return err
	}

	// Move file to folder
	err = s.moveFile(svc, file, folderId)
	if err != nil {
		return err
	}

	return nil
}

func (s *GoogleDriveService) createDriveService(
	ctx context.Context,
) (
	*drive.Service,
	error,
) {
	// Create Drive service
	svc, err := drive.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
		return nil, err
	}
	return svc, nil
}

func (s *GoogleDriveService) createFolderIfNotExists(
	svc *drive.Service,
) (
	string,
	error,
) {
	// Check if the folder exists
	folders, err := svc.Files.List().
		Q("trashed = false and mimeType:'" + FOLDER_MIME_TYPE + "' and name='" + FOLDER_NAME + "'").
		Do()
	if err != nil {
		log.Fatalf("Unable to list folders: %v", err)
		return "", nil
	}

	// Create a new folder
	if len(folders.Files) == 0 {
		folder := &drive.File{
			Name:     FOLDER_NAME,
			MimeType: FOLDER_MIME_TYPE,
		}
		newFolder, err := svc.Files.Create(folder).Do()
		if err != nil {
			log.Fatalf("Unable to create folder: %v", err)
			return "", err
		}
		return newFolder.Id, nil
	}

	return folders.Files[0].Id, nil
}

func (s *GoogleDriveService) getFile(
	svc *drive.Service,
	fileId string,
) (
	*drive.File,
	error,
) {
	file, err := svc.Files.Get(fileId).Do()
	if err != nil {
		log.Fatalf("Unable to get Drive file: %v", err)
		return nil, err
	}
	return file, nil
}

func (s *GoogleDriveService) moveFile(
	svc *drive.Service,
	file *drive.File,
	folderId string,
) error {

	// Move the created sheet to the specified folder
	var newFile drive.File
	_, err := svc.Files.Update(file.Id, &newFile).
		AddParents(folderId).
		Do()
	if err != nil {
		log.Fatalf("Unable to move file to folder: %v", err)
		return err
	}

	return nil
}
