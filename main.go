package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cyphera-labs/cyphera-go"
)

var client *cyphera.Cyphera

func init() {
	var err error
	client, err = cyphera.Load()
	if err != nil {
		log.Printf("Warning: could not load cyphera policy: %v", err)
		log.Printf("Set CYPHERA_POLICY_FILE or place cyphera.json in working directory")
	}
}

// BQ Remote UDF request/response format
type bqRequest struct {
	Calls [][]string `json:"calls"`
}

type bqResponse struct {
	Replies []string `json:"replies"`
}

// POST / — cyphera_protect(policy, value)
func handleProtect(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Error(w, "cyphera not initialized — no policy file found", http.StatusServiceUnavailable)
		return
	}

	var req bqRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := bqResponse{Replies: make([]string, 0, len(req.Calls))}
	for _, call := range req.Calls {
		if len(call) != 2 {
			resp.Replies = append(resp.Replies, "[error: expected (policy, value)]")
			continue
		}
		policyName, value := call[0], call[1]
		protected, err := client.Protect(value, policyName)
		if err != nil {
			resp.Replies = append(resp.Replies, fmt.Sprintf("[error: %s]", err))
		} else {
			resp.Replies = append(resp.Replies, protected)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /access — cyphera_access(value) or cyphera_access(value, policy)
func handleAccess(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Error(w, "cyphera not initialized — no policy file found", http.StatusServiceUnavailable)
		return
	}

	var req bqRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := bqResponse{Replies: make([]string, 0, len(req.Calls))}
	for _, call := range req.Calls {
		var accessed string
		var err error
		switch len(call) {
		case 1: // tag-based access
			accessed, err = client.Access(call[0])
		case 2: // explicit policy
			accessed, err = client.Access(call[0], call[1])
		default:
			resp.Replies = append(resp.Replies, "[error: expected (value) or (value, policy)]")
			continue
		}
		if err != nil {
			resp.Replies = append(resp.Replies, fmt.Sprintf("[error: %s]", err))
		} else {
			resp.Replies = append(resp.Replies, accessed)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if client == nil {
		status = "no policy loaded"
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"%s"}`, status)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleProtect)
	http.HandleFunc("/access", handleAccess)
	http.HandleFunc("/health", handleHealth)

	log.Printf("Cyphera BQ UDF server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
