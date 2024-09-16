package main

func main() {
	ReadConfig("config.yaml")
	init_mastodon()
	init_telegram()
}
