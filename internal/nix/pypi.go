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
	Name        string   // package name as typed by the user
	Version     string   // latest version resolved from PyPI (empty if lookup failed)
	HashExpr    string   // Nix expression: quoted "sha256-..." or bare pkgs.lib.fakeHash
	BuildDeps   []string // nixpkgs attrs for build-system, e.g. ["hatchling", "hatch-vcs"]
	RuntimeDeps []string // nixpkgs attrs for propagatedBuildInputs (from requires_dist)
	Resolved    bool     // true when version + hash were fetched successfully
}

// ResolvePyPIPackage fetches the latest version, SHA-256 hash, and build
// dependencies for a package from the PyPI JSON API. On any failure it returns
// a stub with pkgs.lib.fakeHash and setuptools.
func ResolvePyPIPackage(name string) PyPIPackageInfo {
	info, err := fetchPyPIInfo(name)
	if err != nil {
		return PyPIPackageInfo{
			Name:      name,
			HashExpr:  "pkgs.lib.fakeHash",
			BuildDeps: []string{"setuptools"},
			Resolved:  false,
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
			Name         string   `json:"name"`
			Version      string   `json:"version"`
			RequiresDist []string `json:"requires_dist"`
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

	for _, u := range payload.URLs {
		if u.PackageType != "sdist" {
			continue
		}

		sri, err := hexSHA256ToSRI(u.Digests.SHA256)
		if err != nil {
			return PyPIPackageInfo{}, err
		}

		buildDeps := detectBuildDeps(u.URL)
		runtimeDeps := parseRuntimeDeps(payload.Info.RequiresDist)

		return PyPIPackageInfo{
			Name:        payload.Info.Name,
			Version:     payload.Info.Version,
			HashExpr:    fmt.Sprintf(`"sha256-%s"`, sri),
			BuildDeps:   buildDeps,
			RuntimeDeps: runtimeDeps,
			Resolved:    true,
		}, nil
	}

	return PyPIPackageInfo{}, fmt.Errorf("no sdist found for %q", name)
}

// detectBuildDeps downloads the sdist, reads pyproject.toml, and returns the
// list of nixpkgs Python package attrs needed for build-system.
func detectBuildDeps(sdistURL string) []string {
	deps, err := fetchBuildDeps(sdistURL)
	if err != nil || len(deps) == 0 {
		return []string{"setuptools"}
	}
	return deps
}

func fetchBuildDeps(sdistURL string) ([]string, error) {
	resp, err := http.Get(sdistURL) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(hdr.Name, "/pyproject.toml") || hdr.Name == "pyproject.toml" {
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return parseBuildDeps(string(content)), nil
		}
	}
	return nil, fmt.Errorf("pyproject.toml not found in sdist")
}

// buildSystemRe captures the content of the requires = [...] array inside
// the [build-system] section. Handles multi-line arrays.
var buildSystemRe = regexp.MustCompile(`(?s)\[build-system\][^\[]*requires\s*=\s*\[([^\]]+)\]`)

// quotedPkgRe extracts quoted strings (package specifiers) from the requires list.
var quotedPkgRe = regexp.MustCompile(`['"]([^'"]+)['"]`)

// pkgNameRe strips version specifiers like >=1.0, ~=2, etc.
var pkgNameRe = regexp.MustCompile(`^([A-Za-z0-9_.-]+)`)

func parseBuildDeps(toml string) []string {
	m := buildSystemRe.FindStringSubmatch(toml)
	if len(m) < 2 {
		return []string{"setuptools"}
	}

	var deps []string
	for _, match := range quotedPkgRe.FindAllStringSubmatch(m[1], -1) {
		specifier := match[1]
		nm := pkgNameRe.FindString(specifier)
		if nm == "" {
			continue
		}
		nixAttr := pypiNameToNixAttr(nm)
		deps = append(deps, nixAttr)
	}
	if len(deps) == 0 {
		return []string{"setuptools"}
	}
	return deps
}

// pkgSpecRe extracts the bare package name from a PEP 508 dependency specifier,
// e.g. "httpx>=0.27" → "httpx", "pydantic[email]>=2" → "pydantic".
var pkgSpecNameRe = regexp.MustCompile(`^([A-Za-z0-9_.-]+)`)

// parseRuntimeDeps converts a requires_dist list (PEP 508 specifiers) to a
// list of nixpkgs Python package attrs. Entries with environment markers
// ("; sys_platform ==…" etc.) are skipped — they are optional/platform-specific.
func parseRuntimeDeps(requiresDist []string) []string {
	var deps []string
	for _, spec := range requiresDist {
		// Skip conditional/optional dependencies entirely.
		if strings.Contains(spec, ";") {
			continue
		}
		name := pkgSpecNameRe.FindString(spec)
		if name == "" {
			continue
		}
		deps = append(deps, pypiNameToNixAttr(name))
	}
	return deps
}

// pypiNameToNixAttr converts a PyPI package name to its nixpkgs Python attr.
// Most names match after lowercasing and replacing underscores with hyphens.
func pypiNameToNixAttr(name string) string {
	normalized := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	overrides := map[string]string{
		"scikit-build-core": "scikit-build-core",
		"setuptools-scm":    "setuptools-scm",
		"setuptools":        "setuptools",
		"hatchling":         "hatchling",
		"hatch-vcs":         "hatch-vcs",
		"flit-core":         "flit-core",
		"poetry-core":       "poetry-core",
		"maturin":           "maturin",
		"pdm-backend":       "pdm-backend",
		"wheel":             "wheel",
		"cython":            "cython",
		"ninja":             "ninja",
	}
	if attr, ok := overrides[normalized]; ok {
		return attr
	}
	return normalized
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
