package check

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

type Candidate struct {
	ID          int
	SCREEN_NAME string
	COUNT       int
}

func ReadTXT(accessToken string) ([]Candidate, error) {
	filePath := "input/input.txt"

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error while reading the file 'input.txt'. Make sure that it exists.")
		return nil, err
	}
	defer file.Close()

	fmt.Println("Parse and check the links from the TXT file. Wait please...")

	scanner := bufio.NewScanner(file)
	candidates := make([]Candidate, 0)
	for scanner.Scan() {
		link := scanner.Text()

		re := regexp.MustCompile(`vk\.com/([^/]+)`)
		match := re.FindStringSubmatch(link)

		screen_name := ""
		if len(match) > 1 {
			screen_name = match[1]
		} else {
			fmt.Println("Invalid link: ", link)
			return nil, err
		}

		id, err := _getIdByScreenName(screen_name, accessToken)
		if err != nil {
			return nil, err
		}

		allowedChars := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
		screen_name = allowedChars.ReplaceAllString(screen_name, "")

		candidate := Candidate{
			ID:          id,
			SCREEN_NAME: screen_name,
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

type GetIdResponse struct {
	Response struct {
		Object_ID int    `json:"object_id"`
		Type      string `json:"type"`
	} `json:"response"`
}

func _getIdByScreenName(name string, accessToken string) (int, error) {
	url := fmt.Sprintf("https://api.vk.com/method/utils.resolveScreenName?screen_name=%s&access_token=%s&v=5.131", name, accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body := GetIdResponse{}
	json.NewDecoder(resp.Body).Decode(&body)

	object_id := body.Response.Object_ID
	object_type := body.Response.Type
	if object_type == "user" {
		return object_id, nil
	} else if object_type == "group" {
		return (-1) * object_id, nil
	} else {
		fmt.Print(object_id)
		panic("The object type is neither 'user' nor 'group'.")
	}
}
