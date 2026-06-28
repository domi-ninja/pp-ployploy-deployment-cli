package deploy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"
	"time"
)

type GitMetadata struct {
	SHA        string
	ShortSHA   string
	Dirty      bool
	DiffDigest string
}

func ReadGitMetadata(root string) (GitMetadata, error) {
	sha, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return GitMetadata{}, err
	}
	shortSHA, err := gitOutput(root, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return GitMetadata{}, err
	}

	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return GitMetadata{}, err
	}

	diff, err := gitOutput(root, "diff", "HEAD")
	if err != nil {
		return GitMetadata{}, err
	}
	sum := sha256.Sum256([]byte(status + "\n" + diff))

	return GitMetadata{
		SHA:        sha,
		ShortSHA:   shortSHA,
		Dirty:      strings.TrimSpace(status) != "",
		DiffDigest: hex.EncodeToString(sum[:]),
	}, nil
}

func ReleaseID(now time.Time, shortSHA string) string {
	return now.UTC().Format("20060102T150405Z") + "-" + shortSHA
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
