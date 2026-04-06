package runtimes

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/vyper-tooling/pi/internal/config"
)

// BaseDir is the default root for provisioned runtimes.
var BaseDir = filepath.Join(homeDir(), ".pi", "runtimes")

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return os.TempDir()
}

// KnownRuntimes are the runtimes PI can provision.
var KnownRuntimes = map[string]bool{
	"python": true,
	"node":   true,
	"go":     true,
	"rust":   true,
}

// knownRuntimeList returns a sorted, comma-separated list of known runtimes.
func knownRuntimeList() string {
	names := make([]string, 0, len(KnownRuntimes))
	for name := range KnownRuntimes {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// CmdRunner abstracts command execution so provisioning logic can be tested
// without real binaries or network access. The default implementation
// delegates to os/exec.
type CmdRunner interface {
	// Run executes a command with the given stdout and stderr writers.
	// Returns the exit error (if any).
	Run(bin string, args []string, stdout, stderr io.Writer) error

	// Output executes a command and returns its stdout as a string.
	Output(bin string, args []string, stderr io.Writer) (string, error)
}

// execCmdRunner is the default CmdRunner that delegates to os/exec.
type execCmdRunner struct{}

func (r *execCmdRunner) Run(bin string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (r *execCmdRunner) Output(bin string, args []string, stderr io.Writer) (string, error) {
	cmd := exec.Command(bin, args...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// Provisioner manages runtime provisioning using a pluggable backend.
type Provisioner struct {
	Mode    config.ProvisionMode
	Manager config.RuntimeManager
	BaseDir string
	Stderr  io.Writer

	// PromptFunc is called when Mode is "ask". It should return true if the user
	// confirms provisioning. If nil and mode is "ask", provisioning is skipped
	// (non-interactive / CI mode).
	PromptFunc func(msg string) bool

	// LookPath can be overridden for testing. Defaults to exec.LookPath.
	LookPath func(string) (string, error)

	// Runner executes external commands. If nil, os/exec is used.
	Runner CmdRunner
}

// NewProvisioner creates a Provisioner from project config.
func NewProvisioner(cfg *config.ProjectConfig, stderr io.Writer) *Provisioner {
	return &Provisioner{
		Mode:    cfg.EffectiveProvisionMode(),
		Manager: cfg.EffectiveRuntimeManager(),
		BaseDir: BaseDir,
		Stderr:  stderr,
	}
}

// ProvisionResult holds the result of a provisioning attempt.
type ProvisionResult struct {
	Provisioned bool
	BinDir      string // directory to prepend to PATH for step execution
	Version     string
}

// Provision ensures a runtime is available at the requested version.
// Returns the bin directory to prepend to PATH for step execution.
// If the runtime is already provisioned or provisioning is disabled, returns
// appropriate zero values.
func (p *Provisioner) Provision(runtimeName, minVersion string) (*ProvisionResult, error) {
	if !KnownRuntimes[runtimeName] {
		return nil, fmt.Errorf("unknown runtime %q: PI can only provision %s", runtimeName, knownRuntimeList())
	}

	if p.Mode == config.ProvisionNever {
		return &ProvisionResult{}, nil
	}

	// Check if already provisioned locally
	binDir := p.binDir(runtimeName, minVersion)
	if p.isProvisioned(runtimeName, binDir) {
		return &ProvisionResult{
			Provisioned: true,
			BinDir:      binDir,
		}, nil
	}

	if p.Mode == config.ProvisionAsk {
		if p.PromptFunc == nil {
			return &ProvisionResult{}, nil
		}
		msg := fmt.Sprintf("Runtime %q is not installed. Provision it into %s?", runtimeName, p.BaseDir)
		if minVersion != "" {
			msg = fmt.Sprintf("Runtime %s >= %s is not installed. Provision it into %s?", runtimeName, minVersion, p.BaseDir)
		}
		if !p.PromptFunc(msg) {
			return &ProvisionResult{}, nil
		}
	}

	version := minVersion
	if version == "" {
		version = defaultVersion(runtimeName)
	}

	var err error
	switch p.Manager {
	case config.RuntimeManagerMise, "":
		err = p.provisionWithMise(runtimeName, version)
	case config.RuntimeManagerDirect:
		err = p.provisionDirect(runtimeName, version)
	default:
		return nil, fmt.Errorf("unknown runtime manager %q", p.Manager)
	}

	if err != nil {
		return nil, fmt.Errorf("provisioning %s %s: %w", runtimeName, version, err)
	}

	binDir = p.binDir(runtimeName, version)
	fmt.Fprintf(p.stderr(), "[provisioned] %s %s → %s\n", runtimeName, version, binDir)

	return &ProvisionResult{
		Provisioned: true,
		BinDir:      binDir,
		Version:     version,
	}, nil
}

// BinDirFor returns the bin directory for an already-provisioned runtime,
// or empty string if not provisioned.
func (p *Provisioner) BinDirFor(runtimeName, version string) string {
	if version == "" {
		version = defaultVersion(runtimeName)
	}
	binDir := p.binDir(runtimeName, version)
	if p.isProvisioned(runtimeName, binDir) {
		return binDir
	}
	return ""
}

func (p *Provisioner) binDir(runtimeName, version string) string {
	if version == "" {
		version = defaultVersion(runtimeName)
	}
	return filepath.Join(p.BaseDir, runtimeName, version, "bin")
}

func (p *Provisioner) isProvisioned(runtimeName, binDir string) bool {
	cmdName := runtimeBinary(runtimeName)
	binPath := filepath.Join(binDir, cmdName)
	info, err := os.Stat(binPath)
	return err == nil && !info.IsDir()
}

func defaultVersion(runtimeName string) string {
	switch runtimeName {
	case "python":
		return "3.13"
	case "node":
		return "20"
	case "go":
		return "1.23"
	case "rust":
		return "stable"
	default:
		return "latest"
	}
}

func runtimeBinary(name string) string {
	switch name {
	case "python":
		return "python3"
	case "node":
		return "node"
	case "rust":
		return "rustc"
	default:
		return name
	}
}

func (p *Provisioner) stderr() io.Writer {
	if p.Stderr != nil {
		return p.Stderr
	}
	return os.Stderr
}

func (p *Provisioner) runner() CmdRunner {
	if p.Runner != nil {
		return p.Runner
	}
	return &execCmdRunner{}
}

// provisionWithMise uses mise to install a runtime version into the PI runtimes directory.
func (p *Provisioner) provisionWithMise(runtimeName, version string) error {
	lookPath := p.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	misePath, err := lookPath("mise")
	if err != nil {
		return p.provisionDirect(runtimeName, version)
	}

	installDir := filepath.Join(p.BaseDir, runtimeName, version)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("creating runtime directory: %w", err)
	}

	runner := p.runner()
	spec := fmt.Sprintf("%s@%s", runtimeName, version)

	if err := runner.Run(misePath, []string{"install", spec}, io.Discard, p.stderr()); err != nil {
		return fmt.Errorf("mise install %s: %w", spec, err)
	}

	miseWherePath, err := runner.Output(misePath, []string{"where", spec}, io.Discard)
	if err != nil {
		return fmt.Errorf("mise where %s: %w", spec, err)
	}

	miseBinDir := filepath.Join(miseWherePath, "bin")

	binDir := filepath.Join(installDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	entries, err := os.ReadDir(miseBinDir)
	if err != nil {
		return fmt.Errorf("reading mise bin directory %s: %w", miseBinDir, err)
	}

	for _, entry := range entries {
		src := filepath.Join(miseBinDir, entry.Name())
		dst := filepath.Join(binDir, entry.Name())
		os.Remove(dst)
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("symlinking %s → %s: %w", src, dst, err)
		}
	}

	return nil
}

// provisionDirect downloads a runtime binary directly from official CDNs.
func (p *Provisioner) provisionDirect(runtimeName, version string) error {
	switch runtimeName {
	case "node":
		return p.provisionNodeDirect(version)
	case "python":
		return p.provisionPythonDirect(version)
	case "go", "rust":
		return fmt.Errorf("direct provisioning for %q is not supported — install mise first: curl https://mise.run | sh", runtimeName)
	default:
		return fmt.Errorf("direct provisioning not supported for %q — install mise first: curl https://mise.run | sh", runtimeName)
	}
}

func (p *Provisioner) provisionNodeDirect(version string) error {
	installDir := filepath.Join(p.BaseDir, "node", version)
	binDir := filepath.Join(installDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var platform, arch string
	switch goos {
	case "darwin":
		platform = "darwin"
	case "linux":
		platform = "linux"
	default:
		return fmt.Errorf("direct node provisioning not supported on %s", goos)
	}
	switch goarch {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "arm64"
	default:
		return fmt.Errorf("direct node provisioning not supported on %s", goarch)
	}

	versionTag := version
	if !strings.HasPrefix(version, "v") {
		versionTag = "v" + version
	}

	url := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-%s-%s.tar.gz", versionTag, versionTag, platform, arch)
	extractDir := filepath.Join(installDir, "extract")

	script := fmt.Sprintf(`set -e
mkdir -p %q
curl -fsSL %q | tar xz -C %q --strip-components=1
if [ -d %q ]; then
  rm -rf %q
  mv %q %q
fi
`, extractDir, url, extractDir,
		filepath.Join(extractDir, "bin"),
		binDir,
		filepath.Join(extractDir, "bin"), binDir,
	)

	runner := p.runner()
	if err := runner.Run("bash", []string{"-c", script}, io.Discard, p.stderr()); err != nil {
		os.RemoveAll(extractDir)
		return fmt.Errorf("downloading node %s from %s: %w", version, url, err)
	}

	libSrc := filepath.Join(extractDir, "lib")
	libDst := filepath.Join(installDir, "lib")
	if info, err := os.Stat(libSrc); err == nil && info.IsDir() {
		os.RemoveAll(libDst)
		runner.Run("cp", []string{"-a", libSrc, libDst}, io.Discard, io.Discard)
	}

	os.RemoveAll(extractDir)
	return nil
}

func (p *Provisioner) provisionPythonDirect(version string) error {
	installDir := filepath.Join(p.BaseDir, "python", version)
	binDir := filepath.Join(installDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var platform, arch string
	switch goos {
	case "darwin":
		platform = "apple-darwin"
	case "linux":
		platform = "unknown-linux-gnu"
	default:
		return fmt.Errorf("direct python provisioning not supported on %s", goos)
	}
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return fmt.Errorf("direct python provisioning not supported on %s", goarch)
	}

	extractDir := filepath.Join(installDir, "extract")

	script := fmt.Sprintf(`set -e
mkdir -p %q

RELEASE_TAG=$(curl -fsSL "https://api.github.com/repos/astral-sh/python-build-standalone/releases" | \
  python3 -c "
import sys, json
releases = json.load(sys.stdin)
version = '%s'
for release in releases:
    for asset in release.get('assets', []):
        name = asset['name']
        if 'cpython-' + version in name and '%s' in name and '%s' in name and 'install_only' in name and name.endswith('.tar.gz'):
            print(asset['browser_download_url'])
            sys.exit(0)
sys.exit(1)
" 2>/dev/null) || {
  echo "Could not find Python %s for %s-%s" >&2
  exit 1
}

curl -fsSL "$RELEASE_TAG" | tar xz -C %q --strip-components=0

if [ -d %q/python/bin ]; then
  rm -rf %q
  mv %q/python/bin %q
  if [ -d %q/python/lib ]; then
    rm -rf %q
    mv %q/python/lib %q
  fi
fi
rm -rf %q
`, extractDir, version, arch, platform, version, arch, platform,
		extractDir,
		extractDir, binDir, extractDir, binDir,
		extractDir, filepath.Join(installDir, "lib"), extractDir, filepath.Join(installDir, "lib"),
		extractDir,
	)

	if err := p.runner().Run("bash", []string{"-c", script}, io.Discard, p.stderr()); err != nil {
		os.RemoveAll(extractDir)
		return fmt.Errorf("downloading python %s: %w — try installing mise first: curl https://mise.run | sh", version, err)
	}

	return nil
}

// PrependToPath returns a modified PATH with binDir prepended.
func PrependToPath(binDir string) string {
	return binDir + string(os.PathListSeparator) + os.Getenv("PATH")
}
