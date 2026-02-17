package nix

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// PyPIPackageInfo holds resolved metadata for a single PyPI package.
type PyPIPackageInfo struct {
	Name     string // package name as typed by the user
	Version  string // latest version resolved from PyPI (empty if lookup failed)
	HashExpr string // Nix expression: quoted "sha256-..." or bare pkgs.lib.fakeHash
	Resolved bool   // true when version + hash were fetched successfully
}

// ResolvePyPIPackage fetches the latest version and SHA-256 hash for a package
// from the PyPI JSON API. On any failure it returns a stub with pkgs.lib.fakeHash
// so the generated flake is still valid Nix (it just needs the user to fill in
// the version and run 'nix develop' once to get the real hash).
func ResolvePyPIPackage(name string) PyPIPackageInfo {
	info, err := fetchPyPIInfo(name)
	if err != nil {
		return PyPIPackageInfo{
			Name:     name,
			HashExpr: "pkgs.lib.fakeHash",
			Resolved: false,
		}
	}
	return info
}

func fetchPyPIInfo(name string) (PyPIPackageInfo, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", name)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return PyPIPackageInfo{}, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PyPIPackageInfo{}, fmt.Errorf("pypi returned %d for %q", resp.StatusCode, name)
	}

	var payload struct {
		Info struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"info"`
		URLs []struct {
			PackageType string `json:"packagetype"`
			Digests     struct {
				SHA256 string `json:"sha256"`
			} `json:"digests"`
		} `json:"urls"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return PyPIPackageInfo{}, fmt.Errorf("json: %w", err)
	}

	// Prefer the source distribution (tar.gz) — that's what fetchPypi defaults to.
	for _, u := range payload.URLs {
		if u.PackageType == "sdist" {
			sri, err := hexSHA256ToSRI(u.Digests.SHA256)
			if err != nil {
				return PyPIPackageInfo{}, err
			}
			return PyPIPackageInfo{
				Name:     payload.Info.Name,
				Version:  payload.Info.Version,
				HashExpr: fmt.Sprintf(`"sha256-%s"`, sri),
				Resolved: true,
			}, nil
		}
	}

	return PyPIPackageInfo{}, fmt.Errorf("no sdist found for %q", name)
}

// hexSHA256ToSRI converts a hex-encoded SHA-256 digest to the base64 portion
// of a Nix SRI hash (the part after "sha256-").
func hexSHA256ToSRI(hexStr string) (string, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex decode: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
