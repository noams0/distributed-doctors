package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	siteID := flag.Int("n", 0, "ID du site")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	fmt.Fprintf(os.Stderr, "[APP %d] Démarré\n", *siteID)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(os.Stderr, "[APP %d] Message reçu du contrôleur: %s\n", *siteID, line)

		// Pour simuler une action de demande, on peut répondre au contrôleur
		if line == "ask" {
			fmt.Fprintf(writer, "REQ from %d\n", *siteID)
			writer.Flush()
		}
	}
}
