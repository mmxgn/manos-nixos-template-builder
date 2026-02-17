package nix

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// PyPIPackageInfo holds resolved metadata for a single PyPI package.
type PyPIPackageInfo struct {
	Name        string // package name as typed by the user
	Version     string // latest version resolved from PyPI (empty if lookup failed)
	HashExpr    string // Nix expression: quoted "sha256-..." or bare pkgs.lib.fakeHash
	BuildSystem string // Nix attr for build-system, e.g. "setuptools", "hatchling"
	Resolved    bool   // true when version + hash were fetched successfully
}

// ResolvePyPIPackage fetches the latest version, SHA-256 hash, and build backend
// for a package from the PyPI JSON API. On any failure it returns a stub with
// pkgs.lib.fakeHash and setuptools so the generated flake is still valid Nix.
func ResolvePyPIPackage(name string) PyPIPackageInfo {
	info, err := fetchPyPIInfo(name)
	if err != nil {
		return PyPIPackageInfo{
			Name:        name,
			HashExpr:    "pkgs.lib.fakeHash",
			BuildSystem: "setuptools",
			Resolved:    false,
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
			URL         string `json:"url"`
			Digests     struct {
				SHA256 string `json:"sha256"`
			} `json:"digests"`
		} `json:"urls"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return PyPIPackageInfo{}, fmt.Errorf("json: %w", err)
	}

	// Find the source distribution (sdist).
	for _, u := range payload.URLs {
		if u.PackageType != "sdist" {
			continue
		}

		sri, err := hexSHA256ToSRI(u.Digests.SHA256)
		if err != nil {
			return PyPIPackageInfo{}, err
		}

		// Detect the build backend from pyproject.toml inside the sdist.
		buildSystem := detectBuildSystem(u.URL)

		return PyPIPackageInfo{
			Name:        payload.Info.Name,
			Version:     payload.Info.Version,
			HashExpr:    fmt.Sprintf(`"sha256-%s"`, sri),
			BuildSystem: buildSystem,
			Resolved:    true,
		}, nil
	}

	return PyPIPackageInfo{}, fmt.Errorf("no sdist found for %q", name)
}

// detectBuildSystem downloads the sdist, extracts pyproject.toml, and returns
// the corresponding nixpkgs Python package attr for the build backend.
// Falls back to "setuptools" on any error.
func detectBuildSystem(sdistURL string) string {
	backend, err := fetchBuildBackend(sdistURL)
	if err != nil {
		return "setuptools"
	}
	return backendToNixAttr(backend)
}

func fetchBuildBackend(sdistURL string) (string, error) {
	resp, err := http.Get(sdistURL) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(hdr.Name, "/pyproject.toml") || hdr.Name == "pyproject.toml" {
			content, err := io.ReadAll(tr)
			if err != nil {
				return "", err
			}
			return parseBuildBackend(string(content)), nil
		}
	}
	return "", fmt.Errorf("pyproject.toml not found in sdist")
}

var buildBackendRe = regexp.MustCompile(`build-backend\s*=\s*['"]([^'"]+)['"]`)

func parseBuildBackend(toml string) string {
	m := buildBackendRe.FindStringSubmatch(toml)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// backendToNixAttr maps a PEP 517 build-backend string to its nixpkgs
// Python package attribute name.
func backendToNixAttr(backend string) string {
	switch {
	case strings.Contains(backend, "hatchling"):
		return "hatchling"
	case strings.Contains(backend, "flit_core"), strings.Contains(backend, "flit-core"):
		return "flit-core"
	case strings.Contains(backend, "poetry"):
		return "poetry-core"
	case strings.Contains(backend, "maturin"):
		return "maturin"
	case strings.Contains(backend, "pdm"):
		return "pdm-backend"
	case strings.Contains(backend, "scikit_build"), strings.Contains(backend, "scikit-build"):
		return "scikit-build-core"
	default:
		return "setuptools"
	}
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
