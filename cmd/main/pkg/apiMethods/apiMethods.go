package apiMethods

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

type Mail struct {
	Date    string `json:"Date"`
	From    string `json:"From"`
	Subject string `json:"Subject"`
	To      string `json:"To"` // the author
	Body    string `json:"Body"`
}

type ResponseMails struct {
	Took     int64  `json:"took"`
	Time_out bool   `json:"time_out"`
	ErrorM   string `json:"error"`
	Hits     struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits_2 []struct {
			Index    string `json:"_index"`
			Dtype    string `json:"_type"`
			IdM      string `json:"_id"`
			Score    int64  `json:"_score"`
			Timestap string `json:"@timestamp"`
			Source   Mail   `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// "AllMails" handles GET requests by extracting all the mail in database,
//  calling the "get_all_mails" function, and sending the response as a JSON encoded object.

func AllMails(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-Type", "Application/json")
	w.Header().Set("done-by", "Luis Martinez")

	from := r.URL.Query().Get("from")
	max_items := r.URL.Query().Get("max")

	resp := get_all_mails(from, max_items)

	json.NewEncoder(w).Encode(resp)

}

// "FilterMails" handles GET requests by extracting query parameters from the request,
// calling the "get_filter_mails" function, and sending the response as a JSON encoded object.

func FilterMails(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-Type", "Application/json")
	w.Header().Set("done-by", "Luis Martinez")

	from := r.URL.Query().Get("from")
	max_items := r.URL.Query().Get("max")
	filter := r.URL.Query().Get("filterID")

	resp, err := get_filter_mails(from, max_items, filter)

	if err != nil {
		w.WriteHeader(404)
	}

	json.NewEncoder(w).Encode(resp)
}

// "get_filter_mails" is a helper function that constructs an HTTP request to search
// for emails matching certain conditions and returns the search results.

func get_filter_mails(from string, maxCount string, filter string) (ResponseMails, error) {

	query := `{
        "search_type": "match",
		"query":
        {
            "term":"` + filter + `"
        },
        "from":` + from + `,
        "max_results":` + maxCount + ` ,
        "_source": ["From","To","Date","Subject","body"]
    }`

	req, err := http.NewRequest("POST", "http://localhost:4080/api/maildir/_search", strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Println(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data ResponseMails
	json.Unmarshal(body, &data)

	if data.Hits.Total.Value == 0 {
		return data, errors.New("mail not found")
	}

	return data, nil

}

// "get_all_mails" is a helper function that constructs an HTTP request to search
// for emails matching certain conditions and returns the search results.

func get_all_mails(from string, max_item string) ResponseMails {

	query := `{
        "search_type": "matchall",
        "from":` + from + `,
		"max_results":` + max_item + `,
		"_source": ["From","To","Date","Subject","body"]     }`

	req, err := http.NewRequest("POST", "http://localhost:4080/api/maildir/_search", strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data ResponseMails
	json.Unmarshal(body, &data)
	return data
}
