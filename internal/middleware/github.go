package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/artifour/github-webhook/internal/config"
	"github.com/artifour/github-webhook/internal/git"
)

const (
	headerXHubSignature = "X-Hub-Signature"
	headerXHubEvent     = "X-Hub-Event"

	gitEventPing = "ping"
	gitEventPush = "push"
)

func GitHubMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xHubSignature := r.Header.Get(headerXHubSignature)
		xHubEvent := r.Header.Get(headerXHubEvent)
		if r.Method == "post" && xHubSignature != "" && xHubEvent != "" {
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("Cannot read the request body: %s\n", err)
				return
			}

			if !isSignatureValid("", xHubSignature, &body) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			handled, err := handleGitHubEvent(w, xHubEvent, &body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}

			if !handled {
				fmt.Fprintf(w, "Unknow event: %s\tBODY: %s", xHubEvent, body)
				log.Printf("Unknow event: %s\tBODY: %s\n", xHubEvent, body)
			}

			w.WriteHeader(http.StatusOK)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSignatureValid(secret string, signature string, body *[]byte) bool {
	signatureParts := strings.SplitN(signature, "=", 2)
	if signatureParts[0] != "sha1" {
		log.Printf("GitHub Signature unexpected hash algorithm: %s\n", signatureParts[0])
		return false
	}

	secretHash := hmac.New(sha1.New, []byte(secret))
	if _, err := secretHash.Write(*body); err != nil {
		log.Printf("Cannot compute the HMAC for the request: %s\n", err)
		return false
	}

	expectedHash := hex.EncodeToString(secretHash.Sum(nil))
	if signatureParts[1] != expectedHash {
		log.Printf("GitHub Signature validation failed: %s != %s\n", expectedHash, signatureParts[1])
		return false
	}

	return true
}

func handleGitHubEvent(w http.ResponseWriter, event string, body *[]byte) (bool, error) {
	if event == gitEventPing {
		fmt.Fprint(w, "pong")
		return true, nil
	}

	if event == "push" {
		var data map[string]interface{}
		if err := json.Unmarshal(*body, &data); err != nil {
			return true, err
		}

		ref := data["ref"].(string)
		repository := data["repository"].(map[string]interface{})
		repositoryName := repository["full_name"].(string)

		branchName := strings.SplitN(ref, "/", 3)[2]

		pushRepository(repositoryName, branchName)

		fmt.Fprint(w, "ok")

		return true, nil
	}

	return false, nil
}

func pushRepository(repositoryName string, branchName string) {
	if branchName != "master" {
		return
	}

	path := config.Get(fmt.Sprintf("repository.%s", repositoryName))
	if path == "" {
		log.Printf("New commits was pushed to the repository '%s' but it is not configurated\n", repositoryName)
		return
	}

	if err := git.Pull(path); err != nil {
		log.Printf("Command finished with error: %s\n", err)
		return
	}

	log.Printf("Repository '%s' successfully updated\n", repositoryName)
}
