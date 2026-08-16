package main

import "log"

func main() {
	log.Println("Started collecting matches...")
	CollectMatches()
	log.Println("Done collecting matches!!!")

	log.Println("Started collecting match details...")
	CollectData()
	log.Println("Done collecting match details!!!")

	log.Println("Started collecting mmr details...")
	Mmr()
	log.Println("Done collecting mmr details!!!")

	log.Println("Started tranformaing data...")
	Transform()
	log.Println("Done tranformaing data!!!")
}
