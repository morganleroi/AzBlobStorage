package deploy

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"os"
)

const containerWeb, fileToCheck, metadataToCheck = "$web", "index.html", "version"

func GetDeployedPackage(storageAccountName string, storageAccountKey string) (string, error) {
	os.Setenv("AZURE_STORAGE_ACCOUNT_KEY", storageAccountKey)
	return isAlreadyDeployed(storageAccountName)
}

func isAlreadyDeployed(storageAccountName string) (string, error) {
	accountKey, ok := os.LookupEnv("AZURE_STORAGE_ACCOUNT_KEY")

	if !ok {
		println("AZURE_STORAGE_ACCOUNT_KEY could not be found")
	}

	ctx := context.Background()

	credential, err := azblob.NewSharedKeyCredential(storageAccountName, accountKey)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Invalid credentials with error: %s", err.Error()))
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName), credential, nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Invalid credentials with error: %s", err.Error()))
	}

	client, err := serviceClient.NewContainerClient(containerWeb)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to create a client on %s container with error: %s", containerWeb, err.Error()))
	}

	_, err = client.GetProperties(ctx, nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error when fetching properties on the storage account %s with the following error \n %v ", containerWeb, err.Error()))
	}

	pager := client.ListBlobsFlat(&azblob.ContainerListBlobsFlatOptions{
		Include: []azblob.ListBlobsIncludeItem{"metadata"},
	})

	var deployedVersion string

	fileIsFound, metadataIsFound := false, false
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			if *v.Name == fileToCheck {
				fileIsFound = true
				for key, v := range v.Metadata {
					if key == metadataToCheck {
						deployedVersion = *v
						metadataIsFound = true
					}
				}
				if metadataIsFound {
					return deployedVersion, nil
				}
			}
		}
	}

	if !fileIsFound {
		return deployedVersion, errors.New(fmt.Sprintf("Unable to find %s file", fileToCheck))
	}
	return deployedVersion, errors.New(fmt.Sprintf("Unable to find %s metadata in %s file", metadataToCheck, fileToCheck))

}
