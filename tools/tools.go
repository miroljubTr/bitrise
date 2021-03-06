package tools

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/errorutil"
)

// UnameGOOS ...
func UnameGOOS() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin", nil
	case "linux":
		return "Linux", nil
	}
	return "", fmt.Errorf("Unsupported platform (%s)", runtime.GOOS)
}

// UnameGOARCH ...
func UnameGOARCH() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", nil
	}
	return "", fmt.Errorf("Unsupported architecture (%s)", runtime.GOARCH)
}

// InstallToolFromGitHub ...
func InstallToolFromGitHub(toolname, githubUser, toolVersion string) error {
	unameGOOS, err := UnameGOOS()
	if err != nil {
		return fmt.Errorf("Failed to determine OS: %s", err)
	}
	unameGOARCH, err := UnameGOARCH()
	if err != nil {
		return fmt.Errorf("Failed to determine ARCH: %s", err)
	}
	downloadURL := "https://github.com/" + githubUser + "/" + toolname + "/releases/download/" + toolVersion + "/" + toolname + "-" + unameGOOS + "-" + unameGOARCH

	return InstallFromURL(toolname, downloadURL)
}

// DownloadFile ...
func DownloadFile(downloadURL, targetDirPath string) error {
	outFile, err := os.Create(targetDirPath)
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Warnf("Failed to close (%s)", targetDirPath)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create (%s), error: %s", targetDirPath, err)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body", downloadURL)
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download from (%s), error: %s", downloadURL, err)
	}

	return nil
}

// InstallFromURL ...
func InstallFromURL(toolBinName, downloadURL string) error {
	if len(toolBinName) < 1 {
		return fmt.Errorf("No Tool (bin) Name provided! URL was: %s", downloadURL)
	}

	bitriseToolsDirPath := configs.GetBitriseToolsDirPath()
	destinationPth := filepath.Join(bitriseToolsDirPath, toolBinName)

	if err := DownloadFile(downloadURL, destinationPth); err != nil {
		return fmt.Errorf("Failed to download, error: %s", err)
	}

	if err := os.Chmod(destinationPth, 0755); err != nil {
		return fmt.Errorf("Failed to make file (%s) executable, error: %s", destinationPth, err)
	}

	return nil
}

// ------------------
// --- Stepman

// StepmanSetup ...
func StepmanSetup(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "setup", "--collection", collection}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanActivate ...
func StepmanActivate(collection, stepID, stepVersion, dir, ymlPth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "activate", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--path", dir, "--copyyml", ymlPth}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanUpdate ...
func StepmanUpdate(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "update", "--collection", collection}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanRawStepLibStepInfo ...
func StepmanRawStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanRawLocalStepInfo ...
func StepmanRawLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanJSONStepLibStepInfo ...
func StepmanJSONStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanJSONLocalStepInfo ...
func StepmanJSONLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanRawStepList ...
func StepmanRawStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanJSONStepList ...
func StepmanJSONStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

//
// Share

// StepmanShare ...
func StepmanShare() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareAudit ...
func StepmanShareAudit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "audit", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareCreate ...
func StepmanShareCreate(tag, git, stepID string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "create", "--tag", tag, "--git", git, "--stepid", stepID, "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareFinish ...
func StepmanShareFinish() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "finish", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareStart ...
func StepmanShareStart(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "start", "--collection", collection, "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// EnvmanInit ...
func EnvmanInit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "init"}
	return cmdex.RunCommand("envman", args...)
}

// EnvmanInitAtPath ...
func EnvmanInitAtPath(envstorePth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "init", "--clear"}
	return cmdex.RunCommand("envman", args...)
}

// EnvmanAdd ...
func EnvmanAdd(envstorePth, key, value string, expand, skipIfEmpty bool) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "add", "--key", key, "--append"}
	if !expand {
		args = append(args, "--no-expand")
	}
	if skipIfEmpty {
		args = append(args, "--skip-if-empty")
	}

	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// EnvmanClear ...
func EnvmanClear(envstorePth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "clear"}
	out, err := cmdex.NewCommand("envman", args...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		errorMsg := err.Error()
		if errorutil.IsExitStatusError(err) && out != "" {
			errorMsg = out
		}
		return fmt.Errorf("failed to clear envstore (%s), error: %s", envstorePth, errorMsg)
	}
	return nil
}

// EnvmanRun ...
func EnvmanRun(envstorePth, workDirPth string, cmd []string) (int, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "run"}
	args = append(args, cmd...)

	return cmdex.RunCommandInDirAndReturnExitCode(workDirPth, "envman", args...)
}

// EnvmanJSONPrint ...
func EnvmanJSONPrint(envstorePth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "print", "--format", "json", "--expand"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "envman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}
