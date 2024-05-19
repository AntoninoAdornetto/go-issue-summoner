package scm

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	GH = "github"
	GL = "gitlab"
	BB = "bitbucket"
)

// @TODO can GlobalUserName and RepoName functions be deleted?
// We are now using the device flow and the mentioned functions could be useless since
// we are creating an access token for the user after they authorize the application.

type GitConfig struct {
	UserName       string
	RepositoryName string
	Token          string
	Scm            string // GitHub || GitLab || BitBucket ...
}

// @TODO Change the name of scm.Issue struct. It conflicts with the struct in issue.go
type Issue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// GitConfigManager interface allows us to have different adapters for each
// source code management system that we would like to use. We can have different
// implementations for GitHub, GitLab, BitBucket and so on.
// Authorize creates an access token with scopes that will allow us to read/write issues
// ReadToken checks if there is an access token in ~/.config/issue-summoner/config.json
type GitConfigManager interface {
	Authorize() error
	Report(issues []Issue) <-chan int64
}

func NewGitManager(scm string) GitConfigManager {
	switch scm {
	default:
		return &GitHubManager{}
	}
}

type ScmTokenConfig struct {
	AccessToken string
}

type IssueSummonerConfig = map[string]ScmTokenConfig

// WriteToken accepts an access token and the source code management platform
// (GitHub, GitLab etc...) and will write the token to a configuration file.
// This will be used to authorize future requests for reporting issues.
func WriteToken(token string, scm string) error {
	config := make(map[string]ScmTokenConfig)

	usr, err := user.Current()
	if err != nil {
		return err
	}

	home := usr.HomeDir
	path := filepath.Join(home, ".config", "issue-summoner")

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	configFile := filepath.Join(path, "config.json")
	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	switch scm {
	default:
		config[GH] = ScmTokenConfig{
			AccessToken: token,
		}
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

// @TODO refactor WriteToken & CheckForAccess functions.
// There is some DRY code in the two functions that I would like to refactor.
// Specifically for getting the current directory, home dir and joining the paths
// for the configuration file.
func CheckForAccess(scm string) (bool, error) {
	config := make(map[string]ScmTokenConfig)
	authorized := false

	usr, err := user.Current()
	if err != nil {
		return authorized, err
	}

	home := usr.HomeDir
	configFile := filepath.Join(home, ".config", "issue-summoner", "config.json")

	file, err := os.OpenFile(configFile, os.O_RDONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return authorized, err
		} else {
			return authorized, errors.New("Error opening file")
		}
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return authorized, errors.New("Error decoding config file")
	}

	return config[scm].AccessToken != "", nil
}

func ReadAccessToken(scm string) (string, error) {
	config := make(map[string]ScmTokenConfig)

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	home := usr.HomeDir
	configFile := filepath.Join(home, ".config", "issue-summoner", "config.json")

	file, err := os.OpenFile(configFile, os.O_RDONLY, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		} else {
			return "", errors.New("Error opening file")
		}
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return "", errors.New("Error decoding config file")
	}

	accessToken := config[scm].AccessToken
	if accessToken == "" {
		return "", errors.New("Access token does not exist")
	}

	return accessToken, nil
}

// ExtractUserRepoName takes the output from <git remote --verbose> command
// as input and attempts to extract the user name and repository name from out
func ExtractUserRepoName(out []byte) (string, string, error) {
	if len(out) == 0 {
		return "", "", errors.New(
			"expected to receive the output from <git remote -v> but got empty byte slice",
		)
	}

	line := bytes.Split(out, []byte("\n"))[0]
	// fields will give us -> ["origin", "url (https | ssh)", "(push) | (pull)"]
	// we only care about the url since it contains both the username and repo name
	fields := bytes.Fields(line)
	if len(fields) < 2 {
		return "", "", fmt.Errorf(
			"expected to receive the origin and url but got %s",
			string(fields[0]),
		)
	}

	url := fields[1]
	if bytes.HasPrefix(url, []byte("https")) {
		userName, repoName := extractFromHTTPS(url)
		return userName, repoName, nil
	}

	if bytes.HasPrefix(url, []byte("git")) {
		userName, repoName := extractFromSSH(url)
		return userName, repoName, nil
	}

	return "", "", fmt.Errorf("expected a https or ssh url but got %s", string(url))
}

func extractFromHTTPS(url []byte) (string, string) {
	split := bytes.SplitAfter(url, []byte("https://"))[1]
	sep := bytes.Split(split, []byte("/"))
	userName, repoName := sep[1], sep[2]
	return string(userName), string(bytes.TrimSuffix(repoName, []byte(".git")))
}

func extractFromSSH(url []byte) (string, string) {
	split := bytes.SplitAfter(url, []byte(":"))[1]
	sep := bytes.Split(split, []byte("/"))
	userName, repoName := sep[0], sep[1]
	return string(userName), string(bytes.TrimSuffix(repoName, []byte(".git")))
}

// GlobalUserName uses the **git config** command to retrieve the global
// configuration options. Specifically, the user.name option. The userName is
// read and set onto the reciever's (GitConfig) UserName property. This will be used
func GlobalUserName() (string, error) {
	var out strings.Builder
	cmd := exec.Command("git", "config", "--global", "user.name")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	userName := strings.TrimSpace(out.String())
	if userName == "" {
		return "", errors.New("global userName option not set. See man git config for more details")
	}

	return userName, nil
}

// @TODO write unit test for RepoName/extractRepoName function.
func RepoName() (string, error) {
	var out strings.Builder
	cmd := exec.Command("git", "remote", "-v")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	repoName := extractRepoName(out.String())
	if repoName == "" {
		return "", errors.New("Failed to get repo name")
	}

	return repoName, nil
}

// extractRepoName takes the output from the `git remote -v` command as input (origins) and outputs the repository name.
// The function can handle both ssh and https origins.
// Git does not offer a command that outputs the repository name directly
func extractRepoName(origins string) string {
	for _, line := range strings.Split(origins, "\n") {
		if strings.Contains(line, "(push)") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				repoURL := fields[1]
				parts := strings.Split(repoURL, "/")
				if len(parts) > 1 {
					repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
					return repo
				}
			}
		}
	}
	return ""
}
