package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type contact struct {
	BusinessPhones    []string  `json:"businessPhones"`
	DisplayName       string    `json:"displayName"`
	GivenName         string    `json:"givenName"`
	JobTitle          *string   `json:"jobTitle"`
	Mail              string    `json:"mail"`
	MobilePhone       string    `json:"mobilePhone"`
	OfficeLocation    *string   `json:"officeLocation"`
	PreferredLanguage *string   `json:"preferredLanguage"`
	Surname           string    `json:"surname"`
	UserPrincipalName string    `json:"userPrincipalName"`
	ID                uuid.UUID `json:"id"`
}

type response struct {
	Context  string     `json:"@odata.context"`
	NextLink string     `json:"@odata.nextLink,omitempty"`
	Value    []*contact `json:"value"`
}

var chars = []rune("abcdefghijklmnopqrstvwxyxzABCDEFGHIJKLMNOPQRSTVWXYZ ")

func randomString(length int) string {
	b := strings.Builder{}
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func randomPhone() string {
	return fmt.Sprintf("%03d-555-%04d", rand.Intn(1000), rand.Intn(10000))
}

func randomContact() *contact {
	id, _ := uuid.NewRandom()
	firstName, lastName := randomString(10), randomString(15)
	displayName := fmt.Sprintf("%s %s", firstName, lastName)
	email := fmt.Sprintf("%s.%s@example.com", firstName, lastName)

	return &contact{
		BusinessPhones:    []string{randomPhone()},
		DisplayName:       displayName,
		GivenName:         firstName,
		Mail:              email,
		MobilePhone:       randomPhone(),
		Surname:           lastName,
		UserPrincipalName: email,
		ID:                id,
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

var (
	all     []*contact
	baseURL = "https://example.com"
)

func main() {
	totalContacts, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	fmt.Print("generating ", totalContacts, " contacts... ")
	all = make([]*contact, totalContacts)
	for i := 0; i < totalContacts; i++ {
		all[i] = randomContact()
	}
	fmt.Println("done.")

	http.HandleFunc("/v1.0/users", func(w http.ResponseWriter, r *http.Request) {
		total, start := 10, 0
		if keys, ok := r.URL.Query()["$top"]; ok {
			total, _ = strconv.Atoi(keys[0])
		}
		if keys, ok := r.URL.Query()["$skiptoken"]; ok {
			start, _ = strconv.Atoi(keys[0])
		}

		end := min(start+total, len(all))
		anyLeft := len(all) != end

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if start < len(all) {
			skipToken := ""
			if anyLeft {
				skipToken = fmt.Sprintf("%s/v1.0/users?$top=%d&$skiptoken=%d", baseURL, total, end)
			}

			_ = json.NewEncoder(w).Encode(response{baseURL, skipToken, all[start:end]})
		}

	})

	fmt.Println("listening on :9999")
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}
