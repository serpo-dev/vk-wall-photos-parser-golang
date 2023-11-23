package photos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"vk-album-downloader-golang/libs/utils"
)

type Size struct {
	URL   string `json:"url"`
	Type  string `json:"type"`
	Width int    `json:"width"`
}

type Response struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			Sizes []Size `json:"sizes"`
		} `json:"items"`
	} `json:"response"`
}

func GetAlbum(albumID string, ownerID int, accessToken string) ([]string, int, error) {
	count := 40
	url := fmt.Sprintf("https://api.vk.com/method/photos.get?album_id=%s&owner_id=%d&access_token=%s&v=5.131&count=%d", albumID, ownerID, accessToken, count)

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body := Response{}
	json.NewDecoder(resp.Body).Decode(&body)

	photoURLs := make([]string, 0)
	for _, item := range body.Response.Items {
		if len(item.Sizes) > 0 {
			photoURLs = append(photoURLs, item.Sizes[len(item.Sizes)-1].URL)
		}
	}

	total_count := body.Response.Count
	if total_count <= count {
		return photoURLs, total_count, nil
	}

	found_count := 0

	for offset := count; offset <= total_count; offset += count {
		next_url := fmt.Sprintf("%s&%s", url, fmt.Sprintf("offset=%d", offset))

		resp, err := http.Get(next_url)
		if err != nil {
			return nil, 0, err
		}
		defer resp.Body.Close()

		body := Response{}
		json.NewDecoder(resp.Body).Decode(&body)

		if resp.StatusCode != 200 {
			fmt.Println("Status code is not 200", offset)
		}
		found_count += len(body.Response.Items)
		for _, item := range body.Response.Items {
			if len(item.Sizes) > 0 {
				maxSize, err := findMaxSize(item.Sizes)
				if err != nil {
					fmt.Println("Cannot find max size for one of the images: ", err)
					continue
				}
				photoURLs = append(photoURLs, maxSize.URL)
			}
		}

		time.Sleep(500 * time.Millisecond)

		percent := float64(offset) / float64(total_count) * 100
		if percent > 100 {
			percent = 100
		}
		fmt.Println(fmt.Sprintf("Parsing URLs of images: %s %%", fmt.Sprintf("%.2f", percent)))
	}

	return photoURLs, total_count, nil
}

func DownloadPhoto(photoURL string, folder_path string) error {
	resp, err := http.Get(photoURL)
	if err != nil {
		fmt.Println("Can't download the photo: ", photoURL, err)
	}
	defer resp.Body.Close()

	filename := _getJPGFileName(photoURL)
	file, err := os.Create(fmt.Sprint(folder_path, "/", filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func _getJPGFileName(url string) string {
	re := regexp.MustCompile(`\/([^\/?#]+\.jpg)`)
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	generated := utils.GenerateRandomString(20)
	return generated
}

func findMaxSize(sizes []Size) (*Size, error) {
	priority := []string{"w", "z", "y", "x"}

	for _, p := range priority {
		for i := range sizes {
			if sizes[i].Type == p {
				return &sizes[i], nil
			}
		}
	}

	maxWidth := 0
	var maxItem *Size
	for i := range sizes {
		if sizes[i].Width > maxWidth {
			maxWidth = sizes[i].Width
			maxItem = &sizes[i]
		}
	}

	return maxItem, nil
}
