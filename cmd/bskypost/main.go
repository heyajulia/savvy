package main

import "github.com/heyajulia/energieprijzen/internal/bsky"

func main() {
	client, err := bsky.Login("energieprijzen.bsky.social", "3e2u-ayvn-uijx-siy7")
	if err != nil {
		panic(err)
	}

	if err := client.Post("Hallo, wereld!", "https://t.me/energieprijzen"); err != nil {
		panic(err)
	}
}
