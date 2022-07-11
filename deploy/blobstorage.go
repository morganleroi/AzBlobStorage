package deploy

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"log"
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
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccountName), credential, nil)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	client, err := serviceClient.NewContainerClient(containerWeb)
	if err != nil {
		log.Fatalf("Unable to create a client on %s container", containerWeb)
	}

	_, err = client.GetProperties(ctx, nil)
	if err != nil {
		log.Fatalf("Error when fetching properties on the storage account %s with the following error \n %v ", containerWeb, err)
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
