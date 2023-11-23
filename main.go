package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"vk-album-downloader-golang/libs/check"
	"vk-album-downloader-golang/libs/photos"
	"vk-album-downloader-golang/libs/utils"

	"github.com/joho/godotenv"
)

func main() {
	start_time := time.Now()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading the file .env: ", err)
	}

	albumID := "wall"
	accessToken := os.Getenv("ACCESS_TOKEN")

	candidates, err := check.ReadTXT(accessToken)
	if err != nil {
		fmt.Println("Error while checking the input links: ", err)
		return
	}
	fmt.Println("Success verifying the links! The total amount: ", len(candidates))
	fmt.Println("Creating folders in the '/output' directory...")
	root_output_path := "output/"
	for _, cnd := range candidates {
		err := utils.CreateFolder(fmt.Sprint(root_output_path, cnd.SCREEN_NAME))
		if err != nil {
			fmt.Println("Error while creating the folder: ", cnd.SCREEN_NAME, err)
			return
		}
	}
	fmt.Println(len(candidates), "folders created successfully!")
	fmt.Println("Start parsing source urls of images and downloading them.")

	error_list_writing_count := 0
	for i, cnd := range candidates {
		fmt.Println("Parse URLs of images:", i+1, "/", len(candidates))

		ownerID := cnd.ID
		folder_path := fmt.Sprint(root_output_path, cnd.SCREEN_NAME)
		filename := fmt.Sprintf("URL list of %s.txt", cnd.SCREEN_NAME)
		file_path := fmt.Sprintf("%s/%s", folder_path, filename)

		urls, count, err := photos.GetAlbum(albumID, ownerID, accessToken)
		if err != nil {
			fmt.Println("Error while fetching photos list: ", err)
		}
		candidates[i].COUNT = count

		file, err := os.Create(file_path)
		if err != nil {
			fmt.Println("Error while creating the URL list: ", filename, err)
		}
		defer file.Close()

		for _, url := range urls {
			_, err := fmt.Fprintln(file, url)
			if err != nil {
				fmt.Println("Error while writing a line: ", url, err)
				error_list_writing_count++
			}
		}
	}

	downloaded_count := 0
	for i, cnd := range candidates {
		folder_path := fmt.Sprint(root_output_path, cnd.SCREEN_NAME)
		filename := fmt.Sprintf("URL list of %s.txt", cnd.SCREEN_NAME)
		file_path := fmt.Sprintf("%s/%s", folder_path, filename)

		file, err := os.Open(file_path)
		if err != nil {
			fmt.Println("Cannot read URL list for: ", cnd.SCREEN_NAME, err)
		}
		scanner := bufio.NewScanner(file)

		fmt.Println("Start downloading photos from the URL list: ", cnd.SCREEN_NAME, "(", i+1, "/", len(candidates), ").")

		scanner_count := 0
		for scanner.Scan() {
			scanner_count++

			link := scanner.Text()
			err := photos.DownloadPhoto(link, folder_path)

			if err != nil {
				fmt.Println("Cannot save the photo", link, err)
			} else {
				downloaded_count++
			}

			percent := float64(scanner_count) / float64(cnd.COUNT) * 100
			if percent > 100 {
				percent = 100
			}
			fmt.Println(fmt.Sprintf("Download images: %d / %d (%s %%)", scanner_count, cnd.COUNT, fmt.Sprintf("%.2f", percent)))

			time.Sleep(100 * time.Millisecond)
		}
	}

	elapsed := time.Since(start_time)
	formatted_time := utils.FormatExecutionTime(elapsed)
	fmt.Println("Success! All wall albums are downloaded. The execution time:", formatted_time)
}
