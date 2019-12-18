package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v4"
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

func randomContact() *contact {
	id, _ := uuid.NewRandom()
	firstName, lastName := gofakeit.FirstName(), gofakeit.LastName()
	displayName := fmt.Sprintf("%s %s", firstName, lastName)
	email := strings.ToLower(fmt.Sprintf("%s.%s@example.com", firstName, lastName))

	return &contact{
		BusinessPhones:    []string{gofakeit.Phone()},
		DisplayName:       displayName,
		GivenName:         firstName,
		Mail:              email,
		MobilePhone:       gofakeit.Phone(),
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
	baseURL string
)

func main() {
	var (
		baseURL       string
		totalContacts int
	)
	flag.StringVar(&baseURL, "baseurl", "https://graph.microsoft.com", "Base URL used in results")
	flag.IntVar(&totalContacts, "total", 100, "Number of available contacts behind the `/users` endpoint")
	flag.Parse()

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

		fmt.Printf("REQ /v1.0/users?total=%d&start=%d\n", total, start)

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

	http.HandleFunc("/v1.0/me/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{baseURL, "", []*contact{}})
	})

	fmt.Println("listening on :9999")
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}
