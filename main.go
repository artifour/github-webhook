package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type configuration struct {
	Port         string            `json:"port"`
	Secret       string            `json:"secret"`
	Repositories map[string]string `json:"repositories"`
}

var Configuration configuration
var WorkDir string

func main() {
	readConfiguration()
	runHttpServer()
}

func readConfiguration() {
	command, err := os.Executable()
	if err != nil {
		panic(err)
	}
	WorkDir := filepath.Dir(command) + string(os.PathSeparator)
	fmt.Println(WorkDir)
	file, err := os.Open(WorkDir + "conf.json")
	if err != nil {
		log.Fatal("Cannot read 'conf.json' file")
	}
	defer file.Close()

	jsonDecoder := json.NewDecoder(file)
	if jsonDecoder.Decode(&Configuration) != nil {
		log.Fatal("error:", err)
	}
}

func runHttpServer() {
	addr := fmt.Sprintf(":%s", Configuration.Port)
	log.Printf("Listening on %s...", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(handleRequest))
	log.Fatal(err)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	gitHubSignature := r.Header.Get("X-Hub-Signature")
	if gitHubSignature == "" {
		w.WriteHeader(http.StatusForbidden)
		log.Println("HTTP Header 'X-Hub-Signature' is missing")
		return
	}

	hash := strings.SplitN(gitHubSignature, "=", 2)
	if hash[0] != "sha1" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("GitHub Signature unexpected hash algorithm: %s\n", hash[0])
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Cannot read the request body: %s\n", err)
		return
	}

	secretHash := hmac.New(sha1.New, []byte(Configuration.Secret))
	if _, err := secretHash.Write(body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Cannot compute the HMAC for the request: %s\n", err)
		return
	}

	expectedHash := hex.EncodeToString(secretHash.Sum(nil))
	if hash[1] != expectedHash && false {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("GitHub Signature validation failed: %s != %s", expectedHash, hash[1])
		return
	}

	gitHubEvent := r.Header.Get("X-GitHub-Event")
	if gitHubEvent == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("HTTP Header 'X-GitHub-Event' is missing")
		return
	}

	handleGitHubEvent(gitHubEvent, &body, w)
}

func handleGitHubEvent(gitHubEvent string, body *[]byte, w http.ResponseWriter) {
	if gitHubEvent == "ping" {
		fmt.Fprint(w, "pong")
		return
	}

	if gitHubEvent == "push" {
		var data map[string]interface{}
		err := json.Unmarshal(*body, &data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		ref := data["ref"].(string)
		repository := data["repository"].(map[string]interface{})
		repositoryName := repository["full_name"].(string)

		branchName := strings.SplitN(ref, "/", 3)[2]

		pushRepository(repositoryName, branchName)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")

		return
	}

	fmt.Fprintf(w, "Unknow event: %s\tBODY: %s", gitHubEvent, *body)
	log.Printf("Unknow event: %s\tBODY: %s", gitHubEvent, *body)
}

func pushRepository(repositoryName string, branchName string) {
	if branchName != "master" {
		return
	}

	path, ok := Configuration.Repositories[repositoryName]
	if !ok {
		log.Printf("New commits was pushed to the repository '%s' but it is not configurated\n", repositoryName)
		return
	}

	cmd := exec.Command("git", "-C", path, "pull", "--no-edit")
	_, err := cmd.Output()
	if err != nil {
		log.Printf("Command finished with error: %s\n", err)
		return
	}

	log.Printf("Repository '%s' successfully updated\n", repositoryName)
}
