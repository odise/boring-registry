package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"io"
	"path/filepath"
	"time"

	"github.com/boring-registry/boring-registry/pkg/core"
	"github.com/go-kit/log"
)

// BlobStorage is a Storage implementation backed by GCS.
// BlobStorage implements module.Storage and provider.Storage
type BlobStorage struct {
	sc                  *container.Client
	accountName         string
	containerName       string
	bucketPrefix        string
	signedURLExpiry     time.Duration
	serviceAccount      string
	moduleArchiveFormat string
}

func (s *BlobStorage) GetModule(ctx context.Context, namespace, name, provider, version string) (core.Module, error) {
	key := modulePath(s.bucketPrefix, namespace, name, provider, version, s.moduleArchiveFormat)

	pager := s.sc.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Include:    container.ListBlobsInclude{Snapshots: true, Versions: true},
		Marker:     nil,
		MaxResults: nil,
		Prefix:     &key,
	})

	/*
		pager := s.sc.NewListBlobsFlatPager(s.containerName, &azblob.ListBlobsFlatOptions{
			Include:    container.ListBlobsInclude{},
			Marker:     nil,
			MaxResults: nil,
			Prefix:     &key,
		})

	*/
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return core.Module{}, fmt.Errorf("failed to iterate over results: %w", err)
		}
		for _, _blob := range resp.Segment.BlobItems {
			fmt.Printf("%v", _blob.Name)
		}
	}

	fmt.Print(key, pager)
	return core.Module{}, nil
}

func (s *BlobStorage) ListModuleVersions(ctx context.Context, namespace, name, provider string) ([]core.Module, error) {
	var modules []core.Module
	return modules, nil
}

func (s *BlobStorage) UploadModule(ctx context.Context, namespace, name, provider, version string, body io.Reader) (core.Module, error) {
	if namespace == "" {
		return core.Module{}, errors.New("namespace not defined")
	}

	if name == "" {
		return core.Module{}, errors.New("name not defined")
	}

	if provider == "" {
		return core.Module{}, errors.New("provider not defined")
	}

	if version == "" {
		return core.Module{}, errors.New("version not defined")
	}

	return s.GetModule(ctx, namespace, name, provider, version)
}

func (s *BlobStorage) MigrateModules(ctx context.Context, logger log.Logger, dryRun bool) error {
	return nil
}

// MigrateProviders is a temporary method needed for the migration from 0.7.0 to 0.8.0 and above
func (s *BlobStorage) MigrateProviders(ctx context.Context, logger log.Logger, dryRun bool) error {
	return nil
}

// GetProvider implements provider.Storage
func (s *BlobStorage) getProvider(ctx context.Context, pt providerType, provider *core.Provider) (*core.Provider, error) {
	return provider, nil
}

func (s *BlobStorage) GetProvider(ctx context.Context, namespace, name, version, os, arch string) (*core.Provider, error) {
	p, err := s.getProvider(ctx, internalProviderType, &core.Provider{
		Namespace: namespace,
		Name:      name,
		Version:   version,
		OS:        os,
		Arch:      arch,
	})
	return p, err
}

func (s *BlobStorage) GetMirroredProvider(ctx context.Context, provider *core.Provider) (*core.Provider, error) {
	return s.getProvider(ctx, mirrorProviderType, provider)
}

func (s *BlobStorage) listProviderVersions(ctx context.Context, pt providerType, provider *core.Provider) ([]*core.Provider, error) {
	var providers []*core.Provider
	return providers, nil
}

func (s *BlobStorage) ListProviderVersions(ctx context.Context, namespace, name string) (*core.ProviderVersions, error) {
	providers, err := s.listProviderVersions(ctx, internalProviderType, &core.Provider{Namespace: namespace, Name: name})
	if err != nil {
		return nil, err
	}

	collection := NewCollection()
	for _, p := range providers {
		collection.Add(p)
	}
	return collection.List(), nil
}

func (s *BlobStorage) ListMirroredProviders(ctx context.Context, provider *core.Provider) ([]*core.Provider, error) {
	return s.listProviderVersions(ctx, mirrorProviderType, provider)
}

func (s *BlobStorage) UploadProviderReleaseFiles(ctx context.Context, namespace, name, filename string, file io.Reader) error {
	if namespace == "" {
		return fmt.Errorf("namespace argument is empty")
	}

	if name == "" {
		return fmt.Errorf("name argument is empty")
	}

	if filename == "" {
		return fmt.Errorf("name argument is empty")
	}

	prefix := providerStoragePrefix(s.bucketPrefix, internalProviderType, "", namespace, name)
	key := filepath.Join(prefix, filename)
	return s.upload(ctx, key, file, false)
}

