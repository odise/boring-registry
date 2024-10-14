package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"log"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	containerName := "boring-registry-container"
	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", "boringregistry2483", containerName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	handleError(err)

	containerClient, err := container.NewClient(containerURL, cred, nil)
	handleError(err)

	pager := containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{Snapshots: true, Versions: true},
	})

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		for _, blob := range resp.Segment.BlobItems {
			fmt.Println(*blob.Name)
		}
	}
}
