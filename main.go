package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ── Dummy cipher (placeholder until cyphera-go is wired in) ──

var alphabets = map[string]string{
	"digits":       "0123456789",
	"alpha_lower":  "abcdefghijklmnopqrstuvwxyz",
	"alphanumeric": "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
}

func deriveShift(keyHex string) int {
	var h uint32
	for _, c := range keyHex {
		h = 31*h + uint32(c)
	}
	return int(h%256) + 1
}

func dummyEncrypt(value, alphabetName, keyHex string) string {
	alpha, ok := alphabets[alphabetName]
	if !ok {
		alpha = alphabets["digits"]
	}
	shift := deriveShift(keyHex)
	var sb strings.Builder
	for _, c := range value {
		idx := strings.IndexRune(alpha, c)
		if idx >= 0 {
			sb.WriteByte(alpha[(idx+shift)%len(alpha)])
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func dummyDecrypt(value, alphabetName, keyHex string) string {
	alpha, ok := alphabets[alphabetName]
	if !ok {
		alpha = alphabets["digits"]
	}
	shift := deriveShift(keyHex)
	var sb strings.Builder
	for _, c := range value {
		idx := strings.IndexRune(alpha, c)
		if idx >= 0 {
			sb.WriteByte(alpha[((idx-(shift%len(alpha)))+len(alpha))%len(alpha)])
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// ── Policy ──

type PolicyDef struct {
	Engine   string `yaml:"engine"`
	Alphabet string `yaml:"alphabet"`
	KeyRef   string `yaml:"key_ref"`
}

type KeyDef struct {
	Material string `yaml:"material"`
}

type Config struct {
	Policies map[string]PolicyDef `yaml:"policies"`
	Keys     map[string]KeyDef    `yaml:"keys"`
}

var cfg Config

func loadPolicies() {
	path := os.Getenv("CYPHERA_POLICY_FILE")
	if path == "" {
		path = "/etc/cyphera/cyphera.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: could not load policy file %s: %v", path, err)
		return
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("Warning: could not parse policy file %s: %v", path, err)
		return
	}
	log.Printf("Loaded %d policies from %s", len(cfg.Policies), path)
}

func resolvePolicy(name string) (string, string) {
	p, ok := cfg.Policies[name]
	if !ok {
		return "", ""
	}
	material := ""
	if k, ok := cfg.Keys[p.KeyRef]; ok {
		material = k.Material
	}
	return p.Alphabet, material
}

// ── BQ Remote UDF HTTP handler ──

type bqRequest struct {
	Calls [][]string `json:"calls"`
}

type bqResponse struct {
	Replies []string `json:"replies"`
}

func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	var req bqRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := bqResponse{Replies: make([]string, 0, len(req.Calls))}
	for _, call := range req.Calls {
		switch len(call) {
		case 2: // (policy_name, value)
			alpha, key := resolvePolicy(call[0])
			if alpha == "" {
				resp.Replies = append(resp.Replies, fmt.Sprintf("[unknown policy: %s]", call[0]))
			} else {
				resp.Replies = append(resp.Replies, dummyEncrypt(call[1], alpha, key))
			}
		case 3: // (value, key_hex, alphabet)
			resp.Replies = append(resp.Replies, dummyEncrypt(call[0], call[2], call[1]))
		default:
			resp.Replies = append(resp.Replies, "[invalid args]")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	var req bqRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := bqResponse{Replies: make([]string, 0, len(req.Calls))}
	for _, call := range req.Calls {
		switch len(call) {
		case 2:
			alpha, key := resolvePolicy(call[0])
			if alpha == "" {
				resp.Replies = append(resp.Replies, fmt.Sprintf("[unknown policy: %s]", call[0]))
			} else {
				resp.Replies = append(resp.Replies, dummyDecrypt(call[1], alpha, key))
			}
		case 3:
			resp.Replies = append(resp.Replies, dummyDecrypt(call[0], call[2], call[1]))
		default:
			resp.Replies = append(resp.Replies, "[invalid args]")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","policies":%d}`, len(cfg.Policies))
}

func main() {
	loadPolicies()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleEncrypt)
	http.HandleFunc("/decrypt", handleDecrypt)
	http.HandleFunc("/health", handleHealth)

	log.Printf("Cyphera BQ UDF server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