func (s *BlobStorage) UploadMirroredFile(ctx context.Context, provider *core.Provider, fileName string, reader io.Reader) error {
	prefix := providerStoragePrefix(s.bucketPrefix, mirrorProviderType, provider.Hostname, provider.Namespace, provider.Name)

	key := filepath.Join(prefix, fileName)
	return s.upload(ctx, key, reader, true)
}

func (s *BlobStorage) signingKeys(ctx context.Context, pt providerType, hostname, namespace string) (*core.SigningKeys, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace argument is empty")
	}
	key := signingKeysPath(s.bucketPrefix, pt, hostname, namespace)
	exists, err := s.objectExists(ctx, key)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, core.ErrObjectNotFound
	}
	signingKeysRaw, err := s.download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to download signing_keys for namespace %s: %w", namespace, err)
	}

	return unmarshalSigningKeys(signingKeysRaw)
}

// SigningKeys downloads the JSON placed in the namespace in GCS and unmarshals it into a core.SigningKeys
func (s *BlobStorage) SigningKeys(ctx context.Context, namespace string) (*core.SigningKeys, error) {
	return s.signingKeys(ctx, internalProviderType, "", namespace)
}

func (s *BlobStorage) MirroredSigningKeys(ctx context.Context, hostname, namespace string) (*core.SigningKeys, error) {
	return s.signingKeys(ctx, mirrorProviderType, hostname, namespace)
}

func (s *BlobStorage) uploadSigningKeys(ctx context.Context, pt providerType, hostname, namespace string, signingKeys *core.SigningKeys) error {
	b, err := json.Marshal(signingKeys)
	if err != nil {
		return err
	}
	key := signingKeysPath(s.bucketPrefix, pt, hostname, namespace)
	return s.upload(ctx, key, bytes.NewReader(b), true)
}

func (s *BlobStorage) UploadMirroredSigningKeys(ctx context.Context, hostname, namespace string, signingKeys *core.SigningKeys) error {
	return s.uploadSigningKeys(ctx, mirrorProviderType, hostname, namespace, signingKeys)
}

func (s *BlobStorage) MirroredSha256Sum(ctx context.Context, provider *core.Provider) (*core.Sha256Sums, error) {
	prefix := providerStoragePrefix(s.bucketPrefix, mirrorProviderType, provider.Hostname, provider.Namespace, provider.Name)
	key := filepath.Join(prefix, provider.ShasumFileName())
	shaSumBytes, err := s.download(ctx, key)
	if err != nil {
		return nil, errors.New("failed to download SHA256SUMS")
	}
	return core.NewSha256Sums(provider.ShasumFileName(), bytes.NewReader(shaSumBytes))
}

func (s *BlobStorage) upload(ctx context.Context, key string, reader io.Reader, overwrite bool) error {
	if !overwrite {
		exists, err := s.objectExists(ctx, key)
		if err != nil {
			return err
		} else if exists {
			return fmt.Errorf("failed to upload key %s: %w", key, core.ErrObjectAlreadyExists)
		}
	}

	return nil
}

func (s *BlobStorage) download(ctx context.Context, key string) ([]byte, error) {
	return []byte{}, nil
}

// presignedURL generates object signed URL with GET method.
func (s *BlobStorage) presignedURL(ctx context.Context, object string) (string, error) {
	var url string
	return url, nil
}

func (s *BlobStorage) objectExists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

// BlobStorageOption provides additional options for the BlobStorage.
type BlobStorageOption func(*BlobStorage)

// WithBlobStorageBucketPrefix configures the s3 storage to work under a given prefix.
func WithBlobStorageBucketPrefix(prefix string) BlobStorageOption {
	return func(s *BlobStorage) {
		s.bucketPrefix = prefix
	}
}

// WithGCSServiceAccount configures Application Default Credentials (ADC) service account email.
func WithAzureServicePrincipal(sa string) BlobStorageOption {
	return func(s *BlobStorage) {
		s.serviceAccount = sa
	}
}

// WithGCSSignedUrlExpiry configures the duration until the signed url expires
func WithBlobStorageSignedUrlExpiry(t time.Duration) BlobStorageOption {
	return func(s *BlobStorage) {
		s.signedURLExpiry = t
	}
}

// WithGCSArchiveFormat configures the module archive format (zip, tar, tgz, etc.)
func WithBlobStorageArchiveFormat(archiveFormat string) BlobStorageOption {
	return func(s *BlobStorage) {
		s.moduleArchiveFormat = archiveFormat
	}
}

func NewBlobStorage(accountName string, containerName string, options ...BlobStorageOption) (*BlobStorage, error) {
	//ctx := context.Background()

	// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to determine Azure Default Credentials: %w", err)
	}
	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName)
	client, err := container.NewClient(containerURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}
	fmt.Println(client.URL())
	s := &BlobStorage{
		sc:            client,
		accountName:   accountName,
		containerName: containerName,
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}
