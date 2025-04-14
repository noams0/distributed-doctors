package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	siteID := flag.Int("n", 0, "ID du site")
	flag.Parse()
	message := "Message périodique"
	var wg sync.WaitGroup
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Canal pour gérer la concurrence
	syncChan := make(chan struct{}, 1) // Limité à 1 action
	syncChan <- struct{}{}             // Initialiser le canal

	wg.Add(2)

	// Booléen pour vérifier si le message a déjà été traité

	// Goroutine pour la réception
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() { // Scanner.Scan permet la lecture bloquante
			<-syncChan // Attente pour la concurrence

			// Lire le message du terminal
			message = scanner.Text()
			fmt.Fprintf(os.Stderr, "[CTL %d] Message reçu: %s\n", *siteID, message)

			// Traiter le message
			if message == "ask" {
				// Effectuer l'action ici, par exemple envoyer une réponse
				fmt.Println("REQ from site")
			}

			syncChan <- struct{}{} // Libérer la concurrence
		}
	}()

	// Goroutine pour l'émission périodique
	go func() {
		defer wg.Done()
		<-syncChan // Attente pour la concurrence

		fmt.Println(message)

		syncChan <- struct{}{} // Libérer la concurrence
	}()

	wg.Wait()
}
