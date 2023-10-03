package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/TierMobility/boring-registry/pkg/core"
	"github.com/go-kit/kit/log"
	"github.com/google/go-github/github"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
)

// GithubStorage is a Storage implementation backed by Github.
type GithubStorage struct {
	client *github.Client
}

func NewGithubStorage(ctx context.Context) (Storage, error) {
	g := GithubStorage{}
	g.client = github.NewClient(nil)

	return &g, nil
}

func (g *GithubStorage) GetProvider(ctx context.Context, namespace, name, version, os, arch string) (core.Provider, error) {
	p, err := g.getProvider(ctx, internalProviderType, &core.Provider{
		Namespace: namespace,
		Name:      name,
		Version:   version,
		OS:        os,
		Arch:      arch,
	})
	return *p, err
}

// GetProvider implements provider.Storage
func (g *GithubStorage) getProvider(ctx context.Context, pt providerType, provider *core.Provider) (*core.Provider, error) {
	tag := fmt.Sprintf("v%s", provider.Version)
	repo := fmt.Sprintf("terraform-provider-%s", provider.Name)
	releaseDownloadPath := fmt.Sprintf("%s/releases/download/v%s", repo, provider.Version)

	archivePath := path.Clean(path.Join(provider.Hostname, provider.Namespace, releaseDownloadPath, provider.ArchiveFileName()))
	//shasumPath := path.Clean(path.Join(provider.Hostname, provider.Namespace, releaseDownloadPath, provider.ShasumFileName()))
	//shasumSigPath := path.Clean(path.Join(provider.Hostname, provider.Namespace, releaseDownloadPath, provider.ShasumSignatureFileName()))

	//fmt.Println(archivePath, shasumPath, shasumSigPath)

	//https://github.com/ably/terraform-provider-ably/releases/download/v0.5.0/terraform-provider-ably_0.5.0_darwin_amd64.zip
	//https://github.com/abbeylabs/terraform-provider-abbey/releases/download/v0.2.6/terraform-provider-abbey_0.2.6_darwin_amd64.zip
	release, _, err := g.client.Repositories.GetReleaseByTag(context.Background(), provider.Namespace, repo, tag)
	if err != nil {
		return nil, err
	} else if release == nil {
		return nil, &core.ProviderError{
			Reason:     "failed to locate provider",
			Provider:   provider,
			StatusCode: http.StatusNotFound,
		}
	}

	for _, asset := range release.Assets {
		switch *asset.Name {
		case provider.ArchiveFileName():
			provider.DownloadURL = asset.GetBrowserDownloadURL()
		case provider.ShasumFileName():
			provider.SHASumsURL = asset.GetBrowserDownloadURL()

		case provider.ShasumSignatureFileName():
			provider.SHASumsSignatureURL = asset.GetBrowserDownloadURL()
		}
	}

	//fmt.Println(release.Assets, err)

	shasumBytes, err := g.download(ctx, provider.SHASumsURL)
	if err != nil {
		return nil, err
	}
	provider.Shasum, err = readSHASums(bytes.NewReader(shasumBytes), path.Base(archivePath))
	if err != nil {
		return nil, err
	}
	shasumSigBytes, err := g.download(ctx, fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/%s/%s/download/%s/%s",
		provider.Namespace, provider.Name, provider.Version, provider.OS, provider.Arch))
	if err != nil {
		return nil, err
	}

	// https://registry.terraform.io/v1/providers/abbeylabs/abbey/0.2.6/download/darwin/arm64
	var prp core.ProviderRegistryProtocol
	if err := json.Unmarshal(shasumSigBytes, &prp); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	provider.Filename = path.Base(archivePath)
	provider.SigningKeys = prp.SigningKeys

	return provider, nil
}

func (g *GithubStorage) download(ctx context.Context, url string) ([]byte, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *GithubStorage) ListProviderVersions(ctx context.Context, namespace, name string) (*core.ProviderVersions, error) {
	providers, err := g.listProviderVersions(ctx, internalProviderType, &core.Provider{Namespace: namespace, Name: name})
	if err != nil {
		return nil, err
	}

	collection := NewCollection()
	for _, p := range providers {
		collection.Add(p)
	}
	return collection.List(), nil
}

func (g *GithubStorage) listProviderVersions(ctx context.Context, pt providerType, provider *core.Provider) ([]*core.Provider, error) {
	opt := &github.ListOptions{}
	repo := fmt.Sprintf("terraform-provider-%s", provider.Name)
	releases, _, err := g.client.Repositories.ListReleases(context.Background(), provider.Namespace, repo, opt)

	var providers []*core.Provider
	if err != nil {
		fmt.Println(err)
	}
	for _, r := range releases {
		// v0.1.0-rc.2
		fmt.Println(r.GetName())
		for _, a := range r.Assets {
			// https://github.com/abbeylabs/terraform-provider-abbey/releases/download/v0.1.0-rc.2/terraform-provider-abbey_0.1.0-rc.2_darwin_amd64.zip
			// https://github.com/abbeylabs/terraform-provider-abbey/releases/download/v0.1.0-rc.2/terraform-provider-abbey_0.1.0-rc.2_manifest.json
			// https://github.com/abbeylabs/terraform-provider-abbey/releases/download/v0.1.0-rc.2/terraform-provider-abbey_0.1.0-rc.2_SHA256SUMS
			// https://github.com/abbeylabs/terraform-provider-abbey/releases/download/v0.1.0-rc.2/terraform-provider-abbey_0.1.0-rc.2_SHA256SUMS.sig
			fmt.Println(a.GetBrowserDownloadURL())
			u, err := url.Parse(a.GetBrowserDownloadURL())
			if err != nil {
				fmt.Println(err)
			}
			p, err := core.NewProviderFromArchive(filepath.Base(u.Path))
			if err != nil {
				continue
			}
			p.Hostname = provider.Hostname
			p.Namespace = provider.Namespace
			p.DownloadURL = a.GetBrowserDownloadURL()
			if provider.Version != "" && provider.Version != p.Version {
				// The provider version doesn't match the requested version
				continue
			}
			providers = append(providers, &p)
		}
	}
	if len(providers) == 0 {
		return nil, noMatchingProviderFound(provider)
	}

	return providers, nil
}

func (g *GithubStorage) UploadProviderReleaseFiles(ctx context.Context, namespace, name, filename string, file io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) SigningKeys(ctx context.Context, namespace string) (*core.SigningKeys, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) MigrateProviders(ctx context.Context, logger log.Logger, dryRun bool) error {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) GetModule(ctx context.Context, namespace, name, provider, version string) (core.Module, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) ListModuleVersions(ctx context.Context, namespace, name, provider string) ([]core.Module, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) UploadModule(ctx context.Context, namespace, name, provider, version string, body io.Reader) (core.Module, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) MigrateModules(ctx context.Context, logger log.Logger, dryRun bool) error {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) ListMirroredProviders(ctx context.Context, provider *core.Provider) ([]*core.Provider, error) {
	return g.listProviderVersions(ctx, mirrorProviderType, provider)
}

func (g *GithubStorage) GetMirroredProvider(ctx context.Context, provider *core.Provider) (*core.Provider, error) {
	//return g.getProvider(ctx, mirrorProviderType, provider)
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) UploadMirroredFile(ctx context.Context, provider *core.Provider, fileName string, reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) MirroredSigningKeys(ctx context.Context, hostname, namespace string) (*core.SigningKeys, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) UploadMirroredSigningKeys(ctx context.Context, hostname, namespace string, signingKeys *core.SigningKeys) error {
	//TODO implement me
	panic("implement me")
}

func (g *GithubStorage) MirroredSha256Sum(ctx context.Context, provider *core.Provider) (*core.Sha256Sums, error) {
	//TODO implement me
	panic("implement me")
}
